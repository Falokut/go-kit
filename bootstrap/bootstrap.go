package bootstrap

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	stdlog "log"

	"github.com/Falokut/go-kit/app"
	"github.com/Falokut/go-kit/cluster"
	"github.com/Falokut/go-kit/healthcheck"
	"github.com/Falokut/go-kit/infra"
	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/remote"
	"github.com/Falokut/go-kit/validator"

	"github.com/pkg/errors"
)

const (
	bootstrapLogFatalDelay = 500 * time.Millisecond
)

type Bootstrap struct {
	App                 *app.Application
	ClusterCli          *cluster.Client
	RemoteConfig        *remote.Config
	InfraServer         *infra.Server
	HealthcheckRegistry *healthcheck.Registry

	BindingAddress string
	MigrationsDir  string
	ModuleName     string
}

func New(moduleVersion string, remoteConfig any, endpoints []cluster.EndpointDescriptor) *Bootstrap {
	isDev := strings.ToLower(os.Getenv("APP_MODE")) == "dev"
	appConfig, err := appConfig(isDev)
	if err != nil {
		stdlog.Fatal(errors.WithMessage(err, "app config"))
	}

	app, err := app.NewFromConfig(*appConfig)
	if err != nil {
		stdlog.Fatal(errors.WithMessage(err, "create app"))
		return nil
	}

	localConfig, err := localConfig(app.Config())
	if err != nil {
		app.Logger().Fatal(app.Context(), errors.WithMessage(err, "create local config"))
	}

	boot, err := bootstrap(isDev, app, *localConfig, moduleVersion, remoteConfig, endpoints)
	if err != nil {
		err = errors.WithMessage(err, "create bootstrap")
		app.Logger().Fatal(app.Context(), err)
	}
	return boot
}

func bootstrap(isDev bool,
	application *app.Application,
	localConfig LocalConfig,
	moduleVersion string,
	remoteConfig any,
	endpoints []cluster.EndpointDescriptor,
) (*Bootstrap, error) {
	broadcastHost := localConfig.OuterAddress.Ip
	var err error
	if broadcastHost == "" {
		broadcastHost, err = resolveHost(localConfig.ConfigServiceAddresses[0])
		if err != nil {
			return nil, errors.WithMessage(err, "resolve local host")
		}
	}

	moduleInfo := cluster.ModuleInfo{
		ModuleName:    localConfig.ModuleName,
		ModuleVersion: moduleVersion,
		LibVersion:    kitVersion(),
		OuterAddress: cluster.AddressConfiguration{
			Ip:   broadcastHost,
			Port: strconv.Itoa(localConfig.OuterAddress.Port),
		},
		Endpoints: endpoints,
	}

	schema := remote.GenerateConfigSchema(remoteConfig)
	schemaData, err := json.Marshal(schema)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal config schema")
	}

	defaultConfig, err := readDefaultRemoteConfig(isDev, localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "read default remote config")
	}
	configData := cluster.ConfigData{
		Version: moduleVersion,
		Schema:  schemaData,
		Config:  defaultConfig,
	}

	cluster.RegisterSecretSubstrings(schema.Secrets)

	clusterCli := cluster.NewClient(
		moduleInfo,
		configData,
		localConfig.ConfigServiceAddresses,
		application.Logger(),
	)

	rc := remote.New(validator.Default, []byte(localConfig.RemoteConfigOverride))

	bindingAddress := net.JoinHostPort(localConfig.InnerAddress.Ip, strconv.Itoa(localConfig.InnerAddress.Port))

	migrationsDir, err := migrationsDirPath(isDev, localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve migrations dir path")
	}

	healthcheckRegistry := healthcheck.NewRegistry(application.Logger())
	healthcheckRegistry.Register("configServiceConnection", clusterCli)

	infraServer := infraServer(localConfig, application)
	infraServer.Handle("/internal/health", healthcheckRegistry.Handler())

	return &Bootstrap{
		App:                 application,
		ClusterCli:          clusterCli,
		RemoteConfig:        rc,
		BindingAddress:      bindingAddress,
		ModuleName:          localConfig.ModuleName,
		MigrationsDir:       migrationsDir,
		InfraServer:         infraServer,
		HealthcheckRegistry: healthcheckRegistry,
	}, nil
}

func (b *Bootstrap) Fatal(err error) {
	b.App.Close()
	time.Sleep(bootstrapLogFatalDelay)
	b.App.Logger().Fatal(context.Background(), err)
}
