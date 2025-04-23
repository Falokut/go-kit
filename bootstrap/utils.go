package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"runtime/debug"

	"github.com/Falokut/go-kit/app"
	"github.com/Falokut/go-kit/config"
	"github.com/Falokut/go-kit/infra"
	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/log/file"
	"github.com/Falokut/go-kit/validator"
	"github.com/pkg/errors"
)

func localConfig(config *config.Config) (*LocalConfig, error) {
	localConfig := LocalConfig{}
	err := config.Read(&localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "read local config")
	}
	if localConfig.InnerAddress.Port != localConfig.OuterAddress.Port {
		return nil, errors.Errorf("innerAddress.port is not equal outerAddress.port. potential mistake")
	}
	return &localConfig, nil
}

func kitVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "0.0.0"
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/Falokut/go-kit" {
			return dep.Version
		}
	}
	return "0.0.0"
}

// nolint:mnd
func appConfig(isDev bool) (*app.Config, error) {
	localConfigPath, err := configFilePath(isDev)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve local config path")
	}
	configsOpts := []config.Option{
		config.WithValidator(validator.Default),
		config.WithEnvPrefix(os.Getenv("APP_CONFIG_ENV_PREFIX")),
	}
	if localConfigPath != "" {
		configsOpts = append(configsOpts, config.WithExtraSource(config.NewYamlConfig(localConfigPath)))
	}

	logConfigSupplier := app.LoggerConfigSupplier(func(cfg *config.Config) app.LogConfig {
		logCfg := app.LogConfig{
			EncoderType:       app.EncoderType(cfg.Optional().String("LOG.ENCODER_TYPE", string(app.JsonEncoderType))),
			DeduplicateFields: cfg.Optional().Bool("LOG.DEDUPLICATE_FIELDS", true),
		}

		logFilePath := cfg.Optional().String("LOGFILE.PATH", "")
		if !isDev && logFilePath != "" {
			logCfg.FileOutput = &file.Config{
				Filepath:   logFilePath,
				MaxSizeMb:  cfg.Optional().Int("LOGFILE.MAXSIZEMB", maxLogFileSizeInMb),
				MaxBackups: cfg.Optional().Int("LOGFILE.MAXBACKUPS", 4),
				Compress:   cfg.Optional().Bool("LOGFILE.COMPRESS", true),
			}
		}

		return logCfg
	})

	return &app.Config{
		ConfigOptions:        configsOpts,
		LoggerConfigSupplier: logConfigSupplier,
	}, nil
}

func readDefaultRemoteConfig(isDev bool, cfg LocalConfig) (json.RawMessage, error) {
	path, err := defaultRemoteConfigPath(isDev, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithMessage(err, "read file")
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, nil
	}

	var raw json.RawMessage
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal default config")
	}

	return raw, nil
}

func defaultRemoteConfigPath(isDev bool, cfg LocalConfig) (string, error) {
	if cfg.DefaultRemoteConfigPath != "" {
		return cfg.DefaultRemoteConfigPath, nil
	}

	if isDev {
		return "conf/default_remote_config.json", nil
	}

	return relativePathFromBin("default_remote_config.json")
}

func configFilePath(isDev bool) (string, error) {
	cfgPath := os.Getenv("APP_CONFIG_PATH")
	if cfgPath != "" {
		return cfgPath, nil
	}

	if isDev {
		return "./conf/dev_config.yml", nil
	}

	return relativePathFromBin("config.yml")
}

func migrationsDirPath(isDev bool, cfg LocalConfig) (string, error) {
	if cfg.MigrationsDirPath != "" {
		return cfg.MigrationsDirPath, nil
	}

	if isDev {
		return "./migrations", nil
	}

	return relativePathFromBin("migrations")
}

func relativePathFromBin(part string) (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", errors.WithMessage(err, "get executable path")
	}
	return path.Join(path.Dir(ex), part), nil
}

// nolint:forcetypeassert
func resolveHost(target string) (string, error) {
	conn, err := net.Dial("udp", target)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil
}

func infraServer(localConfig LocalConfig, application *app.Application) *infra.Server {
	infraServer := infra.NewServer()
	infraServerPort := localConfig.InnerAddress.Port + 1
	if localConfig.InfraServerPort != 0 {
		infraServerPort = localConfig.InfraServerPort
	}
	infraServerAddress := fmt.Sprintf(":%d", infraServerPort)
	application.AddRunners(app.RunnerFunc(func(ctx context.Context) error {
		application.Logger().Info(ctx, "run infra server", log.String("infraServerAddress", infraServerAddress))
		err := infraServer.ListenAndServe(infraServerAddress)
		if err != nil {
			return errors.WithMessagef(err, "run infra server on %s", infraServerAddress)
		}
		return nil
	}))
	application.AddClosers(app.CloserFunc(func(context.Context) error {
		return infraServer.Shutdown()
	}))
	return infraServer
}
