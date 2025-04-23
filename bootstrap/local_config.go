package bootstrap

type LocalConfig struct {
	ConfigServiceAddresses  []string `validate:"required,dive,hostport"`
	OuterAddress            OuterAddr
	InnerAddress            InnerAddr
	ModuleName              string `validate:"required"`
	DefaultRemoteConfigPath string
	MigrationsDirPath       string
	RemoteConfigOverride    string
	LogFile                 LogFile
	InfraServerPort         int
}

type LogFile struct {
	Path       string
	MaxSizeMb  int
	MaxBackups int
	Compress   bool
}

type OuterAddr struct {
	Ip   string `validate:"hostname|ip"`
	Port int    `validate:"required"`
}

type InnerAddr struct {
	Ip   string `validate:"hostname|ip"`
	Port int    `validate:"required"`
}
