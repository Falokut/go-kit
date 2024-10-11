package config

import (
	"os"
	"strings"

	"github.com/Falokut/go-kit/validator"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type Config struct {
	Log Log `yaml:"log"`
}

func Read(ptr any) error {
	return ReadConfig(ptr, getLocalConfigPath())
}

func ReadConfig(ptr any, configPath string) error {
	err := cleanenv.ReadConfig(configPath, ptr)
	if err != nil {
		help, _ := cleanenv.GetDescription(ptr, nil)
		return errors.WithMessage(err, help)
	}
	return validator.Default.ValidateToError(ptr)
}

func getLocalConfigPath() string {
	appMode := os.Getenv("APP_MODE")
	if strings.EqualFold(appMode, "dev") {
		return "conf/dev_config.yml"
	}
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath != "" {
		return cfgPath
	}
	return "conf/config.yml"
}
