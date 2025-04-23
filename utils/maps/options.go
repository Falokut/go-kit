package maps

type flattenExpandOptions struct {
	prefix string
	sep    string
}

type option func(o *flattenExpandOptions)

func WithPrefix(prefix string) option {
	return func(o *flattenExpandOptions) {
		o.prefix = prefix
	}
}

func WithSep(sep string) option {
	return func(o *flattenExpandOptions) {
		o.sep = sep
	}
}
