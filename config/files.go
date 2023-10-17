package config

import (
	"errors"
	"os"
	"path"
)

const DefaultHome = "$HOME/.sequoia"

const ConfigFileName = "config.yaml"

func createIfNotExists(directory string, fileName string, contents []byte) (bool, error) {
	filePath := path.Join(directory, fileName)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(filePath)
		if err != nil {
			return false, err
		}
		defer f.Close()

		_, err = f.Write(contents)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func readFile(directory string, filename string) ([]byte, error) {
	filePath := path.Join(directory, filename)

	dat, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return dat, nil
}

func createFiles(directory string) error {
	config, err := DefaultConfig().Export()
	if err != nil {
		return err
	}
	_, err = createIfNotExists(directory, ConfigFileName, config)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfigFile(directory string) (*Config, error) {
	configData, err := readFile(directory, ConfigFileName)
	if err != nil {
		return nil, err
	}

	config, err := ReadConfig(configData)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func Init() (*Config, error) {
	directory := os.ExpandEnv(DefaultHome)

	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = createFiles(directory)
	if err != nil {
		return nil, err
	}

	return ReadConfigFile(directory)
}
