// nolint
package schema

import (
	"bytes"
	"net"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/utils/cases"
	"slices"
)

type customSchemaImpl interface {
	Schema() *Schema
}

type extendSchemaImpl interface {
	SchemaExtend(schema *Schema)
}

type aliasSchemaImpl interface {
	SchemaAlias() any
}

type propertyAliasSchemaImpl interface {
	SchemaProperty(prop string) any
}

var (
	customAliasSchema         = reflect.TypeOf((*aliasSchemaImpl)(nil)).Elem()
	customPropertyAliasSchema = reflect.TypeOf((*propertyAliasSchemaImpl)(nil)).Elem()

	customType = reflect.TypeOf((*customSchemaImpl)(nil)).Elem()
	extendType = reflect.TypeOf((*extendSchemaImpl)(nil)).Elem()
)

type customSchemaGetFieldDocString interface {
	GetFieldDocString(fieldName string) string
}

type customGetFieldDocString func(fieldName string) string

var customStructGetFieldDocString = reflect.TypeOf((*customSchemaGetFieldDocString)(nil)).Elem()

func Reflect(v any) *Schema {
	return ReflectFromType(reflect.TypeOf(v))
}

func ReflectFromType(t reflect.Type) *Schema {
	r := &Reflector{}
	return r.ReflectFromType(t)
}

type Reflector struct {
	BaseSchemaId              Id
	Anonymous                 bool
	AssignAnchor              bool
	AllowAdditionalProperties bool
	RequiredFromSchemaTags    bool
	DoNotReference            bool
	ExpandedStruct            bool
	FieldNameReflector        func(f reflect.StructField) (string, bool)
	FieldReflector            func(f reflect.StructField, s *Schema)
	FieldNameTag              string
	IgnoredTypes              []any
	Lookup                    func(reflect.Type) Id
	Mapper                    func(reflect.Type) *Schema
	Namer                     func(reflect.Type) string
	KeyNamer                  func(string) string
	AdditionalFields          func(reflect.Type) []reflect.StructField
	CommentMap                map[string]string
}

func (r *Reflector) Reflect(v any) *Schema {
	return r.ReflectFromType(reflect.TypeOf(v))
}

func (r *Reflector) ReflectFromType(t reflect.Type) *Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := getDefinitionName(t)

	s := new(Schema)
	definitions := Definitions{}
	s.Definitions = definitions
	bs := r.reflectTypeToSchemaWithId(definitions, t)
	if r.ExpandedStruct {
		*s = *definitions[name]
		delete(definitions, name)
	} else {
		*s = *bs
	}

	if !r.Anonymous && s.Id == EmptyId {
		baseSchemaId := r.BaseSchemaId
		if baseSchemaId == EmptyId {
			id := Id("https://" + t.PkgPath())
			if err := id.Validate(); err == nil {
				baseSchemaId = id
			}
		}
		if baseSchemaId != EmptyId {
			s.Id = baseSchemaId.Add(cases.ToSnakeCase(name))
		}
	}

	s.Version = Version
	if !r.DoNotReference {
		s.Definitions = definitions
	}

	return s
}

var (
	timeType = reflect.TypeOf(time.Time{})
	ipType   = reflect.TypeOf(net.IP{})
	uriType  = reflect.TypeOf(url.URL{})

	byteSliceType  = reflect.TypeOf([]byte(nil))
	rawMessageType = reflect.TypeOf(json.RawMessage{})
)

type protoEnum interface {
	EnumDescriptor() ([]byte, []int)
}

var protoEnumType = reflect.TypeOf((*protoEnum)(nil)).Elem()

func (r *Reflector) SetBaseSchemaId(id string) {
	r.BaseSchemaId = Id(id)
}

func (r *Reflector) refOrReflectTypeToSchema(definitions Definitions, t reflect.Type) *Schema {
	id := r.lookupId(t)
	if id != EmptyId {
		return &Schema{
			Ref: id.String(),
		}
	}

	if def := r.refDefinition(definitions, t); def != nil {
		return def
	}

	return r.reflectTypeToSchemaWithId(definitions, t)
}

func (r *Reflector) reflectTypeToSchemaWithId(defs Definitions, t reflect.Type) *Schema {
	s := r.reflectTypeToSchema(defs, t)
	if s != nil {
		if r.Lookup != nil {
			id := r.Lookup(t)
			if id != EmptyId {
				s.Id = id
			}
		}
	}
	return s
}

func (r *Reflector) reflectTypeToSchema(definitions Definitions, t reflect.Type) *Schema {
	if t.Kind() == reflect.Ptr {
		return r.refOrReflectTypeToSchema(definitions, t.Elem())
	}

	if t.Implements(customAliasSchema) {
		v := reflect.New(t)
		o := v.Interface().(aliasSchemaImpl)
		t = reflect.TypeOf(o.SchemaAlias())
		return r.refOrReflectTypeToSchema(definitions, t)
	}

	if r.Mapper != nil {
		if t := r.Mapper(t); t != nil {
			return t
		}
	}
	if rt := r.reflectCustomSchema(definitions, t); rt != nil {
		return rt
	}

	st := new(Schema)

	if t.Implements(protoEnumType) {
		st.OneOf = []*Schema{
			{Type: "string"},
			{Type: "integer"},
		}
		return st
	}

	if t == ipType {
		st.Type = "string"
		st.Format = "ipv4"
		return st
	}

	switch t.Kind() {
	case reflect.Struct:
		r.reflectStruct(definitions, t, st)

	case reflect.Slice, reflect.Array:
		r.reflectSliceOrArray(definitions, t, st)

	case reflect.Map:
		r.reflectMap(definitions, t, st)

	case reflect.Interface:

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		st.Type = "integer"

	case reflect.Float32, reflect.Float64:
		st.Type = "number"

	case reflect.Bool:
		st.Type = "boolean"

	case reflect.String:
		st.Type = "string"

	default:
		panic("unsupported type " + t.String())
	}

	r.reflectSchemaExtend(definitions, t, st)

	def := r.refDefinition(definitions, t)
	if def != nil {
		return def
	}

	return st
}

func (r *Reflector) reflectCustomSchema(definitions Definitions, t reflect.Type) *Schema {
	if t.Kind() == reflect.Ptr {
		return r.reflectCustomSchema(definitions, t.Elem())
	}

	if t.Implements(customType) {
		v := reflect.New(t)
		o := v.Interface().(customSchemaImpl)
		st := o.Schema()
		r.addDefinition(definitions, t, st)
		if ref := r.refDefinition(definitions, t); ref != nil {
			return ref
		}
		return st
	}

	return nil
}

func (r *Reflector) reflectSchemaExtend(definitions Definitions, t reflect.Type, s *Schema) *Schema {
	if t.Implements(extendType) {
		v := reflect.New(t)
		o := v.Interface().(extendSchemaImpl)
		o.SchemaExtend(s)
		if ref := r.refDefinition(definitions, t); ref != nil {
			return ref
		}
	}

	return s
}

func (r *Reflector) reflectSliceOrArray(definitions Definitions, t reflect.Type, st *Schema) {
	if t == rawMessageType {
		return
	}

	r.addDefinition(definitions, t, st)

	if st.Description == "" {
		st.Description = r.lookupComment(t, "")
	}

	if t.Kind() == reflect.Array {
		l := uint64(t.Len())
		st.MinItems = &l
		st.MaxItems = &l
	}
	if t.Kind() == reflect.Slice && t.Elem() == byteSliceType.Elem() {
		st.Type = "string"
		st.ContentEncoding = "base64"
	} else {
		st.Type = "array"
		st.Items = r.refOrReflectTypeToSchema(definitions, t.Elem())
	}
}

func (r *Reflector) reflectMap(definitions Definitions, t reflect.Type, st *Schema) {
	r.addDefinition(definitions, t, st)

	st.Type = "object"
	if st.Description == "" {
		st.Description = r.lookupComment(t, "")
	}

	switch t.Key().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		st.PatternProperties = map[string]*Schema{
			"^[0-9]+$": r.refOrReflectTypeToSchema(definitions, t.Elem()),
		}
		st.AdditionalProperties = FalseSchema
		return
	}
	if t.Elem().Kind() != reflect.Interface {
		st.AdditionalProperties = r.refOrReflectTypeToSchema(definitions, t.Elem())
	}
}

func (r *Reflector) reflectStruct(definitions Definitions, t reflect.Type, s *Schema) {
	switch t {
	case timeType:
		s.Type = "string"
		s.Format = "date-time"
		return
	case uriType:
		s.Type = "string"
		s.Format = "uri"
		return
	}

	r.addDefinition(definitions, t, s)
	s.Type = "object"
	s.Properties = NewProperties()
	s.Description = r.lookupComment(t, "")
	if r.AssignAnchor {
		s.Anchor = t.Name()
	}
	if !r.AllowAdditionalProperties && s.AdditionalProperties == nil {
		s.AdditionalProperties = FalseSchema
	}

	ignored := false
	for _, it := range r.IgnoredTypes {
		if reflect.TypeOf(it) == t {
			ignored = true
			break
		}
	}
	if !ignored {
		r.reflectStructFields(s, definitions, t)
	}
}

func (r *Reflector) reflectStructFields(st *Schema, definitions Definitions, t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	var getFieldDocString customGetFieldDocString
	if t.Implements(customStructGetFieldDocString) {
		v := reflect.New(t)
		o := v.Interface().(customSchemaGetFieldDocString)
		getFieldDocString = o.GetFieldDocString
	}

	customPropertyMethod := func(string) any {
		return nil
	}
	if t.Implements(customPropertyAliasSchema) {
		v := reflect.New(t)
		o := v.Interface().(propertyAliasSchemaImpl)
		customPropertyMethod = o.SchemaProperty
	}

	handleField := func(f reflect.StructField) {
		name, shouldEmbed, required, nullable, isSecret := r.reflectFieldName(f)
		if name == "" {
			if shouldEmbed {
				r.reflectStructFields(st, definitions, f.Type)
			}
			return
		}

		var property *Schema
		if alias := customPropertyMethod(name); alias != nil {
			property = r.refOrReflectTypeToSchema(definitions, reflect.TypeOf(alias))
		} else {
			property = r.refOrReflectTypeToSchema(definitions, f.Type)
		}

		property.structKeywordsFromTags(f, st, name)
		if property.Description == "" {
			property.Description = r.lookupComment(t, f.Name)
		}
		if getFieldDocString != nil {
			property.Description = getFieldDocString(f.Name)
		}

		if nullable {
			property = &Schema{
				OneOf: []*Schema{
					property,
					{
						Type: "null",
					},
				},
			}
		}

		if r.FieldReflector != nil {
			r.FieldReflector(f, property)
		}

		st.Properties.Set(name, property)
		if required {
			st.Required = appendUniqueString(st.Required, name)
		}

		if isSecret {
			st.Secrets = appendUniqueString(st.Secrets, name)
		}
	}

	for i := range t.NumField() {
		f := t.Field(i)
		handleField(f)
	}
	if r.AdditionalFields != nil {
		if af := r.AdditionalFields(t); af != nil {
			for _, sf := range af {
				handleField(sf)
			}
		}
	}
}

func appendUniqueString(base []string, value string) []string {
	for _, v := range base {
		if v == value {
			return base
		}
	}
	return append(base, value)
}

func (r *Reflector) lookupComment(t reflect.Type, name string) string {
	if r.CommentMap == nil {
		return ""
	}

	n := fullyQualifiedTypeName(t)
	if name != "" {
		n = n + "." + name
	}

	return r.CommentMap[n]
}

func (r *Reflector) addDefinition(definitions Definitions, t reflect.Type, s *Schema) {
	name := getDefinitionName(t)
	if name == "" {
		return
	}
	definitions[name] = s
}

func (r *Reflector) refDefinition(definitions Definitions, t reflect.Type) *Schema {
	if r.DoNotReference {
		return nil
	}

	name := getDefinitionName(t)
	if name == "" {
		return nil
	}
	if _, ok := definitions[name]; !ok {
		return nil
	}
	return &Schema{
		Ref: "#/$defs/" + name,
	}
}

func (r *Reflector) lookupId(t reflect.Type) Id {
	if r.Lookup != nil {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return r.Lookup(t)
	}
	return EmptyId
}

func (t *Schema) structKeywordsFromTags(f reflect.StructField, parent *Schema, propertyName string) {
	t.Description = f.Tag.Get("schema_description")

	tags := splitOnUnescapedCommas(f.Tag.Get(tagSchema))
	tags = t.genericKeywords(tags, parent, propertyName)

	switch t.Type {
	case "string":
		t.stringKeywords(tags)
	case "number":
		t.numericalKeywords(tags)
	case "integer":
		t.numericalKeywords(tags)
	case "array":
		t.arrayKeywords(tags)
	case "boolean":
		t.booleanKeywords(tags)
	}
	extras := strings.Split(f.Tag.Get("schema_extras"), ",")
	t.extraKeywords(extras)
}

func (t *Schema) genericKeywords(tags []string, parent *Schema, propertyName string) []string { //nolint:gocyclo
	unprocessed := make([]string, 0, len(tags))
	for _, tag := range tags {
		nameValue := strings.SplitN(tag, "=", 2)
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "title":
				t.Title = val
			case "description":
				t.Description = val
			case "type":
				t.Type = val
			case "anchor":
				t.Anchor = val
			case "oneof_required":
				var typeFound *Schema
				for i := range parent.OneOf {
					if parent.OneOf[i].Title == nameValue[1] {
						typeFound = parent.OneOf[i]
					}
				}
				if typeFound == nil {
					typeFound = &Schema{
						Title:    nameValue[1],
						Required: []string{},
					}
					parent.OneOf = append(parent.OneOf, typeFound)
				}
				typeFound.Required = append(typeFound.Required, propertyName)
			case "anyof_required":
				var typeFound *Schema
				for i := range parent.AnyOf {
					if parent.AnyOf[i].Title == nameValue[1] {
						typeFound = parent.AnyOf[i]
					}
				}
				if typeFound == nil {
					typeFound = &Schema{
						Title:    nameValue[1],
						Required: []string{},
					}
					parent.AnyOf = append(parent.AnyOf, typeFound)
				}
				typeFound.Required = append(typeFound.Required, propertyName)
			case "oneof_ref":
				subSchema := t
				if t.Items != nil {
					subSchema = t.Items
				}
				if subSchema.OneOf == nil {
					subSchema.OneOf = make([]*Schema, 0, 1)
				}
				subSchema.Ref = ""
				refs := strings.Split(nameValue[1], ";")
				for _, r := range refs {
					subSchema.OneOf = append(subSchema.OneOf, &Schema{
						Ref: r,
					})
				}
			case "oneof_type":
				if t.OneOf == nil {
					t.OneOf = make([]*Schema, 0, 1)
				}
				t.Type = ""
				types := strings.Split(nameValue[1], ";")
				for _, ty := range types {
					t.OneOf = append(t.OneOf, &Schema{
						Type: ty,
					})
				}
			case "anyof_ref":
				subSchema := t
				if t.Items != nil {
					subSchema = t.Items
				}
				if subSchema.AnyOf == nil {
					subSchema.AnyOf = make([]*Schema, 0, 1)
				}
				subSchema.Ref = ""
				refs := strings.Split(nameValue[1], ";")
				for _, r := range refs {
					subSchema.AnyOf = append(subSchema.AnyOf, &Schema{
						Ref: r,
					})
				}
			case "anyof_type":
				if t.AnyOf == nil {
					t.AnyOf = make([]*Schema, 0, 1)
				}
				t.Type = ""
				types := strings.Split(nameValue[1], ";")
				for _, ty := range types {
					t.AnyOf = append(t.AnyOf, &Schema{
						Type: ty,
					})
				}
			default:
				unprocessed = append(unprocessed, tag)
			}
		}
	}
	return unprocessed
}

func (t *Schema) booleanKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.Split(tag, "=")
		if len(nameValue) != 2 {
			continue
		}
		name, val := nameValue[0], nameValue[1]
		if name == "default" {
			if val == "true" {
				t.Default = true
			} else if val == "false" {
				t.Default = false
			}
		}
	}
}

func (t *Schema) stringKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.SplitN(tag, "=", 2)
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "minLength":
				t.MinLength = parseUint(val)
			case "maxLength":
				t.MaxLength = parseUint(val)
			case "pattern":
				t.Pattern = val
			case "format":
				switch val {
				case "date-time", "email", "hostname", "ipv4", "ipv6", "uri", "uuid":
					t.Format = val
				}
			case "readOnly":
				i, _ := strconv.ParseBool(val)
				t.ReadOnly = i
			case "writeOnly":
				i, _ := strconv.ParseBool(val)
				t.WriteOnly = i
			case "default":
				t.Default = val
			case "example":
				t.Examples = append(t.Examples, val)
			case "enum":
				t.Enum = append(t.Enum, val)
			}
		}
	}
}

func (t *Schema) numericalKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.Split(tag, "=")
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "multipleOf":
				t.MultipleOf = parseInt(val)
			case "minimum":
				t.Minimum = parseInt(val)
			case "maximum":
				t.Maximum = parseInt(val)
			case "exclusiveMaximum":
				t.ExclusiveMaximum = parseInt(val)
			case "exclusiveMinimum":
				t.ExclusiveMinimum = parseInt(val)
			case "default":
				num := parseInt(val)
				if num != nil {
					t.Default = *num
				}
			case "example":
				num := parseInt(val)
				if num != nil {
					t.Examples = append(t.Examples, *num)
				}
			case "enum":
				num := parseInt(val)
				if num != nil {
					t.Enum = append(t.Enum, *num)
				}
			}
		}
	}
}

func (t *Schema) arrayKeywords(tags []string) {
	var defaultValues []any

	unprocessed := make([]string, 0, len(tags))
	for _, tag := range tags {
		nameValue := strings.Split(tag, "=")
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "minItems":
				t.MinItems = parseUint(val)
			case "maxItems":
				t.MaxItems = parseUint(val)
			case "uniqueItems":
				t.UniqueItems = true
			case "default":
				defaultValues = append(defaultValues, val)
			case "format":
				t.Items.Format = val
			case "pattern":
				t.Items.Pattern = val
			default:
				unprocessed = append(unprocessed, tag) // left for further processing by underlying type
			}
		}
	}
	if len(defaultValues) > 0 {
		t.Default = defaultValues
	}

	if len(unprocessed) == 0 {
		return
	}

	switch t.Items.Type {
	case "string":
		t.Items.stringKeywords(unprocessed)
	case "number":
		t.Items.numericalKeywords(unprocessed)
	case "integer":
		t.Items.numericalKeywords(unprocessed)
	case "array":
	case "boolean":
		t.Items.booleanKeywords(unprocessed)
	}
}

func (t *Schema) extraKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.SplitN(tag, "=", 2)
		if len(nameValue) == 2 {
			t.setExtra(nameValue[0], nameValue[1])
		}
	}
}

func (t *Schema) setExtra(key, val string) {
	if t.Extras == nil {
		t.Extras = map[string]any{}
	}
	existingVal, ok := t.Extras[key]
	if ok {
		switch existingVal := existingVal.(type) {
		case string:
			t.Extras[key] = []string{existingVal, val}
		case []string:
			t.Extras[key] = append(existingVal, val)
		case int:
			t.Extras[key], _ = strconv.Atoi(val)
		case bool:
			t.Extras[key] = (val == "true" || val == "t")
		}
		return
	}

	switch key {
	case "minimum":
		t.Extras[key], _ = strconv.Atoi(val)
	default:
		var x any
		if val == "true" {
			x = true
		} else if val == "false" {
			x = false
		} else {
			x = val
		}
		t.Extras[key] = x
	}
}

func requiredFromTags(tags []string, val *bool) {
	if ignoredByTags(tags) {
		return
	}

	for _, tag := range tags[1:] {
		if tag == "omitempty" {
			*val = false
			return
		}
	}
	*val = true
}

func requiredFromSchemaTags(tags []string, val *bool) {
	for _, tag := range tags {
		if tag == "required" {
			*val = true
		}
	}
}

func nullableFromSchemaTags(tags []string) bool {
	return slices.Contains(tags, "nullable")
}

func isSecretFromSchemaTags(tags []string) bool {
	return slices.Contains(tags, "secret")
}

func ignoredByTags(tags []string) bool {
	return tags[0] == "-"
}

func ignoredBySchemaTags(tags []string) bool {
	return tags[0] == "-"
}

func (r *Reflector) fieldNameTag() string {
	if r.FieldNameTag != "" {
		return r.FieldNameTag
	}
	return "json"
}

func (r *Reflector) reflectFieldName(f reflect.StructField) (string, bool, bool, bool, bool) {
	name, shouldEmbed, isRequired, isNullable, isSecret := r.defaultReflectFieldName(f)
	if r.FieldNameReflector != nil {
		name, isRequired = r.FieldNameReflector(f)
	}
	return name, shouldEmbed, isRequired, isNullable, isSecret
}

func (r *Reflector) defaultReflectFieldName(f reflect.StructField) (string, bool, bool, bool, bool) {
	tagString := f.Tag.Get(r.fieldNameTag())
	tags := strings.Split(tagString, ",")

	if ignoredByTags(tags) {
		return "", false, false, false, false
	}

	schemaTags := strings.Split(f.Tag.Get(tagSchema), ",")
	if ignoredBySchemaTags(schemaTags) {
		return "", false, false, false, false
	}

	var required bool
	if !r.RequiredFromSchemaTags {
		requiredFromTags(tags, &required)
	} else {
		requiredFromSchemaTags(schemaTags, &required)
	}

	nullable := nullableFromSchemaTags(schemaTags)
	isSecret := isSecretFromSchemaTags(schemaTags)

	if f.Anonymous && tags[0] == "" {
		if f.Type.Kind() == reflect.Struct {
			return "", true, false, false, false
		}

		if f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct {
			return "", true, false, false, false
		}
	}

	name := f.Name
	if tags[0] != "" {
		name = tags[0]
	}
	if !f.Anonymous && f.PkgPath != "" {
		name = ""
	} else if r.KeyNamer != nil {
		name = r.KeyNamer(name)
	}

	return name, false, required, nullable, isSecret
}

func (t *Schema) UnmarshalJson(data []byte) error {
	if bytes.Equal(data, []byte("true")) {
		*t = *TrueSchema
		return nil
	} else if bytes.Equal(data, []byte("false")) {
		*t = *FalseSchema
		return nil
	}
	type SchemaAlt Schema
	aux := &struct {
		*SchemaAlt
	}{
		SchemaAlt: (*SchemaAlt)(t),
	}
	return json.Unmarshal(data, aux)
}

func (t *Schema) MarshalJson() ([]byte, error) {
	if t.boolean != nil {
		if *t.boolean {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	}
	if reflect.DeepEqual(&Schema{}, t) {
		return []byte("true"), nil
	}
	type SchemaAlt Schema
	b, err := json.Marshal((*SchemaAlt)(t))
	if err != nil {
		return nil, err
	}
	if len(t.Extras) == 0 {
		return b, nil
	}
	m, err := json.Marshal(t.Extras)
	if err != nil {
		return nil, err
	}
	if len(b) == 2 {
		return m, nil
	}
	b[len(b)-1] = ','
	return append(b, m[1:]...), nil
}

func splitOnUnescapedCommas(tagString string) []string {
	ret := make([]string, 0)
	separated := strings.Split(tagString, ",")
	ret = append(ret, separated[0])
	i := 0
	for _, nextTag := range separated[1:] {
		if len(ret[i]) == 0 {
			ret = append(ret, nextTag)
			i++
			continue
		}

		if ret[i][len(ret[i])-1] == '\\' {
			ret[i] = ret[i][:len(ret[i])-1] + "," + nextTag
		} else {
			ret = append(ret, nextTag)
			i++
		}
	}

	return ret
}

func fullyQualifiedTypeName(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

func getDefinitionName(t reflect.Type) string {
	name := t.Name()
	if name == "" {
		return ""
	}
	return path.Base(t.PkgPath()) + name
}

func (r *Reflector) AddGoComments(base, path string) error {
	if r.CommentMap == nil {
		r.CommentMap = make(map[string]string)
	}
	return ExtractGoComments(base, path, r.CommentMap)
}
