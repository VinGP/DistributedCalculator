package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"sync"
)

type Config struct {
	Port                  string `env:"PORT" env-default:"80"`
	Host                  string `env:"HOST" env-default:"0.0.0.0"`
	Env                   string `env:"ENV" env-default:"local"`
	Version               string `env:"VERSION" env-default:"1"`
	TimeAdditionMs        int    `env:"TIME_ADDITION_MS" env-default:"1000"`
	TimeSubtractionMs     int    `env:"TIME_SUBTRACTION_MS" env-default:"1000"`
	TimeMultiplicationsMs int    `env:"TIME_MULTIPLICATIONS_MS" env-default:"1000"`
	TimeDivisionsMs       int    `env:"TIME_DIVISIONS_MS" env-default:"1000"`
	TimePowerMs           int    `env:"TIME_POWER_MS" env-default:"1000"`
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
