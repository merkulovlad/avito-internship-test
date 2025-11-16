// Package config provides configuration loading and management.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Database DatabaseConfig
}

// ServerConfig contains server-related configuration.
type ServerConfig struct {
	Host string
	Port int
}

// LogConfig contains logging configuration.
type LogConfig struct {
	Filename  string
	Level     string
	ToConsole bool
}

// DatabaseConfig contains database connection configuration.
type DatabaseConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxConnections    int
	ConnectionTimeout int
	ConnMaxLifetime   int
}

// mustGetEnv retrieves the value of the environment variable named by the key.
// It panics if the variable is not set.
func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var %s", key)
	}

	return v
}

// mustGetEnvBool retrieves a boolean environment variable.
// It panics if the variable is not set or cannot be parsed.
func mustGetEnvBool(key string) bool {
	s := mustGetEnv(key)

	b, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatalf("invalid boolean for %s: %v", key, err)
	}

	return b
}

// mustGetEnvInt retrieves an integer environment variable.
// It panics if the variable is not set or cannot be parsed.
func mustGetEnvInt(key string) int {
	s := mustGetEnv(key)

	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid integer for %s: %v", key, err)
	}

	return i
}

// MustLoad loads the configuration from environment variables.
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
			ConnMaxLifetime:   mustGetEnvInt("POSTGRES_CONN_MAX_LIFETIME"),
		},
	}
}

// DSN constructs the Data Source Name for database connection.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
