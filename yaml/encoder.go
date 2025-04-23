package yaml

import (
	"io"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v any) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = e.w.Write(data)
	return err
}
