package config

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	DB_USERNAME     string
	DB_PASSWORD     string
	DB_HOST         string
	DB_NAME         string
	NATS_HOSTS      string
	NATS_CLUSTER_ID string
	NATS_CLIENT_ID  string
	NATS_SUBJECT    string
	CACHE_SIZE      string
	APP_KEY         string
}

func ConfigSetup() {
	bs, err := ioutil.ReadFile("cmd/config/config.toml")
	if err != nil {
		log.Fatalf("Ошибка чтения конфиг-файла: /cmd/config/config.toml: %v\n", err)
	}

	cfg := Config{}
	err = toml.Unmarshal(bs, &cfg)
	if err != nil {
		log.Fatalf("Ошибка десиариализации toml-документа\n")
	}

	// Database settings
	os.Setenv("DB_USERNAME", cfg.DB_USERNAME)
	os.Setenv("DB_PASSWORD", cfg.DB_PASSWORD)
	os.Setenv("DB_HOST", cfg.DB_HOST)
	os.Setenv("DB_NAME", cfg.DB_NAME)

	// NATS-Streaming settings
	os.Setenv("NATS_HOSTS", cfg.NATS_HOSTS)
	os.Setenv("NATS_CLUSTER_ID", cfg.NATS_CLUSTER_ID)
	os.Setenv("NATS_CLIENT_ID", cfg.NATS_CLIENT_ID)
	os.Setenv("NATS_SUBJECT", cfg.NATS_SUBJECT)

	// Cache settings
	os.Setenv("CACHE_SIZE", cfg.CACHE_SIZE)
	os.Setenv("APP_KEY", cfg.APP_KEY)
}
