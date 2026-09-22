package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `env:"APP_ENV" envDefault:"development"`

	Logger struct {
		Level string `env:"LOG_LEVEL" envDefault:"info"`
	}

	App struct {
		Port int `env:"APP_PORT" validate:"required,min=1,max=65535"`
	}

	DB struct {
		Name     string `env:"DB_NAME" validate:"required"`
		Host     string `env:"DB_HOST" validate:"required,hostname|ip"`
		Port     int    `env:"DB_PORT" env-default:"5432" validate:"required"`
		User     string `env:"DB_USER" validate:"required"`
		Password string `env:"DB_PASSWORD" validate:"required"`
	}

	Redis struct {
		Host     string `env:"REDIS_HOST" validate:"required,hostname|ip"`
		Port     int    `env:"REDIS_PORT" validate:"required,min=1,max=65535"`
		Password string `env:"REDIS_PASSWORD" validate:"required"`
	}

	Email struct {
		ResendAPIKey string `env:"EMAIL_RESEND_API_KEY" validate:"required"`
		FromEmail    string `env:"EMAIL_FROM" validate:"required,email"`
	}

	Files struct {
		UploadDir string `env:"FILE_UPLOAD_DIR" validate:"required,dir"`
	}
}

func LoadConfig() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		err = cleanenv.ReadEnv(".env")
		if err != nil {
			return nil, fmt.Errorf("error loading config: %s", err)
		}
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("error validating config: %s", err)
	}

	return &cfg, nil
}
