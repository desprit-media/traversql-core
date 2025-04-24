package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct{}

var instance Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		err := cleanenv.ReadEnv(&instance)
		if err != nil {
			log.Fatalln(err)
		}
	})
	return &instance
}
