package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	AppEnv         string
	AppKey         string
	AdminUsername  string
	AdminPassword  string
	DBConnection   string
	DBHost         string
	DBPort         string
	DBDatabase     string
	DBUsername     string
	DBPassword     string
	DBSSLMode      string
	OpenAIEndpoint       string
	OpenAIKey            string
	OpenAIModel          string
	SearchEngineEndpoint string
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func LoadConfig(envPath ...string) *Config {
	if len(envPath) > 0 && envPath[0] != "" {
		_ = godotenv.Load(envPath[0])
	} else {
		_ = godotenv.Load()
	}

	return &Config{
		AppPort:        getEnv("APP_PORT", "7071"),
		AppEnv:         getEnv("APP_ENV", "development"),
		AppKey:         getEnv("APP_KEY", "pabwe-secret-key-2026-todolist-golang-gin"),
		AdminUsername:  getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "admin"),
		DBConnection:   getEnv("DB_CONNECTION", "pgsql"),
		DBHost:         getEnv("DB_HOST", "127.0.0.1"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBDatabase:     getEnv("DB_DATABASE", "pabwe_p9_todolist"),
		DBUsername:     getEnv("DB_USERNAME", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		OpenAIEndpoint:       getEnv("OPENAI_ENDPOINT", "https://delcom.org/api/v1"),
		OpenAIKey:            getEnv("OPENAI_KEY", "key"),
		OpenAIModel:          getEnv("OPENAI_MODEL", "nama-model"),
		SearchEngineEndpoint: getEnv("SEARCH_ENGINE_ENDPOINT", "https://id.wikipedia.org/w/api.php"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		c.DBHost, c.DBUsername, c.DBPassword, c.DBDatabase, c.DBPort, c.DBSSLMode,
	)
}
