package gocket

import (
	"bytes"
	"strconv"

	"github.com/Falokut/go-kit/json"
	"github.com/pkg/errors"
)

const (
	strDelim         = ";|"
	noBodyEventParts = 2
)

var (
	delimiter = []byte(strDelim) // nolint:gochecknoglobals
)

type Message struct {
	Event string
	Data  json.RawMessage `json:",omitempty"`
	AckId uint64          `json:",omitempty"`
}

func (m *Message) IsAckRequired() bool {
	return m.AckId > 0
}

// nolint:mnd
func (m *Message) Unmarshal(data []byte) error {
	parts := bytes.SplitN(data, delimiter, 3)
	if len(parts) < noBodyEventParts {
		return errors.Errorf("expected format: <eventName>;|<ackId>;|<data>, got: %s", data)
	}
	m.Event = string(parts[0])

	ackId, err := strconv.ParseUint(string(parts[1]), 10, 64)
	if err != nil {
		return errors.WithMessage(err, "parse ackId")
	}
	m.AckId = ackId

	if len(parts) > noBodyEventParts {
		m.Data = parts[2]
	}

	return nil
}

func (m *Message) Marshal() []byte {
	buff := bytes.NewBuffer(nil)
	m.Encode(buff)
	return buff.Bytes()
}

func (m *Message) Encode(buff *bytes.Buffer) {
	ackId := strconv.FormatUint(m.AckId, 10)
	buff.Grow(len(m.Event) + len(ackId) + len(m.Data) + len(delimiter)*2)
	buff.WriteString(m.Event)
	buff.Write(delimiter)
	buff.WriteString(ackId)
	if len(m.Data) > 0 {
		buff.Write(delimiter)
		buff.Write(m.Data)
	}
}
