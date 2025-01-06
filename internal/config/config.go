package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AuthKey string
	DBPath  string
	Port    string
}

var AppConfig Config

func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	AppConfig.AuthKey = getEnv("AUTH_KEY", "omgwow")
	AppConfig.DBPath = getEnv("DB_PATH", "./catbox-scanner-db.db")
	AppConfig.Port = getEnv("PORT", "6969")

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
