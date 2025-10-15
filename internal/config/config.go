package config

import (
	"os"
	"path/filepath"
	"encoding/json"
)


// consts
const configFileName = ".gatorconfig.json"


// structs
type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name`
}


// functions
func Read() (Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err = json.Unmarshal(fileContent, &cfg); err != nil {
		return Config{}, err
	}
	
	return cfg, nil
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	return write(*c)
}


// helpers
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configFileName), nil
}

func write(cfg Config) error {
	jsonConfig, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, jsonConfig, 0666)
}
