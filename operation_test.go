package sashay_test

import (
	"fmt"
	"github.com/rgalanakis/sashay"
)

func ExampleNewMethod() {
	fmt.Println(sashay.NewMethod("GET"))
	fmt.Println(sashay.NewMethod("get"))
	// Output:
	// get
	// get
}

func ExampleNewPath() {
	fmt.Println(sashay.NewPath("/users/:id"))
	// Output:
	// /users/{id}
}

func ExampleNewOperationID() {
	op := sashay.NewOperation("GET", "/users/:id", "", nil, nil, nil)
	fmt.Println(sashay.NewOperationID(op))
	// Output:
	// getUsersId
}
