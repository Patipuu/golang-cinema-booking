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
	Redis    RedisConfig
	VNPay    VNPayConfig
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

// RedisConfig lưu cấu hình kết nối Redis (dùng cho cache, idempotency, rate limit)
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`     // Địa chỉ Redis (ví dụ: localhost:6379)
	Password string `mapstructure:"password"` // Mật khẩu Redis (nếu có)
	DB       int    `mapstructure:"db"`       // Số DB (thường là 0)
}

// VNPayConfig lưu thông tin kết nối VNPay sandbox/production
type VNPayConfig struct {
	PayURL     string `mapstructure:"pay_url"`     // URL thanh toán (sandbox: https://sandbox.vnpayment.vn/paymentv2/vpcpay.html)
	TmnCode    string `mapstructure:"tmn_code"`    // Mã TMN từ VNPay
	HashSecret string `mapstructure:"hash_secret"` // Khóa hash (Hash Secret) từ VNPay
	ReturnURL  string `mapstructure:"return_url"`  // URL callback sau thanh toán (ví dụ: http://localhost:8080/api/payments/callback)
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
			Port: getStr(v, "PORT", "8081"),
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
		Redis: RedisConfig{
			Addr:     getStr(v, "REDIS_ADDR", "localhost:6379"),
			Password: getStr(v, "REDIS_PASSWORD", ""),
			DB:       getInt(v, "REDIS_DB", 0),
		},
		VNPay: VNPayConfig{
			PayURL:    getStr(v, "VNPAY_PAY_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"),
			TmnCode:   getStr(v, "VNPAY_TMN_CODE", "GJQAO3Y6"),
			HashSecret: getStr(v, "VNPAY_HASH_SECRET", "FVN21HE81C05RF4FIBIH8HIZZ4EMTU6D"),
			ReturnURL: getStr(v, "VNPAY_RETURN_URL", "http://localhost:8081/api/v1/payments/callback"),
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
