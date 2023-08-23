package utils

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment          string `mapstructure:"ENVIRONMENT"`
	HTTPServerAddress    string `mapstructure:"HTTP_SERVER_ADDRESS"`
	KServeAddress        string `mapstructure:"KSERVE_ADDRESS"`
	KServeCustomDomain   string `mapstructure:"KSERVE_CUSTOM_DOMAIN"`
	KServeRequestTimeout int    `mapstructure:"KSERVE_REQUEST_TIMEOUT"`
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
