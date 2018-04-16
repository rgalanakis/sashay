package sashay

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// Sashay describes an OpenAPI document,
// including meta info, servers, security, schemas, and all Operations.
// See https://swagger.io/specification/
type Sashay struct {
	// The default content type for all request bodies and responses.
	// Defaults to application/json. This can only be set document-wide,
	// and cannot vary per-endpoint right now.
	DefaultContentType                    string
	title, desc, version                  string
	operations                            []internalOperation
	servers                               []swaggerServer
	securities                            []swaggerSecurity
	tos                                   string
	contactName, contactURL, contactEmail string
	licenseName, licenseURL               string
	tags                                  []swaggerTag
	dataTypesForTypes                     map[reflect.Type]DataTypeDef
}

// New returns a pointer to a new Sashay instance,
// initialized with the provided values.
// See https://swagger.io/specification/#infoObject
func New(title, description, version string) *Sashay {
	sw := &Sashay{
		DefaultContentType: "application/json",
		title:              title,
		desc:               description,
		version:            version,
		operations:         make([]internalOperation, 0),
		servers:            make([]swaggerServer, 0),
		securities:         make([]swaggerSecurity, 0),
		dataTypesForTypes:  make(map[reflect.Type]DataTypeDef),
	}

	defaultDT := DefaultDataTyper()
	sw.DefineDataType(int(0), ChainDataTyper(SimpleDataTyper("integer", "int64"), defaultDT))
	sw.DefineDataType(int64(0), ChainDataTyper(SimpleDataTyper("integer", "int64"), defaultDT))
	sw.DefineDataType(int32(0), ChainDataTyper(SimpleDataTyper("integer", "int32"), defaultDT))
	sw.DefineDataType("", ChainDataTyper(SimpleDataTyper("string", ""), defaultDT))
	sw.DefineDataType(false, ChainDataTyper(SimpleDataTyper("boolean", ""), defaultDT))
	sw.DefineDataType(float64(0), ChainDataTyper(SimpleDataTyper("number", "double"), defaultDT))
	sw.DefineDataType(float32(0), ChainDataTyper(SimpleDataTyper("number", "float"), defaultDT))
	sw.DefineDataType(time.Time{}, ChainDataTyper(SimpleDataTyper("string", "date-time"), defaultDT))
	return sw
}

// Add registers a Swagger operations and all the associated types.
func (sa *Sashay) Add(op Operation) Operation {
	sa.operations = append(sa.operations, op.toInternalOperation())
	return op
}

// AddServer adds a server to the swagger file.
// See https://swagger.io/specification/#serverObject
func (sa *Sashay) AddServer(url, description string) *Sashay {
	sa.servers = append(sa.servers, swaggerServer{url, description})
	return sa
}

type swaggerServer struct {
	url, desc string
}

// SetTermsOfService sets the termsOfService in the swagger file info.
// See https://swagger.io/specification/#infoObject
func (sa *Sashay) SetTermsOfService(url string) *Sashay {
	sa.tos = url
	return sa
}

// SetContact sets the contact fields in the swagger file info.
// See https://swagger.io/specification/#contactObject
func (sa *Sashay) SetContact(name, url, email string) *Sashay {
	sa.contactName = name
	sa.contactURL = url
	sa.contactEmail = email
	return sa
}

// SetLicense sets the license fields in the swagger file info.
// See https://swagger.io/specification/#licenseObject
func (sa *Sashay) SetLicense(name, url string) *Sashay {
	sa.licenseName = name
	sa.licenseURL = url
	return sa
}

func (sa *Sashay) AddTag(name, desc string) *Sashay {
	sa.tags = append(sa.tags, swaggerTag{name: name, desc: desc})
	return sa
}

type swaggerTag struct {
	name, desc string
}

// AddBasicAuthSecurity adds type:http scheme:basic security schema and global scope.
// See https://swagger.io/specification/#securitySchemeObject
// https://swagger.io/docs/specification/authentication/basic-authentication/
func (sa *Sashay) AddBasicAuthSecurity() *Sashay {
	sec := swaggerSecurity{"id": "basicAuth", "type": "http", "scheme": "basic"}
	sa.securities = append(sa.securities, sec)
	return sa
}

// AddJWTSecurity adds type:http scheme:bearer security schema and global scope.
// See https://swagger.io/specification/#securitySchemeObject
// https://swagger.io/docs/specification/authentication/bearer-authentication/
func (sa *Sashay) AddJWTSecurity() *Sashay {
	sec := swaggerSecurity{"id": "bearerAuth", "type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
	sa.securities = append(sa.securities, sec)
	return sa
}

// AddAPIKeySecurity adds type:apiKey security schema and global scope.
// See https://swagger.io/specification/#securitySchemeObject
// https://swagger.io/docs/specification/authentication/api-keys/
func (sa *Sashay) AddAPIKeySecurity(in, name string) *Sashay {
	sec := swaggerSecurity{"id": "apiKeyAuth", "type": "apiKey", "in": in, "name": name}
	sa.securities = append(sa.securities, sec)
	return sa
}

type swaggerSecurity ObjectFields

func (ss swaggerSecurity) ID() string {
	return ss["id"]
}

func (ss swaggerSecurity) Fields() ObjectFields {
	of := make(ObjectFields)
	for k, v := range ss {
		if k != "id" {
			of[k] = v
		}
	}
	return of
}

// DefineDataType defines the DataTyper to use for values with the same type as i.
// For example, DefineDataType(int(0), SimpleDataTyper("integer", "int64")) means that
// whenever the options for a boolean field are written out,
// it will get the properties {type: "integer", format: "int64"}.
//
// Normally Go structs are not data types- they are either walked (parameter objects)
// or receive schemas (response objects).
// However, some structs, like time.Time, should be represented as data types.
// To achieve this, the DataTyper for time.Time is defined as:
//
//     sw.DefineDataType(time.Time{}, SimpleDataTyper("string", "date-time"))
//
// So whenever a time.Time value is seen, the fields {type: "string", format:"date-time"} are used.
//
// Callers can use DefineDataType(myStruct{}, provide define their own DataTyper for structs that they.
// They can use SimpleDataTyper, or provide a function with dynamic logic for what fields to add:
//
//     sw.DefineDataType(FormattableString{}, func(tvp Field) ObjectFields {
//       res := ObjectFields{"type": "string"}
//       if val, ok := tvp.StructField.Tag.Lookup("format"); ok {
//         res["format"] = val
//       }
//		 return res
//     })
//
// The DataTyper above will be called for any struct field with a type of FormattableString,
// and use a value for the "format" field based on the struct field's tag.
//
// See https://swagger.io/specification/#dataTypes
func (sa *Sashay) DefineDataType(i interface{}, dt DataTyper) {
	tvp := NewField(i)
	sa.dataTypesForTypes[tvp.Type] = DataTypeDef{tvp, dt}
}

func (sa *Sashay) WriteYAML(buf io.Writer) error {
	bb := &baseBuilder{buf, sa}
	db := docBuilder{bb}
	db.writeInfo()
	db.writeTags()
	db.writeServers()
	pb := pathBuilder{bb}
	pb.writePaths()
	cp := componentsBuilder{bb}
	cp.writeComponents()
	return nil
}

// BuildYAML returns the YAML Swagger string for the receiver.
func (sa *Sashay) BuildYAML() string {
	buf := bytes.NewBuffer(nil)
	sa.WriteYAML(buf)
	return buf.String()
}

func (sa *Sashay) dataTypeDefFor(tvp Field) (DataTypeDef, bool) {
	dtd, ok := sa.dataTypesForTypes[tvp.Type]
	return dtd, ok
}

// Return true if a Go struct type is mapped to a data type (like time.Time is mapped to string).
func (sa *Sashay) isMappedToDataType(tvp Field) bool {
	_, found := sa.dataTypesForTypes[tvp.Type]
	return found
}

// A type will end up in the schema if it has a name and is exported.
// No name is anonymous, so much be traversed. Assume lowercase named isn't meant for the swagger doc.
func isTypeForSchema(t reflect.Type) bool {
	if t.Name() == "" {
		return false
	}
	return isExportedName(t.Name())
}

// Assuming tvp.Type is a struct, enumerate all the exported fields as Fields.
func enumerateStructFields(tvp Field) Fields {
	result := make(Fields, 0)
	for i := 0; i < tvp.Type.NumField(); i++ {
		fieldDef := tvp.Type.Field(i)
		if !isExportedField(fieldDef) {
			continue
		}
		getterField := tvp.Value.FieldByName(fieldDef.Name)
		if !getterField.CanInterface() {
			panic(fmt.Sprintf("Cannot get value of unexported field: %s %s", tvp.Type.Name(), fieldDef.Name))
		}
		val := getterField.Interface()
		tvp := NewField(val, fieldDef)
		result = append(result, tvp)
	}
	return result
}

// Return true if f is exported.
// Anonymous structs and those with no name are exported/meant for Swagger.
func isExportedField(f reflect.StructField) bool {
	if f.Name == "" || f.Anonymous {
		return true
	}
	return isExportedName(f.Name)
}

// Return true if s is exported (leading char of type name is uppercase).
// User => true, user => false, mypkg.User => true
// The empty string is ambiguous and this method will panic if called with it.
func isExportedName(s string) bool {
	if s == "" {
		panic("isExportedName cannot be used with an empty string, it is ambiguous. Front it with something else.")
	}
	parts := strings.Split(s, ".")
	typename := parts[len(parts)-1]
	c := typename[0]
	return c >= 65 && c <= 90
}

// Return the link for a $ref field, like "#/components/schemas/User".
func schemaRefLink(tvp Field) string {
	return fmt.Sprintf("#/components/schemas/%s", tvp.Type.Name())
}
