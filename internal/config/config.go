package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

// Read a JSON config file and return Config struct
func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(fmt.Sprintf("%s", path))
	if err != nil {
		return Config{}, fmt.Errorf("Failed to read path %s", path)
	}
	config := Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, errors.New("Unmarshal failed")
	}
	return config, nil
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUsername = userName
	err := write(c)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Variable for home is not set properly")
	}
	path := fmt.Sprintf("%s/%s", homeDir, configFileName)
	return path, nil
}

// Write to file
func write(c *Config) error {
	jsonData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}
