package config

import (
	"carsawa/utils/email"
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	AppPort           string `mapstructure:"APP_PORT"`
	DatabaseURL       string `mapstructure:"DATABASE_URL"`
	Env               string `mapstructure:"ENV"`
	JWTSecret         string `mapstructure:"JWT_SECRET"`
	LogLevel          string `mapstructure:"LOG_LEVEL"`
	MaxRequestsPerMin int    `mapstructure:"MAX_REQUESTS_PER_MIN"`

	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisCacheDB  int    `mapstructure:"REDIS_CACHE_DB"`
	RedisAuthDB   int    `mapstructure:"REDIS_AUTH_DB"`
	RedisOTPDB    int    `mapstructure:"REDIS_OTP_DB"`

	GoogleAPIKey             string `mapstructure:"GOOGLE_API_KEY"`
	GoogleServiceAccountFile string `mapstructure:"GOOGLE_SERVICE_ACCOUNT_FILE"`

	SMTPHost        string `mapstructure:"SMTP_HOST"`
	SMTPPort        int    `mapstructure:"SMTP_PORT"`
	SMTPUser        string `mapstructure:"SMTP_USER"`
	SMTPPassword    string `mapstructure:"SMTP_PASSWORD"`
	SMTPFrom        string `mapstructure:"SMTP_FROM"`
	SMTPTimeoutSecs int    `mapstructure:"SMTP_TIMEOUT_SECS"`
}

// AppConfig is the global configuration instance.
var AppConfig Config

// LoadConfig reads from config.yaml (if present) and environment variables.
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("MAX_REQUESTS_PER_MIN", 100)
	viper.SetDefault("DATABASE_URL", "mongodb://localhost:27017")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_CACHE_DB", 0)
	viper.SetDefault("REDIS_AUTH_DB", 1)
	viper.SetDefault("REDIS_OTP_DB", 2)
	viper.SetDefault("GOOGLE_API_KEY", "")
	viper.SetDefault("GOOGLE_SERVICE_ACCOUNT_FILE", "")

	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("SMTP_USER", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_FROM", "no-reply@carsawa.com")
	viper.SetDefault("SMTP_TIMEOUT_SECS", 10)

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No config file found, using environment variables")
	}
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

// GetEnv returns the current environment (development, staging, production).
func GetEnv() string {
	return AppConfig.Env
}

// IsProduction returns true in production environment.
func IsProduction() bool {
	return AppConfig.Env == "production"
}

// SMTPConfig builds the email.SMTPConfig from AppConfig.
func SMTPConfig() email.SMTPConfig {
	return email.SMTPConfig{
		Host:     AppConfig.SMTPHost,
		Port:     AppConfig.SMTPPort,
		Username: AppConfig.SMTPUser,
		Password: AppConfig.SMTPPassword,
		From:     AppConfig.SMTPFrom,
		Timeout:  time.Duration(AppConfig.SMTPTimeoutSecs) * time.Second,
	}
}
