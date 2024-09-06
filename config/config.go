package config

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

func (c Config) Validate() error {
	if c.DataDirectory == "" {
		return errors.New("invalid data directory")
	}

	switch c.DataStoreConfig.Backend {
	case OptFlatFS:
	case OptBadgerDS:
		if c.DataStoreConfig.Directory != c.DataDirectory {
			return errors.New("badger ds directory must be the same as data directory")
		}
	default:
		return errors.New("invalid data store backend")
	}

	return nil
}

// ReadConfig parses data and returns Config.
// Error during parsing or an invalid configuration in the Config will return an error.
func ReadConfig(data []byte) (*Config, error) {
	//not using a default config to detect badger ds users
	config := Config{}

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.DataStoreConfig.Backend == "" && config.DataStoreConfig.Directory == "" {
		config.DataStoreConfig.Backend = OptBadgerDS
		config.DataStoreConfig.Directory = config.DataDirectory
	}

	return &config, config.Validate()
}

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
