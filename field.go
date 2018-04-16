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
// If field is provided, indicate the Field was parsed from a StructField.
func NewField(v interface{}, field ...reflect.StructField) Field {
	if v == nil {
		return Field{}
	}
	f := reflect.StructField{}
	hasStructField := false
	if len(field) > 0 {
		f = field[0]
		hasStructField = true
	}
	t := reflect.TypeOf(v)
	return Field{
		v,
		t,
		t.Kind(),
		reflect.ValueOf(v),
		nil,
		f,
		hasStructField,
	}
}

// Return true if f was created from nil.
func (f Field) Nil() bool {
	return f.Interface == nil
}

func (f Field) String() string {
	if f.Nil() {
		return "Field{}"
	}
	return "Field{" + f.Kind.String() + "-" + f.Type.Name() + "}"
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

func (t Fields) Len() int {
	return len(t)
}

func (t Fields) Less(i, j int) bool {
	return t[i].Type.Name() < t[j].Type.Name()
}

func (t Fields) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// Compact returns a new Fields with Nil values removed.
func (t Fields) Compact() Fields {
	res := make(Fields, 0, len(t))
	for _, p := range t {
		if !p.Nil() {
			res = append(res, p)
		}
	}
	return res
}

// FlattenSliceTypes replaces Fields with slice types with their underlying value
// (see ZeroSliceTypeField).
func (t Fields) FlattenSliceTypes() Fields {
	res := make(Fields, 0, len(t))
	for _, tvp := range t {
		if tvp.Type.Kind() == reflect.Slice {
			res = append(res, ZeroSliceValueField(tvp.Type))
		} else {
			res = append(res, tvp)
		}
	}
	return res
}

// Distinct eliminates Fields with the same Type.
func (t Fields) Distinct() Fields {
	res := make(Fields, 0, len(t))
	seen := make(map[reflect.Type]bool, len(t))
	for _, p := range t {
		if found := seen[p.Type]; !found {
			seen[p.Type] = true
			res = append(res, p)
		}
	}
	return res
}

// RemoveAnonymousTypes removes Fields that have no PkgPath, such as anonymous types.
func (t Fields) RemoveAnonymousTypes() Fields {
	res := make(Fields, 0, len(t))
	for _, tvp := range t {
		if tvp.Type.PkgPath() != "" {
			res = append(res, tvp)
		}
	}
	return res
}
