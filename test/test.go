package test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/Falokut/go-kit/config"
	"github.com/Falokut/go-kit/log"
	"github.com/stretchr/testify/require"
)

type Test struct {
	id         string
	t          *testing.T
	cfg        *config.Config
	logger     *log.Adapter
	assertions *require.Assertions
}

func New(t *testing.T) (*Test, *require.Assertions) {
	t.Helper()

	assert := require.New(t)
	logger := log.New(log.WithLevel(log.DebugLevel), log.WithEncoder(log.JsonEncoder{}))

	cfg, err := config.New()
	assert.NoError(err)

	idBytes := make([]byte, 4) // nolint:mnd
	_, err = rand.Read(idBytes)
	assert.NoError(err)

	return &Test{
		id:         hex.EncodeToString(idBytes),
		t:          t,
		cfg:        cfg,
		logger:     logger,
		assertions: assert,
	}, assert
}

func (t *Test) Config() *config.Config {
	return t.cfg
}

func (t *Test) Logger() *log.Adapter {
	return t.logger
}

func (t *Test) Assert() *require.Assertions {
	return t.assertions
}

func (t *Test) Id() string {
	return t.id
}

func (t *Test) T() *testing.T {
	return t.t
}
