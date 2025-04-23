// nolint:gochecknoglobals
package schema

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var Version = "v1"

type Schema struct {
	Version              string                                  `json:"$schema,omitempty"`
	Id                   Id                                      `json:"$id,omitempty"`
	Anchor               string                                  `json:"$anchor,omitempty"`
	Ref                  string                                  `json:"$ref,omitempty"`
	DynamicRef           string                                  `json:"$dynamicRef,omitempty"`
	Definitions          Definitions                             `json:"$defs,omitempty"`
	Comments             string                                  `json:"$comment,omitempty"`
	AllOf                []*Schema                               `json:"allOf,omitempty"`
	AnyOf                []*Schema                               `json:"anyOf,omitempty"`
	OneOf                []*Schema                               `json:"oneOf,omitempty"`
	Not                  *Schema                                 `json:"not,omitempty"`
	If                   *Schema                                 `json:"if,omitempty"`
	Then                 *Schema                                 `json:"then,omitempty"`
	Else                 *Schema                                 `json:"else,omitempty"`
	DependentSchemas     map[string]*Schema                      `json:"dependentSchemas,omitempty"`
	PrefixItems          []*Schema                               `json:"prefixItems,omitempty"`
	Items                *Schema                                 `json:"items,omitempty"`
	Contains             *Schema                                 `json:"contains,omitempty"`
	Properties           *orderedmap.OrderedMap[string, *Schema] `json:"properties,omitempty"`
	PatternProperties    map[string]*Schema                      `json:"patternProperties,omitempty"`
	AdditionalProperties *Schema                                 `json:"additionalProperties,omitempty"`
	PropertyNames        *Schema                                 `json:"propertyNames,omitempty"`
	Type                 string                                  `json:"type,omitempty"`
	Enum                 []any                                   `json:"enum,omitempty"`
	Const                any                                     `json:"const,omitempty"`
	MultipleOf           *int64                                  `json:"multipleOf,omitempty"`
	Maximum              *int64                                  `json:"maximum,omitempty"`
	ExclusiveMaximum     *int64                                  `json:"exclusiveMaximum,omitempty"`
	Minimum              *int64                                  `json:"minimum,omitempty"`
	ExclusiveMinimum     *int64                                  `json:"exclusiveMinimum,omitempty"`
	MaxLength            *uint64                                 `json:"maxLength,omitempty"`
	MinLength            *uint64                                 `json:"minLength,omitempty"`
	Pattern              string                                  `json:"pattern,omitempty"`
	MaxItems             *uint64                                 `json:"maxItems,omitempty"`
	MinItems             *uint64                                 `json:"minItems,omitempty"`
	UniqueItems          bool                                    `json:"uniqueItems,omitempty"`
	MaxContains          *uint64                                 `json:"maxContains,omitempty"`
	MinContains          *uint64                                 `json:"minContains,omitempty"`
	MaxProperties        *uint64                                 `json:"maxProperties,omitempty"`
	MinProperties        *uint64                                 `json:"minProperties,omitempty"`
	Secrets              []string                                `json:"secrets,omitempty"`
	Required             []string                                `json:"required,omitempty"`
	DependentRequired    map[string][]string                     `json:"dependentRequired,omitempty"`
	Format               string                                  `json:"format,omitempty"`
	ContentEncoding      string                                  `json:"contentEncoding,omitempty"`
	ContentMediaType     string                                  `json:"contentMediaType,omitempty"`
	ContentSchema        *Schema                                 `json:"contentSchema,omitempty"`
	Title                string                                  `json:"title,omitempty"`
	Description          string                                  `json:"description,omitempty"`
	Default              any                                     `json:"default,omitempty"`
	Deprecated           bool                                    `json:"deprecated,omitempty"`
	ReadOnly             bool                                    `json:"readOnly,omitempty"`
	WriteOnly            bool                                    `json:"writeOnly,omitempty"`
	Examples             []any                                   `json:"examples,omitempty"`

	Extras map[string]any `json:"-"`

	boolean *bool
}

var (
	TrueSchema  = &Schema{boolean: &[]bool{true}[0]}
	FalseSchema = &Schema{boolean: &[]bool{false}[0]}
)

type Definitions map[string]*Schema
