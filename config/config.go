package config

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

func (c Config) validate() error {
	if c.DataDirectory == "" {
		return errors.New("invalid data directory")
	}

	switch c.BlockStoreConfig.Type {
	case OptFlatFS:
	case OptBadgerDS:
		if c.BlockStoreConfig.Directory != c.DataDirectory {
			return errors.New("badger ds directory must be the same as data directory")
		}
	default:
		return errors.New("invalid data store backend")
	}

	if c.SSHConfig.Enable {
		for i := range c.SSHConfig.AuthorizedPubKeys {
			_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(c.SSHConfig.AuthorizedPubKeys[i]))
			if err != nil {
				return errors.Join(errors.New("invalid ssh authorized key"), err)
			}
		}

		if _, err := os.Stat(c.SSHConfig.HostKeyFile); err != nil && c.SSHConfig.HostKeyFile != "" {
			return errors.Join(errors.New("invalid host key file path"), err)
		}
	}

	return nil
}

// readConfig parses data and returns Config.
// Error during parsing or an invalid configuration in the Config will return an error.
func readConfig(data []byte) (*Config, error) {
	// not using a default config to detect badger ds users
	config := Config{}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.BlockStoreConfig.Type == "" && config.BlockStoreConfig.Directory == "" {
		config.BlockStoreConfig.Type = OptBadgerDS
		config.BlockStoreConfig.Directory = config.DataDirectory
	}

	config.DataDirectory = expandPath(config.DataDirectory)
	config.LogFile = expandPath(config.DataDirectory)
	config.SSHConfig.HostKeyFile = expandPath(config.SSHConfig.HostKeyFile)

	return &config, config.validate()
}

// Export converts the config to yaml format
func (c Config) Export() ([]byte, error) {
	sb := strings.Builder{}
	sb.WriteString("######################\n")
	sb.WriteString("### Sequoia Config ###\n")
	sb.WriteString("######################\n\n")

	d, err := yaml.Marshal(&c)
	if err != nil {
		return nil, err
	}

	sb.Write(d)

	sb.WriteString("\n######################\n")

	return []byte(sb.String()), nil
}
