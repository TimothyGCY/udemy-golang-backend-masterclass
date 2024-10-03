package util

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	App      app      `mapstructure:"app"`
	Database database `mapstructure:"database"`
	Paseto   paseto   `mapstructure:"paseto"`
}

type app struct {
	Port string `mapstructure:"port"`
}

type database struct {
	Driver string `mapstructure:"driver"`
	Uri    string `mapstructure:"uri"`
}

type paseto struct {
	SymmetricToken     string        `mapstructure:"symmetric-token"`
	AccessTokenTimeout time.Duration `mapstructure:"access-token-timeout"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err == nil {
		err = viper.Unmarshal(&config)
	}

	return
}
