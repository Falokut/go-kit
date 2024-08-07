package log

type Fields map[string]any

type Field struct {
	Name  string
	Value any
}

func Any(name string, value any) Field {
	return Field{
		Name:  name,
		Value: value,
	}
}

// toFieldsMap converts a slice of Field structs into a map of string keys and any values.
func toFieldsMap(fields ...Field) Fields {
	res := make(Fields)
	for _, field := range fields {
		res[field.Name] = field.Value
	}
	return res
}
