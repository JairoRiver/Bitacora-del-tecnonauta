package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Host     string
	Port     string
	LogLevel string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// DSN construye el connection string para pgx.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type CDNConfig struct {
	Provider string
}

type AdminConfig struct {
	SessionDuration string
	AuthSecret      string
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	CDN      CDNConfig
	Admin    AdminConfig
}

// Load lee config.yaml y sobreescribe secretos con variables de entorno.
// Secretos esperados: DB_USER, DB_PASSWORD.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// Variables de entorno sobreescriben el yaml (útil en producción/CI)
	viper.AutomaticEnv()

	// Bindings explícitos para secretos que no siguen el patrón APP_*
	if err := viper.BindEnv("database.user", "DB_USER"); err != nil {
		return nil, fmt.Errorf("binding DB_USER: %w", err)
	}
	if err := viper.BindEnv("database.password", "DB_PASSWORD"); err != nil {
		return nil, fmt.Errorf("binding DB_PASSWORD: %w", err)
	}
	if err := viper.BindEnv("admin.auth_secret", "AUTH_SECRET"); err != nil {
		return nil, fmt.Errorf("binding AUTH_SECRET: %w", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host:     viper.GetString("server.host"),
			Port:     viper.GetString("server.port"),
			LogLevel: viper.GetString("server.log_level"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			Name:     viper.GetString("database.name"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		CDN: CDNConfig{
			Provider: viper.GetString("cdn.provider"),
		},
		Admin: AdminConfig{
			SessionDuration: viper.GetString("admin.session_duration"),
			AuthSecret:      viper.GetString("admin.auth_secret"),
		},
	}

	return cfg, nil
}
