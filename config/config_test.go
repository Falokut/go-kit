package config_test

import (
	"testing"

	"github.com/Falokut/go-kit/config"
	"github.com/stretchr/testify/assert"
)

const baseCfgPath = "./test_data/"

func TestListenConfig_HappyPath(t *testing.T) {
	assert := assert.New(t)
	var listen config.Listen
	err := config.ReadConfig(&listen, baseCfgPath+"listen_valid_ip.yml")
	assert.NoError(err)
	expected := config.Listen{
		Host: "0.0.0.0",
		Port: 9909,
	}
	assert.Equal(expected, listen)
}

func TestListenConfig_EmptyIp(t *testing.T) {
	assert := assert.New(t)

	var listen config.Listen
	err := config.ReadConfig(&listen, baseCfgPath+"listen_valid_empty_ip.yml")
	assert.NoError(err)
	expected := config.Listen{
		Host: "0.0.0.0",
		Port: 8080,
	}
	assert.Equal(expected, listen)
}

func TestListenConfig_InvalidIp(t *testing.T) {
	assert := assert.New(t)
	var listen config.Listen
	err := config.ReadConfig(&listen, baseCfgPath+"listen_invalid_ip.yml")
	assert.Error(err)
}

func TestListenConfig_InvalidPort(t *testing.T) {
	assert := assert.New(t)
	var listen config.Listen
	err := config.ReadConfig(&listen, baseCfgPath+"listen_invalid_port.yml")
	assert.Error(err)
}
func TestListenConfig_Empty(t *testing.T) {
	assert := assert.New(t)
	var listen config.Listen
	err := config.ReadConfig(&listen, baseCfgPath+"empty.yml")
	assert.Error(err)
}

func TestDBConfig_HappyPath(t *testing.T) {
	assert := assert.New(t)
	var db config.Database
	err := config.ReadConfig(&db, baseCfgPath+"database_valid_host.yml")
	assert.NoError(err)
	expected := config.Database{
		Host:     "postgres",
		Port:     5432,
		Database: "postgres",
		Username: "user",
		Password: "pass",
	}
	assert.Equal(expected, db)
}

func TestDBConfig_Ip_HappyPath(t *testing.T) {
	assert := assert.New(t)
	var db config.Database
	err := config.ReadConfig(&db, baseCfgPath+"database_valid_ip.yml")
	assert.NoError(err)
	expected := config.Database{
		Host:     "193.0.22.4",
		Port:     5432,
		Database: "postgres",
		Username: "user",
		Password: "pass",
	}
	assert.Equal(expected, db)
}

func TestDBConfig_InvalidIp(t *testing.T) {
	assert := assert.New(t)
	var db config.Database
	err := config.ReadConfig(&db, baseCfgPath+"database_invalid_ip.yml")
	assert.Error(err)
}

func TestDBConfig_InvalidHost(t *testing.T) {
	assert := assert.New(t)
	var db config.Database
	err := config.ReadConfig(&db, baseCfgPath+"database_invalid_host.yml")
	assert.Error(err)
}
func TestDBConfig_Empty(t *testing.T) {
	assert := assert.New(t)
	var db config.Database
	err := config.ReadConfig(&db, baseCfgPath+"empty.yml")
	assert.Error(err)
}

func TestLogConfig_Console_HappyPath(t *testing.T) {
	assert := assert.New(t)
	var log config.Log
	err := config.ReadConfig(&log, baseCfgPath+"log_valid_level_console.yml")
	assert.NoError(err)
	expected := config.Log{
		LogLevel:      "debug",
		ConsoleOutput: true,
	}
	assert.Equal(expected, log)
}

func TestLogConfig_Filepath_HappyPath(t *testing.T) {
	assert := assert.New(t)
	var log config.Log
	err := config.ReadConfig(&log, baseCfgPath+"log_valid_level_file.yml")
	assert.NoError(err)
	expected := config.Log{
		LogLevel: "info",
		Filepath: "./some/dir/path/log.log",
	}
	assert.Equal(expected, log)
}

func TestLogConfig_InvalidLevel(t *testing.T) {
	assert := assert.New(t)
	var log config.Log
	err := config.ReadConfig(&log, baseCfgPath+"log_invalid_level.yml")
	assert.Error(err)
}

func TestLogConfig_InvalidFilepath(t *testing.T) {
	assert := assert.New(t)
	var log config.Log
	err := config.ReadConfig(&log, baseCfgPath+"log_invalid_filepath.yml")
	assert.Error(err)
}

func TestLogConfig_Empty(t *testing.T) {
	assert := assert.New(t)
	var log config.Log
	err := config.ReadConfig(&log, baseCfgPath+"empty.yml")
	assert.Error(err)
}
