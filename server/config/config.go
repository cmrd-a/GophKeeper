package config

import (
	"errors"
	"log/slog"

	"github.com/spf13/viper"

	"github.com/cmrd-a/GophKeeper/server/logger"
)

type Config struct {
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	GRPCPort    int16  `mapstructure:"GRPC_PORT"`
	HTTPPort    int16  `mapstructure:"HTTP_PORT"`
	DatabaseURI string `mapstructure:"DATABASE_URI"`
	SaltSecret  string `mapstructure:"SALT_SECRET"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
}

func NewConfig(log *slog.Logger, lvl *slog.LevelVar) (*Config, error) {
	viper.SetDefault("LOG_LEVEL", "DEBUG")
	viper.SetDefault("GRPC_PORT", "8082")
	viper.SetDefault("HTTP_PORT", "8080")

	viper.SetDefault("SALT_SECRET", "changeme")
	viper.SetDefault("JWT_SECRET", "changeme")

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("../../.")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Info("No .env file found, relying on environment variables.")
		} else {
			log.Error("Error reading config file", "error", err)
			return nil, err
		}
	}
	config := Config{}

	if err := viper.Unmarshal(&config); err != nil {
		log.Error("Unable to decode config into struct", "error", err)
		return nil, err
	}
	newLvl := logger.GetLogLevelFromEnv(config.LogLevel)
	lvl.Set(newLvl)

	log.Info("Configuration loaded",
		"LogLevel", config.LogLevel,
		"HTTPPort", config.HTTPPort,
		"DatabaseURI", config.DatabaseURI,
	)
	return &config, nil
}
