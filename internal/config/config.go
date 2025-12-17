package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	filename := ".gatorconfig.json"
	home_dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home_dir + "/" + filename, nil
}

func Read() (Config, error) {
	filename, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	json.Unmarshal(contents, &config)
	return config, nil
}

func (c *Config) SetUser(username string) error {
	config, err := Read()
	if err != nil {
		return err
	}
	config.Current_user_name = username
	filename, err := getConfigFilePath()
	if err != nil {
		return err
	}
	file_contents, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, file_contents, 0666)
	if err != nil {
		return err
	}
	return nil
}
