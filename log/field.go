package log

type Field struct {
	Name string
	Type FieldType

	Int       int64
	String    string
	Interface any
}
