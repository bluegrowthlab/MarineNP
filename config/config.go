/*
 * MarineNP Configuration
 * Purpose: Application configuration management and environment variable handling
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file manages the application's configuration, including server settings,
 * database connection parameters, and API configuration. It supports both
 * SQLite and PostgreSQL databases.
 */

package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config represents the complete application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	API      APIConfig
	Version  string
	LastUpdate string
}

// ServerConfig contains server-specific configuration parameters
type ServerConfig struct {
	Port int    // HTTP server port
	Env  string // Environment (development/production)
}

// DatabaseConfig contains database connection parameters
type DatabaseConfig struct {
	Type     string // "sqlite" or "postgres"
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	Path     string // Path to SQLite database file
}

// APIConfig contains API-related configuration settings
type APIConfig struct {
	Prefix        string
	CorsAllowOrigin string
}

// GetDSN returns the appropriate database connection string based on the database type
func (c *DatabaseConfig) GetDSN() string {
	if c.Type == "sqlite" {
		return c.Path
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// LoadConfig initializes and returns the application configuration
// It loads settings from environment variables with sensible defaults
func LoadConfig() *Config {
	// Attempt to load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Parse port numbers from environment variables
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	return &Config{
		Server: ServerConfig{
			Port: port,
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "marinenp"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			Path:     getEnv("DB_PATH", "marinenp-sqlite.db"),
		},
		API: APIConfig{
			Prefix:          getEnv("API_PREFIX", "/api/v1"),
			CorsAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
		},
		Version:    getEnv("APP_VERSION", "1.0.0"),
		LastUpdate: getEnv("LAST_UPDATE", "2025-06-08"),
	}
}

// getEnv retrieves an environment variable value or returns a default if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 