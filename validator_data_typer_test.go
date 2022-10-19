package sashay_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rgalanakis/sashay"
	"reflect"
	"strings"
)

// RegisterValidatorDataTypes registers override DataTypers for all builtin data types.
func RegisterValidatorDataTypes(sa *sashay.Sashay) {
	for _, v := range sashay.BuiltinDataTypeValues {
		sa.DefineDataType(v, sashay.BuiltinDataTyperFor(v, ParseValidations))
	}
}

func ParseValidations(field sashay.Field, of sashay.ObjectFields) {
	validations := strings.Split(field.StructField.Tag.Get("validate"), ",")
	for _, v := range validations {
		parts := strings.Split(v, "=")
		switch parts[0] {
		case "len":
			of["minLength"] = parts[1]
			of["maxLength"] = parts[1]
		case "min":
			if field.Kind == reflect.String {
				of["minLength"] = parts[1]
			} else {
				of["min"] = parts[1]
			}
		case "max":
			if field.Kind == reflect.String {
				of["maxLength"] = parts[1]
			} else {
				of["max"] = parts[1]
			}
		case "regexp":
			of["pattern"] = parts[1]
		case "nonzero":
			of["required"] = "true"
		}
	}
}

var _ = Describe("ValidatorDataTyper", func() {
	var (
		sa *sashay.Sashay
	)

	BeforeEach(func() {
		sa = sashay.New(
			"SwaggerGenAPI",
			"Demonstrate auto-generating Swagger",
			"0.1.9",
		)
		RegisterValidatorDataTypes(sa)
	})

	It("parses go-validator tags", func() {
		sa.Add(sashay.NewOperation(
			"GET",
			"/empty",
			"",
			struct {
				MinmaxStr string  `query:"minmaxstr" validate:"min=1,max=5"`
				MinmaxNum float32 `query:"minmaxnum" validate:"min=1,max=5.5"`
				Regexp    string  `query:"regexp" validate:"regexp=.*[wy](i|o)bble$"`
				Len       string  `query:"len" validate:"len=4"`
				Nonzero   string  `query:"nonzero" validate:"nonzero"`
			}{},
			nil,
			nil,
		))
		yaml := sa.BuildYAML()
		Expect(yaml).To(ContainSubstring(`paths:
  /empty:
    get:
      operationId: getEmpty
      parameters:
        - name: minmaxstr
          in: query
          schema:
            type: string
            maxLength: 5
            minLength: 1
        - name: minmaxnum
          in: query
          schema:
            type: number
            format: float
            max: 5.5
            min: 1
        - name: regexp
          in: query
          schema:
            type: string
            pattern: .*[wy](i|o)bble$
        - name: len
          in: query
          schema:
            type: string
            maxLength: 4
            minLength: 4
        - name: nonzero
          in: query
          schema:
            type: string
            required: true
`))
	})
})
