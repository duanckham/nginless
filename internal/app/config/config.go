package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// LogConfig ...
type LogConfig struct {
	Path       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

// Config ...
type Config struct {
	Version string
	Log     LogConfig
}

// ReadConfig read config from JSON file.
func ReadConfig() Config {
	c := Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		panic(fmt.Errorf("fatal error config unmarshal: %s", err))
	}

	return c
}
