package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName(".bb-lite")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("")
	viper.BindEnv("api_key", "ALPHA_VANTAGE_API_KEY")

	viper.ReadInConfig() // ignore error — env var may suffice
}

func GetAPIKey() (string, error) {
	key := viper.GetString("api_key")
	if key == "" {
		return "", fmt.Errorf("API key not configured. Set api_key in ~/.bb-lite.yaml or ALPHA_VANTAGE_API_KEY env var")
	}
	return key, nil
}
