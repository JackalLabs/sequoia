package config

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/JackalLabs/sequoia/cmd/types"
	"github.com/JackalLabs/sequoia/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ConfigCmd returns the parent command for config operations
func ConfigCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Config subcommands",
	}

	c.AddCommand(getCmd(), setCmd(), showCmd())

	return c
}

// showCmd returns the command to show the entire config
func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the entire configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfg, err := config.Init(home)
			if err != nil {
				return err
			}

			data, err := cfg.Export()
			if err != nil {
				return err
			}

			fmt.Print(string(data))
			return nil
		},
	}
}

// getCmd returns the command to get a config value
func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get a config value",
		Long: `Get a config value by key. Use dot notation for nested values.

Examples:
  sequoia config get queue_interval
  sequoia config get api_config.port
  sequoia config get chain_config.rpc_addr`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfg, err := config.Init(home)
			if err != nil {
				return err
			}

			value, err := getConfigValue(cfg, key)
			if err != nil {
				return err
			}

			fmt.Println(value)
			return nil
		},
	}
}

// setCmd returns the command to set a config value
func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a config value",
		Long: `Set a config value by key. Use dot notation for nested values.

Examples:
  sequoia config set queue_interval 50
  sequoia config set api_config.port 8080
  sequoia config set chain_config.rpc_addr http://localhost:26657`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfg, err := config.Init(home)
			if err != nil {
				return err
			}

			if err := setConfigValue(cfg, key, value); err != nil {
				return err
			}

			// Write the updated config back to file
			directory := os.ExpandEnv(home)
			if err := writeConfigFile(directory, cfg); err != nil {
				return err
			}

			fmt.Printf("%s set to %s\n", key, value)
			return nil
		},
	}
}

// writeConfigFile writes the config to the config file
func writeConfigFile(directory string, cfg *config.Config) error {
	data, err := cfg.Export()
	if err != nil {
		return err
	}

	filePath := path.Join(directory, config.ConfigFileName)
	return os.WriteFile(filePath, data, 0600)
}

// getConfigValue gets a config value by key using reflection
func getConfigValue(cfg *config.Config, key string) (string, error) {
	parts := strings.Split(key, ".")
	v := reflect.ValueOf(cfg).Elem()

	for _, part := range parts {
		field, err := findFieldByTag(v, part)
		if err != nil {
			return "", err
		}
		v = field
	}

	// For struct types, serialize as YAML for better readability
	if v.Kind() == reflect.Struct {
		data, err := yaml.Marshal(v.Interface())
		if err != nil {
			return "", fmt.Errorf("failed to serialize struct: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	return fmt.Sprintf("%v", v.Interface()), nil
}

// setConfigValue sets a config value by key using reflection
func setConfigValue(cfg *config.Config, key string, value string) error {
	parts := strings.Split(key, ".")
	v := reflect.ValueOf(cfg).Elem()

	// Navigate to the parent of the field we want to set
	for i := 0; i < len(parts)-1; i++ {
		field, err := findFieldByTag(v, parts[i])
		if err != nil {
			return err
		}
		v = field
	}

	// Find and set the final field
	field, err := findFieldByTag(v, parts[len(parts)-1])
	if err != nil {
		return err
	}

	if !field.CanSet() {
		return fmt.Errorf("cannot set field %s", key)
	}

	// Convert and set the value based on the field type
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: %s", value)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s (use true/false)", value)
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

// findFieldByTag finds a struct field by its yaml/mapstructure tag
func findFieldByTag(v reflect.Value, tag string) (reflect.Value, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check yaml tag
		yamlTag := field.Tag.Get("yaml")
		if yamlTag != "" {
			yamlTag = strings.Split(yamlTag, ",")[0]
			if yamlTag == tag {
				return v.Field(i), nil
			}
		}

		// Check mapstructure tag
		msTag := field.Tag.Get("mapstructure")
		if msTag != "" {
			msTag = strings.Split(msTag, ",")[0]
			if msTag == tag {
				return v.Field(i), nil
			}
		}

		// Also check field name (case-insensitive)
		if strings.EqualFold(field.Name, tag) {
			return v.Field(i), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("unknown config key: %s", tag)
}
