package sashay_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rgalanakis/sashay"
	"reflect"
)

func ExampleZeroSliceValueField() {
	type User struct{}

	slice := make([]User, 0)
	sliceField := sashay.NewField(slice)
	userField := sashay.NewField(User{})
	zeroSliceField := sashay.ZeroSliceValueField(reflect.TypeOf(slice))
	fmt.Println("sliceField type name:", sliceField.Type)
	fmt.Println("userField type name:", userField.Type)
	fmt.Println("zeroSliceField type name:", zeroSliceField.Type)
	// Output:
	// sliceField type name: []sashay_test.User
	// userField type name: sashay_test.User
	// zeroSliceField type name: sashay_test.User
}

var _ = Describe("Field", func() {
	It("can render itself as a string", func() {
		f := sashay.NewField(5)
		Expect(f.String()).To(Equal("Field{kind: int, type:int}"))
	})
})

var _ = Describe("Fields", func() {
	Describe("FlattenSliceTypes", func() {
		It("replaces Fields with slice types with their underlying value", func() {
			intField := sashay.NewField(5)
			strSliceField := sashay.NewField([]string{})
			flattened := sashay.Fields{intField, strSliceField}.FlattenSliceTypes()
			Expect(flattened[0].Kind.String()).To(Equal(reflect.Int.String()))
			Expect(flattened[1].Kind.String()).To(Equal(reflect.String.String()))
		})
	})
})
