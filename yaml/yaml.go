package yaml

import (
	"github.com/Falokut/go-kit/utils/cases"
	"github.com/goccy/go-yaml"
)

// nolint:gochecknoglobals
var KeyTransform = cases.ToLowerCamelCase

func Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func Unmarshal(data []byte, output any) error {
	return yaml.Unmarshal(data, output)
}

func UnmarshalToMap(data []byte) (map[string]any, error) {
	var raw map[string]any
	err := yaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
