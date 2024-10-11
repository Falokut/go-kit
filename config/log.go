package config

type Log struct {
	LogLevel      string `yaml:"level" env:"LOG_LEVEL" validate:"oneof=panic fatal error warning warn info debug trace"`
	ConsoleOutput bool   `yaml:"console_output" env:"LOG_CONSOLE_OUTPUT"`
	Filepath      string `yaml:"filepath" env:"LOG_FILEPATH" validate:"filepath|len=0"`
}
