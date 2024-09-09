package config

type Log struct {
	LogLevel      string `yaml:"level" env:"LOG_LEVEL"`
	ConsoleOutput bool   `yaml:"console_output" env:"LOG_CONSOLE_OUTPUT"`
	Filepath      string `yaml:"filepath" env:"LOG_FILEPATH"`
}
