package sashay

import (
	"sort"
)

// DataTypeDef associates a Field with the DataTyper for that field.
type DataTypeDef struct {
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

func SimpleDataTyper(typ, format string) DataTyper {
	return func(tvp Field) ObjectFields {
		f := ObjectFields{"type": typ}
		if format != "" {
			f["format"] = format
		}
		return f
	}
}

func DefaultDataTyper() DataTyper {
	return func(tvp Field) ObjectFields {
		f := ObjectFields{}
		if d := tvp.StructField.Tag.Get("default"); d != "" {
			f["default"] = d
		}
		return f
	}
}
