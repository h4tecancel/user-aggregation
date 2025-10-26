package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App        App        `yaml:"app"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Storage    Storage    `yaml:"storage"`
}

type App struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type HTTPServer struct {
	Address         string        `yaml:"address"`
	Timeout         time.Duration `yaml:"timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type Storage struct {
	DBURL string `yaml:"db_url"`
}

func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/config.yaml"
	}
	if _, err := os.Stat(path); err != nil {
		panic(fmt.Errorf("config file %q not accessible: %w", path, err))
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(fmt.Errorf("read config %q: %w", path, err))
	}

	if err := cfg.validate(); err != nil {
		panic(fmt.Errorf("invalid config: %w", err))
	}
	return &cfg
}

func (c *Config) IsProd() bool  { return c != nil && c.App.Env == "prod" }
func (c *Config) IsLocal() bool { return c != nil && c.App.Env == "local" }

func (c *Config) validate() error {
	if c.App.Name == "" {
		return errors.New("app.name is required")
	}
	if c.App.Env == "" {
		return errors.New("app.env is required (local|prod)")
	}
	if c.HTTPServer.Address == "" {
		return errors.New("http_server.address is required")
	}
	if c.Storage.DBURL == "" {
		return errors.New("DBURL is required")
	}
	return nil
}
