package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string      `yaml:"env"`
	HTTP HTTPServer  `yaml:"http"`
}

type HTTPServer struct {
	Address            string        `yaml:"address" env-default:"0.0.0.0:8080"`
	AuthServiceAddress string        `yaml:"authaddress" env-default:"0.0.0.0:44044"`
	Timeout            time.Duration `yaml:"timeout" env-default:"5s"`
}

func MustLoad() Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("can't get env var")
	}

	if _, err := os.Stat(configPath); err != nil {
		panic("config file is not exists")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("can't read config")
	}

	return cfg
}

func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config path")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	return path
}