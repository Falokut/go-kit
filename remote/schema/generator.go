package schema

type Generator struct {
	Reflector *Reflector
}

func NewGenerator() *Generator {
	return &Generator{
		Reflector: &Reflector{
			FieldNameReflector: GetNameAndRequiredFlag,
			FieldReflector:     SetProperties,
			ExpandedStruct:     true,
			DoNotReference:     true,
		},
	}
}

func (g *Generator) Generate(obj any) *Schema {
	return g.Reflector.Reflect(obj)
}
