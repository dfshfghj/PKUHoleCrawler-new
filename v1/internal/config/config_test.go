package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, cfg map[string]interface{}) string {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	err = os.WriteFile(filepath.Join(dir, "config.json"), data, 0644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
	return dir
}

func TestLoadConfigValid(t *testing.T) {
	dir := writeTempConfig(t, map[string]interface{}{
		"username":   "test",
		"password":   "pass",
		"secret_key": "key",
	})
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.Username != "test" {
		t.Errorf("Username = %s, want test", cfg.Username)
	}
	if cfg.Database.Type != "sqlite3" {
		t.Errorf("Database.Type = %s, want sqlite3", cfg.Database.Type)
	}
	if cfg.Database.DBFile != "./treehole.db" {
		t.Errorf("Database.DBFile = %s, want ./treehole.db", cfg.Database.DBFile)
	}
}

func TestLoadConfigMissingFields(t *testing.T) {
	dir := writeTempConfig(t, map[string]interface{}{
		"username":   "",
		"password":   "pass",
		"secret_key": "key",
	})
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() expected error for empty username")
	}
}

func TestLoadConfigNoFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("LoadConfig() expected error when config.json missing")
	}
}

func TestLoadConfigPostgresDefaults(t *testing.T) {
	dir := writeTempConfig(t, map[string]interface{}{
		"username":   "test",
		"password":   "pass",
		"secret_key": "key",
		"database": map[string]interface{}{
			"type":     "postgres",
			"user":     "myuser",
			"password": "mypass",
			"name":     "mydb",
		},
	})
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("Host = %s, want localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Port = %d, want 5432", cfg.Database.Port)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("SSLMode = %s, want disable", cfg.Database.SSLMode)
	}
}

func TestGetDatabaseDSNSQLite(t *testing.T) {
	cfg := &Config{
		Username:  "test",
		Password:  "pass",
		SecretKey: "key",
		Database: DatabaseConfig{
			Type:   "sqlite3",
			DBFile: "./test.db",
		},
	}
	dsn, err := cfg.GetDatabaseDSN()
	if err != nil {
		t.Fatalf("GetDatabaseDSN() error: %v", err)
	}
	if dsn != "./test.db" {
		t.Errorf("DSN = %s, want ./test.db", dsn)
	}
}

func TestGetDatabaseDSNPostgres(t *testing.T) {
	cfg := &Config{
		Username:  "test",
		Password:  "pass",
		SecretKey: "key",
		Database: DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "myuser",
			Password: "mypass",
			Name:     "mydb",
			SSLMode:  "disable",
		},
	}
	dsn, err := cfg.GetDatabaseDSN()
	if err != nil {
		t.Fatalf("GetDatabaseDSN() error: %v", err)
	}
	expected := "host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=disable"
	if dsn != expected {
		t.Errorf("DSN = %s, want %s", dsn, expected)
	}
}

func TestGetDatabaseDSNPostgresMissingFields(t *testing.T) {
	cfg := &Config{
		Username:  "test",
		Password:  "pass",
		SecretKey: "key",
		Database: DatabaseConfig{
			Type: "postgres",
		},
	}
	_, err := cfg.GetDatabaseDSN()
	if err == nil {
		t.Fatal("GetDatabaseDSN() expected error for missing postgres fields")
	}
}

func TestGetDatabaseDSNCustomDSN(t *testing.T) {
	cfg := &Config{
		Username:  "test",
		Password:  "pass",
		SecretKey: "key",
		Database: DatabaseConfig{
			Type: "postgres",
			DSN:  "custom://connection-string",
		},
	}
	dsn, err := cfg.GetDatabaseDSN()
	if err != nil {
		t.Fatalf("GetDatabaseDSN() error: %v", err)
	}
	if dsn != "custom://connection-string" {
		t.Errorf("DSN = %s, want custom://connection-string", dsn)
	}
}

func TestGetDatabaseDSNUnsupportedType(t *testing.T) {
	cfg := &Config{
		Username:  "test",
		Password:  "pass",
		SecretKey: "key",
		Database: DatabaseConfig{
			Type: "mysql",
		},
	}
	_, err := cfg.GetDatabaseDSN()
	if err == nil {
		t.Fatal("GetDatabaseDSN() expected error for unsupported type")
	}
}

func TestGetDatabaseDSNSQLiteMissingFile(t *testing.T) {
	cfg := &Config{
		Username:  "test",
		Password:  "pass",
		SecretKey: "key",
		Database: DatabaseConfig{
			Type:   "sqlite3",
			DBFile: "",
		},
	}
	_, err := cfg.GetDatabaseDSN()
	if err == nil {
		t.Fatal("GetDatabaseDSN() expected error for empty DBFile")
	}
}
