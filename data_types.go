package sashay

import (
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

// DataTyper returns the ObjectFields for a Field (which should represent a data type, not a deep/struct type).
type DataTyper func(tvp Field) ObjectFields

// SimpleDataTyper returns a DataTyper that will specify a "type" field of type,
// and a "format" field of format, if not empty.
func SimpleDataTyper(typ, format string) DataTyper {
	return func(tvp Field) ObjectFields {
		f := ObjectFields{"type": typ}
		if format != "" {
			f["format"] = format
		}
		return f
	}
}

// DefaultDataTyper returns a DataTyper that sets the "default" field of the data type
// to the "default" value of the struct tag on the Field passed to it.
func DefaultDataTyper() DataTyper {
	return func(tvp Field) ObjectFields {
		f := ObjectFields{}
		if d := tvp.StructField.Tag.Get("default"); d != "" {
			f["default"] = d
		}
		return f
	}
}

// ChainDataTyper returns a DataTyper that will call each function in typers in order,
// merging all the returned ObjectFields.
// In conflict, later typers will take priority.
func ChainDataTyper(typers ...DataTyper) DataTyper {
	return func(tvp Field) ObjectFields {
		result := ObjectFields{}
		for _, t := range typers {
			for k, v := range t(tvp) {
				result[k] = v
			}
		}
		return result
	}
}

var defaultDataTyper = DefaultDataTyper()

func noopDataTyper(_ Field) ObjectFields {
	return ObjectFields{}
}

// BuiltinDataTyperFor returns the default/builtin DataTyper for type of value.
// The default data typers are always SimpleDataTyper with the right type and format fields.
// If value is an unsupported type, return only the DefaultDataTyper.
func BuiltinDataTyperFor(value interface{}) DataTyper {
	dt := noopDataTyper
	switch value.(type) {
	case int, int64:
		dt = SimpleDataTyper("integer", "int64")
	case int32:
		dt = SimpleDataTyper("integer", "int32")
	case string:
		dt = SimpleDataTyper("string", "")
	case bool:
		dt = SimpleDataTyper("boolean", "")
	case float64:
		dt = SimpleDataTyper("number", "double")
	case float32:
		dt = SimpleDataTyper("number", "float")
	case time.Time:
		dt = SimpleDataTyper("string", "date-time")
	}
	return ChainDataTyper(dt, defaultDataTyper)
}
