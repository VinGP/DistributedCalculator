package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"sync"
)

type Config struct {
	Env             string `env:"ENV" env-default:"local"`
	Version         string `env:"VERSION" env-default:"1"`
	ComputingPower  int    `env:"COMPUTING_POWER" env-default:"2"`
	OrchestratorURL string `env:"API_URL" env-default:"http://localhost"`
}

var (
	config Config
	once   sync.Once
)

func Get() *Config {
	once.Do(func() {
		err := godotenv.Load()

		if err != nil {
			log.Println("error loading .env file")
		}
		err = cleanenv.ReadEnv(&config)
		if err != nil {
			panic(fmt.Sprintf("Failed to get config: %s", err))
		}
	})
	return &config
}
