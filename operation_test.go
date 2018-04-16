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
	fmt.Println(sashay.NewPath("/users/:id/pets"))
	fmt.Println(sashay.NewPath("/users/{id}"))
	fmt.Println(sashay.NewPath("/users/{id}/pets"))
	// Output:
	// /users/{id}
	// /users/{id}/pets
	// /users/{id}
	// /users/{id}/pets
}

func ExampleNewOperationID() {
	op := sashay.NewOperation("GET", "/users/:id", "", nil, nil, nil)
	fmt.Println(sashay.NewOperationID(op))
	// Output:
	// getUsersId
}
