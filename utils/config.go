package utils

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment          string `mapstructure:"ENVIRONMENT"`
	HTTPServerAddress    string `mapstructure:"HTTP_SERVER_ADDRESS"`
	ServingAgentAddress  string `mapstructure:"SERVING_AGENT_ADDRESS"`
	WebhookServerAddress string `mapstructure:"WEBHOOK_SERVER_ADDRESS"`
	WebhookAPIKey        string `mapstructure:"WEBHOOK_APIKEY"`
	// For rate limiter
	RedisAddress       string `mapstructure:"REDIS_ADDRESS"`
	FormattedRateSync  string `mapstructure:"FORMATTED_RATE_SYNC"`
	FormattedRateAsync string `mapstructure:"FORMATTED_RATE_ASYNC"`
	FormattedRateTask  string `mapstructure:"FORMATTED_RATE_TASK"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
