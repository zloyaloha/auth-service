package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env 			string			`yaml:"env" env-default:"local"`
	PGConnectionConfig PGConnectionConfig `yaml:"pgconnection"`
	GRPC 			GRPCConfig		`yaml:"grpc"`
	MigrationsPath 	string
	TokenTTL 		time.Duration 	`yaml:"token_ttl" env-default:"1h"`
}

type GRPCConfig struct {
	Port 			int				`yaml:"port"`
	Timeout 		time.Duration	`yaml:"timeout"`
}

type PGConnectionConfig struct {
	Name     string		`yaml:"postgres_db"`
	User     string		`yaml:"postgres_user"`
	Password string		`yaml:"postgres_password"`
	Host     string		`yaml:"postgres_host"`
	Port     string		`yaml:"postgres_port"`
}

func MustLoad() Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
	}

	return cfg
}


func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}