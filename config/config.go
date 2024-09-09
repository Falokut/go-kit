package config

import (
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type Config struct {
	Log Log `yaml:"log"`
}

func Read(ptr any) error {
	cfgPath := getLocalConfigPath()
	err := cleanenv.ReadConfig(cfgPath, ptr)
	if err != nil {
		help, _ := cleanenv.GetDescription(ptr, nil)
		return errors.WithMessage(err, help)
	}
	return nil
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
