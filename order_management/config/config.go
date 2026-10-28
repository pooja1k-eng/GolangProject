package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GatewayPort        string
	ProductServicePort string
	ProductServiceURL  string
	OrderServicePort   string
	OrderServiceURL    string
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	LogLevel           string
	AuthToken          string
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env", "../.env")

	return &Config{
		GatewayPort:        getEnv("GATEWAY_PORT", "8082"),
		ProductServicePort: getEnv("PRODUCT_SERVICE_PORT", "8081"),
		ProductServiceURL:  getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		OrderServicePort:   getEnv("ORDER_SERVICE_PORT", "10004"),
		OrderServiceURL:    getEnv("ORDER_SERVICE_URL", "http://localhost:10004"),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBName:             getEnv("DB_NAME", "orderdb"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		AuthToken:          getEnv("AUTH_TOKEN", "secret-bearer-token"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}