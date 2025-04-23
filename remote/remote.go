package remote

import "github.com/Falokut/go-kit/remote/schema"

func GenerateConfigSchema(cfgPtr any) *schema.Schema {
	s := schema.NewGenerator().Generate(cfgPtr)
	s.Title = "Remote config"
	s.Version = ""
	return s
}
