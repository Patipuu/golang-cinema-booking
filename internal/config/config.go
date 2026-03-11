package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds application configuration (loaded via Viper from env / .env).
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	OTP      OTPConfig
	SMTP     SMTPConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN returns PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type OTPConfig struct {
	ExpiryMinutes int
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// Load reads config from environment and optional .env file.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // .env is optional
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{
		Server: ServerConfig{
			Port: getStr(v, "PORT", "8080"),
			Env:  getStr(v, "ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getStr(v, "DB_HOST", "localhost"),
			Port:     getStr(v, "DB_PORT", "5432"),
			User:     getStr(v, "DB_USER", "postgres"),
			Password: getStr(v, "DB_PASSWORD", ""),
			Name:     getStr(v, "DB_NAME", "booking_cinema"),
			SSLMode:  getStr(v, "DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:      getStr(v, "JWT_SECRET", "change-me-in-production"),
			ExpiryHours: getInt(v, "JWT_EXPIRY_HOURS", 24),
		},
		OTP: OTPConfig{
			ExpiryMinutes: getInt(v, "OTP_EXPIRY_MINUTES", 10),
		},
		SMTP: SMTPConfig{
			Host:     getStr(v, "SMTP_HOST", "smtp.gmail.com"),
			Port:     getInt(v, "SMTP_PORT", 587),
			User:     getStr(v, "SMTP_USER", ""),
			Password: getStr(v, "SMTP_PASSWORD", ""),
			From:     getStr(v, "SMTP_FROM", "noreply@booking-cinema.local"),
		},
	}
	return cfg, nil
}

func getStr(v *viper.Viper, key, def string) string {
	if v.IsSet(key) {
		return v.GetString(key)
	}
	return def
}

func getInt(v *viper.Viper, key string, def int) int {
	if v.IsSet(key) {
		switch v.Get(key).(type) {
		case int:
			return v.GetInt(key)
		case string:
			n, _ := strconv.Atoi(v.GetString(key))
			return n
		}
	}
	return def
}
