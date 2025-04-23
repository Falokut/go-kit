package yaml

import (
	"bytes"
	"io"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v any) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, d.r)
	if err != nil {
		return err
	}
	return Unmarshal(buf.Bytes(), v)
}
