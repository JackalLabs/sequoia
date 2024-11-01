package config

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

func (c Config) Validate() error {
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

	if c.LogSSHConfig.Enable {
		for i := range c.LogSSHConfig.AuthorizedPubKeys {
			_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(c.LogSSHConfig.AuthorizedPubKeys[i]))
			if err != nil {
				return errors.Join(errors.New("invalid ssh authorized key"), err)
			}
		}
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
