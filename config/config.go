package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Mode struct {
		Mode       string
		Type       string
		ACKNoDelay bool
		Mtu        int
	}
	Kcp struct {
		ACKNoDelay bool
		Mtu        int
		Internal   int
	}
	Server struct {
		Address string
		Key     string
	}
	Agent struct {
		Address string
		Key     string
		Ports   []string
	}
}

func LoadConfig() (config *Config) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		panic("ERROR load config file!")
	}
	log.Println("================ Loaded Configuration ================")
	return
}
