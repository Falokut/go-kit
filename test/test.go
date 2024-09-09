package test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/Falokut/go-kit/log"
	"github.com/stretchr/testify/require"
)

type Test struct {
	id         string
	t          *testing.T
	logger     log.Logger
	assertions *require.Assertions
}

func New(t *testing.T) (*Test, *require.Assertions) {
	assert := require.New(t)
	logger := log.DefaultLoggerWithLevel(log.DebugLevel)

	idBytes := make([]byte, 4)
	_, err := rand.Read(idBytes)
	assert.NoError(err)
	return &Test{
		id:         hex.EncodeToString(idBytes),
		t:          t,
		logger:     logger,
		assertions: assert,
	}, assert
}

func (t *Test) Logger() log.Logger {
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
