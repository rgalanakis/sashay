package sashay

import "reflect"

// Field is a container for reflection information about a value.
// Since we need this repeatedly, we parse it once and pass it around.
type Field struct {
	// Interface is the original passed-in field value.
	Interface interface{}
	// Type is reflect.TypeOf(field.Interface)
	Type reflect.Type
	// Kind is reflect.TypeOf(field.Interface).Kind()
	Kind reflect.Kind
	// Value is reflect.ValueOf(field.Interface)
	Value reflect.Value
	// DataTyper is the mapping function for this value and type.
	// It may be nil.
	// In general it is only set for a single Field instance for a type,
	// when defined through Sashay#DefineDataType.
	DataTyper DataTyper
	// StructField is the StructField the Field was created from.
	// If it was not created from a field, FromStructField will be false.
	StructField     reflect.StructField
	FromStructField bool
}

// NewField returns a Field initialized from v.
// If fields is provided, its first item indicates the Field was parsed from a StructField.
func NewField(v interface{}, fields ...reflect.StructField) Field {
	if v == nil {
		return Field{}
	}
	var structField *reflect.StructField
	if len(fields) > 0 {
		structField = &fields[0]
	}
	return newField(v, true, structField)
}

func newField(v interface{}, deference bool, field *reflect.StructField) Field {
	t := reflect.TypeOf(v)
	k := t.Kind()
	if deference && k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}
	result := Field{
		Interface: v,
		Type:      t,
		Kind:      k,
		Value:     reflect.ValueOf(v),
	}
	if field != nil {
		result.StructField = *field
		result.FromStructField = true
	}
	return result
}

// Return true if f was created from nil.
func (f Field) Nil() bool {
	return f.Interface == nil
}

func (f Field) String() string {
	if f.Nil() {
		return "Field{}"
	}
	return "Field{kind: " + f.Kind.String() + ", type:" + f.Type.Name() + "}"
}

// For a reflect.Type for a slice, return a Field representing an item of the slice's underlying type.
// So ZeroSliceValueField(reflect.TypeOf([]MyType{}) would be the same as NewField(MyType{}).
func ZeroSliceValueField(t reflect.Type) Field {
	sliceVal := reflect.MakeSlice(t, 1, 1)
	r := sliceVal.Index(0)
	return NewField(r.Interface())
}

// Fields is a slice of Field instances.
type Fields []Field

func (fs Fields) Len() int {
	return len(fs)
}

func (fs Fields) Less(i, j int) bool {
	return fs[i].Type.Name() < fs[j].Type.Name()
}

func (fs Fields) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

// Compact returns a new Fields with Nil values removed.
func (fs Fields) Compact() Fields {
	res := make(Fields, 0, len(fs))
	for _, p := range fs {
		if !p.Nil() {
			res = append(res, p)
		}
	}
	return res
}

// FlattenSliceTypes replaces Fields with slice types with their underlying value
// (see ZeroSliceTypeField).
func (fs Fields) FlattenSliceTypes() Fields {
	res := make(Fields, 0, len(fs))
	for _, f := range fs {
		if f.Type.Kind() == reflect.Slice {
			res = append(res, ZeroSliceValueField(f.Type))
		} else {
			res = append(res, f)
		}
	}
	return res
}

// Distinct eliminates Fields with the same Type.
func (fs Fields) Distinct() Fields {
	res := make(Fields, 0, len(fs))
	seen := make(map[reflect.Type]bool, len(fs))
	for _, p := range fs {
		if found := seen[p.Type]; !found {
			seen[p.Type] = true
			res = append(res, p)
		}
	}
	return res
}

// RemoveAnonymousTypes removes Fields that have no PkgPath, such as anonymous types.
func (fs Fields) RemoveAnonymousTypes() Fields {
	res := make(Fields, 0, len(fs))
	for _, f := range fs {
		if f.Type.PkgPath() != "" {
			res = append(res, f)
		}
	}
	return res
}
