package config

import (
	"errors"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

const (
	// MaxSensibleQueueSizeBytes is the maximum reasonable queue size to catch configuration typos
	// Set to 1GB (1024 * 1024 * 1024 bytes) as an upper bound
	MaxSensibleQueueSizeBytes = 1024 * 1024 * 1024
)

func (c Config) Validate() error {
	if c.DataDirectory == "" {
		return errors.New("invalid data directory")
	}

	// Validate MaxSizeBytes bounds
	if c.MaxSizeBytes <= 0 {
		return errors.New("MaxSizeBytes must be greater than 0")
	}
	if c.MaxSizeBytes > MaxSensibleQueueSizeBytes {
		return errors.New("MaxSizeBytes exceeds maximum sensible limit (1GB), check for configuration typos")
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

	return nil
}

// ReadConfig parses data and returns Config.
// Error during parsing or an invalid configuration in the Config will return an error.
func ReadConfig(data []byte) (*Config, error) {
	// not using a default config to detect badger ds users
	config := Config{}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.BlockStoreConfig.Type == "" && config.BlockStoreConfig.Directory == "" {
		config.BlockStoreConfig.Type = OptBadgerDS
		config.BlockStoreConfig.Directory = config.DataDirectory
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
