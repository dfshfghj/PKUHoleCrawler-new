package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 配置结构
type Config struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	SecretKey string `json:"secret_key"`
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	currentDir, _ := os.Getwd()
	configPath := filepath.Join(currentDir, "config.json")

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}

	if config.Username == "" || config.Password == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("请填写配置文件")
	}

	return &config, nil
}
