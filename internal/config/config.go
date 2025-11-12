package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type LogConfig struct {
	Filename  string
	Level     string
	ToConsole bool
}

type DatabaseConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxConnections    int
	ConnectionTimeout int
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var %s", key)
	}
	return v
}

func mustGetEnvBool(key string) bool {
	s := mustGetEnv(key)
	b, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatalf("invalid boolean for %s: %v", key, err)
	}
	return b
}

func mustGetEnvInt(key string) int {
	s := mustGetEnv(key)
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid integer for %s: %v", key, err)
	}
	return i
}

func MustLoad() *Config {
	// ignore error if there's no .env in CI/etc
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Host: mustGetEnv("BACKEND_HOST"),
			Port: mustGetEnvInt("BACKEND_PORT"),
		},
		Log: LogConfig{
			Filename:  mustGetEnv("LOG_FILE"),
			Level:     mustGetEnv("LOG_LEVEL"),
			ToConsole: mustGetEnvBool("LOG_TO_CONSOLE"),
		},
		Database: DatabaseConfig{
			Host:              mustGetEnv("POSTGRES_HOST"),
			Port:              mustGetEnvInt("POSTGRES_PORT"),
			User:              mustGetEnv("POSTGRES_USER"),
			Password:          mustGetEnv("POSTGRES_PASSWORD"),
			Name:              mustGetEnv("POSTGRES_DB"),
			SSLMode:           mustGetEnv("POSTGRES_SSLMODE"),
			MaxConnections:    mustGetEnvInt("POSTGRES_MAX_CONNECTIONS"),
			ConnectionTimeout: mustGetEnvInt("POSTGRES_CONNECTION_TIMEOUT"),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
