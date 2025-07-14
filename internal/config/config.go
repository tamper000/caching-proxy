package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"

	internalErrors "github.com/tamper000/caching-proxy/internal/errors"
	"github.com/tamper000/caching-proxy/internal/models"
	"github.com/tamper000/caching-proxy/internal/utils"
)

func LoadConfig() (*models.Config, error) {
	// Init
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// Server section
	server := viper.Sub("server")
	server.SetDefault("port", "8080")

	origin := server.GetString("origin")
	if origin == "" {
		return nil, internalErrors.ErrOriginServer
	}

	secret := server.GetString("secret")
	if secret == "" {
		return nil, internalErrors.ErrSecret
	}

	port := server.GetString("port")

	// Redis section
	redis := loadRedis()

	// Regexp section
	blacklist := viper.GetStringSlice("blacklist")
	regexpList := utils.GenerateRegexp(blacklist)

	// Logger section
	logger := LoadLogger()

	// Loading values
	return &models.Config{
		Origin:     origin,
		Port:       port,
		Secret:     secret,
		Redis:      redis,
		RegexpList: regexpList,
		Logger:     logger,
	}, nil
}

func loadRedis() models.Redis {
	redis := viper.Sub("redis")
	redis.SetDefault("port", "6379")
	redis.SetDefault("db", "0")
	redis.SetDefault("server", "localhost")
	redis.SetDefault("enabled", false)

	return models.Redis{
		Port:     redis.GetString("port"),
		Addr:     redis.GetString("server"),
		Password: redis.GetString("password"),
		DB:       redis.GetInt("db"),
		TTL:      time.Duration(redis.GetInt("TTL")) * time.Minute,
	}
}

func LoadLogger() models.Logger {
	logger := viper.Sub("logger")
	logger.SetDefault("level", "INFO")
	logger.SetDefault("file", "app.log")

	level := logger.GetString("level")
	level = strings.ToUpper(level)

	return models.Logger{
		Level: level,
		File:  logger.GetString("file"),
	}
}
