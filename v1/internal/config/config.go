package config

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	Conf *Config
)

type DatabaseConfig struct {
	Type     string `json:"type"`     // "sqlite3" or "postgres"
	Host     string `json:"host"`     // PostgreSQL host
	Port     int    `json:"port"`     // PostgreSQL port
	User     string `json:"user"`     // PostgreSQL user
	Password string `json:"password"` // PostgreSQL password
	Name     string `json:"name"`     // PostgreSQL database name
	DBFile   string `json:"db_file"`  // SQLite file path
	SSLMode  string `json:"ssl_mode"` // PostgreSQL SSL mode
	DSN      string `json:"dsn"`      // Custom DSN (optional)
}

type CorsConfig struct {
	AllowOrigins []string `json:"allow_origins"`
	AllowMethods []string `json:"allow_methods"`
	AllowHeaders []string `json:"allow_headers"`
}

type Config struct {
	Username   string         `json:"username"`
	Password   string         `json:"password"`
	SecretKey  string         `json:"secret_key"`
	DeviceUUID string         `json:"device_uuid"` // 设备标识符，用于API请求的uuid header
	Database   DatabaseConfig `json:"database"`
	Cors       CorsConfig     `json:"cors"`
}

func LoadConfig() (*Config, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取工作目录失败: %w", err)
	}
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

	// 设置默认数据库配置
	if config.Database.Type == "" {
		config.Database.Type = "sqlite3"
	}
	if config.Database.Type == "sqlite3" && config.Database.DBFile == "" {
		config.Database.DBFile = "./treehole.db"
	}
	if config.Database.Type == "postgres" {
		if config.Database.Host == "" {
			config.Database.Host = "localhost"
		}
		if config.Database.Port == 0 {
			config.Database.Port = 5432
		}
		if config.Database.SSLMode == "" {
			config.Database.SSLMode = "disable"
		}
	}

	// 生成并保存 device_uuid（如果为空）
	if config.DeviceUUID == "" {
		config.DeviceUUID = generateDeviceUUID()
		if err := saveDeviceUUID(configPath, config.DeviceUUID); err != nil {
			log.Printf("[Config] 保存 device_uuid 失败: %v", err)
		} else {
			log.Printf("[Config] 已自动生成 device_uuid: %s", config.DeviceUUID)
		}
	}

	Conf = &config

	return &config, nil
}

func generateDeviceUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("Web_PKUHOLE_2.0.0_WEB_UUID_%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func saveDeviceUUID(configPath, uuid string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var config map[string]interface{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return err
	}

	config["device_uuid"] = uuid

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) GetDatabaseDSN() (string, error) {
	if c.Database.DSN != "" {
		return c.Database.DSN, nil
	}

	switch c.Database.Type {
	case "sqlite3":
		if c.Database.DBFile == "" {
			return "", fmt.Errorf("sqlite3 database file path is required")
		}
		return c.Database.DBFile, nil
	case "postgres":
		if c.Database.User == "" || c.Database.Password == "" || c.Database.Name == "" {
			return "", fmt.Errorf("postgres database requires user, password, and name")
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name, c.Database.SSLMode), nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}
}
