package util

import "github.com/spf13/viper"

type Config struct {
	App      app      `mapstructure:"app"`
	Database database `mapstructure:"database"`
}

type app struct {
	Port string `mapstructure:"port"`
}

type database struct {
	Driver string `mapstructure:"driver"`
	Uri    string `mapstructure:"uri"`
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
