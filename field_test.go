package sashay_test

import (
	"fmt"
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
