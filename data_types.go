package sashay

import (
	"reflect"
	"sort"
	"time"
)

// dataTypeDef associates a Field with the DataTyper for that field.
type dataTypeDef struct {
	Field     Field
	DataTyper DataTyper
}

// ObjectFields is a mapping of the fields for some object,
// such as a data type (https://swagger.io/specification/#dataTypes)
// or security object (https://swagger.io/specification/#securitySchemeObject).
// In general, this always includes a "type" key, and other fields are based on the object being represented
// (data type fields often have a "format", security objects a "scheme").
type ObjectFields map[string]string

// Sorted returns a slice of string tuples suitable for writing to YAML.
// "type" should always be first, otherwise sort keys alphabetically.
func (dtf ObjectFields) Sorted() [][]string {
	result := make([][]string, 0)
	for k, v := range dtf {
		result = append(result, []string{k, v})
	}
	sort.Slice(result, func(i, j int) bool {
		keyI := result[i][0]
		keyJ := result[j][0]
		if keyI == "type" {
			return true
		}
		if keyJ == "type" {
			return false
		}
		return keyI < keyJ
	})
	return result
}

// DataTyper modifies the ObjectFields for the passed field Field.
// The ObjectFields are written into the schema for a data type.
type DataTyper func(f Field, of ObjectFields)

// SimpleDataTyper returns a DataTyper that will specify a "type" field of type,
// and a "format" field of format, if not empty.
func SimpleDataTyper(swaggerType, format string) DataTyper {
	return func(f Field, of ObjectFields) {
		of["type"] = swaggerType
		if format != "" {
			of["format"] = format
		}
		if f.Kind == reflect.Ptr {
			of["nullable"] = "true"
		}
	}
}

// DefaultDataTyper returns a DataTyper that sets the "default" field of the data type
// to the "default" value of the struct tag on the Field passed to it.
func DefaultDataTyper() DataTyper {
	return func(f Field, of ObjectFields) {
		if d := f.StructField.Tag.Get("default"); d != "" {
			of["default"] = d
		}
	}
}

// ChainDataTyper returns a DataTyper that will call each function in typers in order,
// merging all the returned ObjectFields.
// In conflict, later typers will take priority.
func ChainDataTyper(typers ...DataTyper) DataTyper {
	return func(f Field, of ObjectFields) {
		for _, t := range typers {
			t(f, of)
		}
	}
}

var defaultDataTyper = DefaultDataTyper()

func noopDataTyper(_ Field, _ ObjectFields) {}

// BuiltinDataTyperFor returns the default/builtin DataTyper for type of value.
// The default data typers are always SimpleDataTyper with the right type and format fields.
// If value is an unsupported type, return only the DefaultDataTyper.
func BuiltinDataTyperFor(value interface{}, chained ...DataTyper) DataTyper {
	dt := noopDataTyper
	switch value.(type) {
	case int, int64, *int, *int64:
		dt = SimpleDataTyper("integer", "int64")
	case int32, *int32:
		dt = SimpleDataTyper("integer", "int32")
	case string, *string:
		dt = SimpleDataTyper("string", "")
	case bool, *bool:
		dt = SimpleDataTyper("boolean", "")
	case float64, *float64:
		dt = SimpleDataTyper("number", "double")
	case float32, *float32:
		dt = SimpleDataTyper("number", "float")
	case time.Time, *time.Time:
		dt = SimpleDataTyper("string", "date-time")
	}
	typers := []DataTyper{dt, defaultDataTyper}
	typers = append(typers, chained...)
	return ChainDataTyper(typers...)
}
