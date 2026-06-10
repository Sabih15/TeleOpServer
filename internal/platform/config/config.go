package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	MQTT     MQTTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type MQTTConfig struct {
	Broker   string
	Username string
	Password string
	ClientID string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	expiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=teleopserver port=5432 sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "change-me-in-production"),
			ExpiryHours: expiryHours,
		},
		MQTT: MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", ""),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
			ClientID: getEnv("MQTT_CLIENT_ID", ""),
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
