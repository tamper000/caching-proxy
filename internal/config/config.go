package config

import (
	"time"

	"github.com/spf13/viper"

	"github.com/tamper000/caching-proxy/internal/models"
	"github.com/tamper000/caching-proxy/utils"
)

func LoadConfig() models.Config {
	// Init
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic("no config file")
	}

	// Server section
	server := viper.Sub("server")
	server.SetDefault("port", "8080")

	origin := server.GetString("origin")
	if origin == "" {
		panic("no origin server")
	}

	secret := server.GetString("secret")
	if secret == "" {
		panic("empty secret")
	}

	port := server.GetString("port")

	// Redis section
	redis := loadRedis()

	// Regexp section
	blacklist := viper.GetStringSlice("blacklist")
	regexpList := utils.GenerateRegexp(blacklist)

	// Loading values
	return models.Config{
		Origin:     origin,
		Port:       port,
		Secret:     secret,
		Redis:      redis,
		RegexpList: regexpList,
	}
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
