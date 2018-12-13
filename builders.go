package sashay

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

type baseBuilder struct {
	buf     io.Writer
	swagger *Sashay
}

func (b *baseBuilder) writeLn(indent int, format string, i ...interface{}) {
	for i := 0; i < indent; i++ {
		b.buf.Write([]byte("  "))
	}
	b.buf.Write([]byte(fmt.Sprintf(format, i...)))
	b.buf.Write([]byte("\n"))
}

func (b *baseBuilder) writeOnce(indent int, format string, i ...interface{}) func() {
	var ran = false
	return func() {
		if ran {
			return
		}
		b.writeLn(indent, format, i...)
		ran = true
	}
}

func (b *baseBuilder) writeNotEmpty(indent int, format string, s string) {
	if s != "" {
		b.writeLn(indent, format, s)
	}
}

func (b *baseBuilder) writeDataType(indent int, f Field) {
	dataTypeDef, found := b.swagger.dataTypeDefFor(f)
	if !found {
		panic(fmt.Sprintf("No dataTypeDef defined for kind %s, type %s. You should either change the type, "+
			"or add a custom data type mapper. See Representing Custom Types at "+
			"https://godoc.org/github.com/cloudability/sashay#hdr-Sashay_Detail__Representing_Custom_Types "+
			"for more information.", f.Kind.String(), f.Type.String()))
	}
	objectFields := ObjectFields{}
	dataTypeDef.DataTyper(f, objectFields)
	for _, kv := range objectFields.Sorted() {
		b.writeLn(indent, "%s: %s", kv[0], kv[1])
	}
}

// Write struct f and all its fields recursively.
// If recurse returns true for a struct field, call writeStructSchema on it.
// If it doesn't, write the field as concrete ($ref for data type).
func (b *baseBuilder) writeStructSchema(indent int, f Field, recurse func(Field) bool) {
	b.writeLn(indent, "type: object")
	writeProps := b.writeOnce(indent, "properties:")
	for _, field := range enumerateStructFields(f) {
		fieldJSONName := jsonName(field.StructField)
		if fieldJSONName == "" {
			continue
		}
		writeProps()
		if field.Kind == reflect.Struct {
			b.writeLn(indent+1, "%s:", fieldJSONName)
			if recurse(field) {
				b.writeStructSchema(indent+2, field, recurse)
			} else {
				b.writeRefSchema(indent+2, field)
			}
		} else if field.Kind == reflect.Slice {
			b.writeLn(indent+1, "%s:", fieldJSONName)
			b.writeLn(indent+2, "type: array")
			b.writeLn(indent+2, "items:")
			sliceField := ZeroSliceValueField(field.Type)
			if sliceField.Kind == reflect.Struct {
				if recurse(sliceField) {
					b.writeStructSchema(indent+3, sliceField, recurse)
				} else {
					b.writeRefSchema(indent+3, sliceField)
				}
			} else {
				b.writeDataType(indent+3, sliceField)
			}
		} else {
			b.writeLn(indent+1, "%s:", fieldJSONName)
			b.writeDataType(indent+2, field)
		}
	}
}

func (b *baseBuilder) writeRefSchema(indent int, f Field) {
	if f.Kind == reflect.Slice {
		b.writeLn(indent, "type: array")
		b.writeLn(indent, "items:")
		b.writeRefSchema(indent+1, ZeroSliceValueField(f.Type))
	} else if f.Kind == reflect.Struct {
		isEmptyStruct := f.Type.NumField() == 0
		if isEmptyStruct {
			b.writeLn(indent, "type: object")
		} else if b.swagger.isMappedToDataType(f) {
			b.writeDataType(indent, f)
		} else {
			b.writeLn(indent, "$ref: '%s'", schemaRefLink(f))
		}
	} else {
		b.writeDataType(indent, f)
	}
}

// Parse the struct field tag and pull out the JSON name.
// In general, this is only used when walking structs with JSON only,
// like ReturnErr/ReturnOK values, or Params values when it is meant for the request body.
// Recall that Params can be parsed from query/header/path as well,
// this _only_ returns the json tag name.
func jsonName(f reflect.StructField) string {
	jsonTag := f.Tag.Get("json")
	if jsonTag != "-" {
		parts := strings.Split(jsonTag, ",")
		if len(parts) > 1 && parts[0] == "" {
			return f.Name
		}
		return parts[0]
	}
	return ""
}

type docBuilder struct {
	base *baseBuilder
}

func (b *docBuilder) writeLn(indent int, format string, i ...interface{}) {
	b.base.writeLn(indent, format, i...)
}

func (b *docBuilder) writeInfo() {
	b.writeLn(0, "openapi: 3.0.0")
	b.writeLn(0, "info:")
	sw := b.base.swagger
	b.writeLn(1, "title: %s", sw.title)
	b.writeLn(1, "description: %s", sw.desc)
	b.base.writeNotEmpty(1, "termsOfService: %s", sw.tos)
	if sw.contactName != "" || sw.contactURL != "" || sw.contactEmail != "" {
		b.writeLn(1, "contact:")
		b.base.writeNotEmpty(2, "name: %s", sw.contactName)
		b.base.writeNotEmpty(2, "url: %s", sw.contactURL)
		b.base.writeNotEmpty(2, "email: %s", sw.contactEmail)
	}
	if sw.licenseName != "" || sw.licenseURL != "" {
		b.writeLn(1, "license:")
		b.base.writeNotEmpty(2, "name: %s", sw.licenseName)
		b.base.writeNotEmpty(2, "url: %s", sw.licenseURL)
	}
	b.writeLn(1, "version: %s", sw.version)
}

func (b *docBuilder) writeTags() {
	if len(b.base.swagger.tags) == 0 {
		return
	}
	b.writeLn(0, "tags:")
	for _, t := range b.base.swagger.tags {
		b.writeLn(1, "- name: %s", t.name)
		b.writeLn(1, "  description: %s", t.desc)
	}
}

func (b *docBuilder) writeServers() {
	if len(b.base.swagger.servers) == 0 {
		return
	}
	b.writeLn(0, "servers:")
	for _, srv := range b.base.swagger.servers {
		b.writeLn(1, "- url: %s", srv.url)
		b.writeLn(1, "  description: %s", srv.desc)
	}
}

type pathBuilder struct {
	base *baseBuilder
}

func (b *pathBuilder) writeLn(indent int, format string, i ...interface{}) {
	b.base.writeLn(indent, format, i...)
}

func (b *pathBuilder) writePaths() {
	b.writeLn(0, "paths:")

	contentType := b.base.swagger.DefaultContentType

	lastPath := Path("")
	lastMethod := Method("")
	for _, op := range b.sortedOperations() {
		// Only write the path and method when they change from operation to operation.
		if lastPath != op.Path {
			b.writeLn(1, "%s:", op.Path)
			b.writeLn(2, "%s:", op.Method)
			lastPath = op.Path
			lastMethod = op.Method
		} else if lastMethod != op.Method {
			b.writeLn(2, "%s:", op.Method)
			lastMethod = op.Method
		}

		if len(op.Tags) > 0 {
			if len(op.Tags) > 0 {
				b.writeLn(3, `tags: ["%s"]`, strings.Join(op.Tags, `", "`))
			}
		}

		b.writeLn(3, "operationId: %s", op.OperationID)
		b.base.writeNotEmpty(3, "summary: %s", op.Summary)
		b.base.writeNotEmpty(3, "description: %s", op.Description)

		if !op.Params.Nil() {
			b.writeParams(3, op.Params)
		}
		if op.useRequestBody() {
			b.writeLn(3, "requestBody:")
			b.writeLn(4, "required: true")
			b.writeLn(4, "content:")
			b.writeLn(5, "%s:", contentType)
			b.writeLn(6, "schema:")
			b.base.writeStructSchema(7, op.Params, func(f Field) bool {
				// We *always* want to recurse/expand request body struct fields that are structs/slices,
				// unless they are being terminated into a data type.
				return !b.base.swagger.isMappedToDataType(f)
			})
		}
		b.writeLn(3, "responses:")
		for _, resp := range op.Responses {
			b.writeLn(4, "'%s':", resp.Code)
			b.writeLn(5, "description: %s", resp.Description)
			if !resp.Field.Nil() {
				b.writeLn(5, "content:")
				switch resp.Field.Kind {
				case reflect.String:
					b.writeLn(6, "text/plain:")
				default:
					b.writeLn(6, "%s:", contentType)
				}
				b.writeLn(7, "schema:")
				b.base.writeRefSchema(8, resp.Field)
			}
		}
	}
}

func (b *pathBuilder) writeParams(indent int, f Field) {
	writeParams := b.base.writeOnce(indent, "parameters:")
	for _, field := range enumerateStructFields(f) {
		tag := field.StructField.Tag
		var name, in string

		if path := tag.Get("path"); path != "" {
			name = path
			in = "path"
		} else if query := tag.Get("query"); query != "" {
			name = query
			in = "query"
		} else if header := tag.Get("header"); header != "" {
			name = header
			in = "header"
		} else {
			continue
		}
		writeParams()
		b.writeLn(indent+1, "- name: %s", name)
		b.writeLn(indent+1, "  in: %s", in)
		if in == "path" {
			b.writeLn(indent+1, "  required: true")
		}
		b.base.writeNotEmpty(indent+1, "  description: %s", tag.Get("description"))
		b.writeLn(indent+1, "  schema:")
		b.base.writeRefSchema(indent+3, field)
	}
}

// Return a slice of operations such that they are sorted by path and then by method.
// So ops of {/xyz POST, /abc GET, /xyz GET, /abc POST}
// will sort to {/abc GET, /abc POST, /xyz GET, /xyz POST}.
func (b *pathBuilder) sortedOperations() []internalOperation {
	ops := make([]internalOperation, 0, len(b.base.swagger.operations))
	ops = append(ops, b.base.swagger.operations...)

	sort.Slice(ops, func(i, j int) bool {
		oi := ops[i]
		oj := ops[j]
		if oi.Path == oj.Path {
			wi := methodWeights[oi.Method]
			wj := methodWeights[oj.Method]
			return wi < wj
		}
		return oi.Path < oj.Path
	})
	return ops
}

var methodWeights = map[Method]int{
	"get":    1,
	"post":   2,
	"put":    3,
	"patch":  4,
	"delete": 5,
}

type componentsBuilder struct {
	base *baseBuilder
}

func (b *componentsBuilder) writeComponents() {
	writeComponents := b.base.writeOnce(0, "components:")

	sortedSchemas := b.sortedFieldsForSchema()
	if len(sortedSchemas) > 0 {
		writeComponents()
		b.writeSchemas(sortedSchemas)
	}

	if len(b.base.swagger.securities) > 0 {
		writeComponents()
		b.writeSecuritySchemas()
		b.writeSecurityScopes()
	}
}

func (b *componentsBuilder) writeSchemas(sortedSchemas Fields) {
	b.base.writeLn(1, "schemas:")
	for _, tv := range sortedSchemas {
		b.base.writeLn(2, "%s:", tv.Type.Name())
		b.base.writeStructSchema(3, tv, b.shouldRecurseStructField)
	}
}

// A type will end up in the schema if it has a name and is exported.
// Inline types (no name) amd embedded structs (Anonymous) should be traversed.
// Assume lowercase named isn't meant for the swagger doc.
func (b *componentsBuilder) shouldRecurseStructField(f Field) bool {
	if f.Type.Name() == "" {
		return true
	}
	if f.StructField.Anonymous {
		return true
	}
	return !isExportedName(f.Type.Name())
}

// Each struct type should be in the map only once,
// and in alphabetical order.
func (b *componentsBuilder) sortedFieldsForSchema() Fields {
	allFields := make(Fields, 0, len(b.base.swagger.operations))
	visitor := func(f Field) {
		allFields = append(allFields, f)
	}
	for _, op := range b.base.swagger.operations {
		for _, resp := range op.Responses {
			b.visitStructs(resp.Field, visitor)
		}
	}
	relevantSortedFields := allFields.
		Compact().
		FlattenSliceTypes().
		Distinct().
		RemoveAnonymousTypes()
	sort.Sort(relevantSortedFields)
	return relevantSortedFields
}

func (b *componentsBuilder) visitStructs(f Field, visitor func(Field)) {
	if f.Kind == reflect.Slice {
		f = ZeroSliceValueField(f.Type)
	}
	if mappedType, found := b.base.swagger.dataTypeDefFor(f); found {
		f = mappedType.Field
	}

	if f.Kind != reflect.Struct {
		return
	}

	visitor(f)
	for _, fieldTVP := range enumerateStructFields(f) {
		b.visitStructs(fieldTVP, visitor)
	}
}

func (b *componentsBuilder) writeSecuritySchemas() {
	b.base.writeLn(1, "securitySchemes:")
	for _, sec := range b.base.swagger.securities {
		b.base.writeLn(2, "%s:", sec.ID())
		for _, tuple := range sec.Fields().Sorted() {
			b.base.writeLn(3, "%s: %s", tuple[0], tuple[1])
		}
	}
}

func (b *componentsBuilder) writeSecurityScopes() {
	b.base.writeLn(0, "security:")
	for _, sec := range b.base.swagger.securities {
		b.base.writeLn(1, "- %s: []", sec.ID())
	}
}
