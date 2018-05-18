package sashay_test

import (
	"fmt"
	"github.com/rgalanakis/sashay"
)

func Example_readme() {
	type Pet struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}
	type Error struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	}

	sw := sashay.New("PetStore API", "Manage your pet store with our API", "1.0.0")
	sw.Add(sashay.NewOperation(
		"GET",
		"/pets",
		"Return all pets.",
		struct {
			Status string `query:"status"`
		}{},
		[]Pet{},
		Error{},
	))
	sw.Add(sashay.NewOperation(
		"POST",
		"/pets",
		"Create a pet.",
		struct {
			Pretty bool   `query:"pretty" default:"true" description:"If true, return pretty-printed JSON."`
			Name   string `json:"name"`
		}{},
		Pet{},
		Error{},
	))
	sw.Add(sashay.NewOperation(
		"GET",
		"/pets/:id",
		"Fetch info about a pet.",
		struct {
			ID   int    `path:"id"`
			Name string `json:"name"`
		}{},
		Pet{},
		Error{},
	))
	fmt.Println(sw.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: PetStore API
	//   description: Manage your pet store with our API
	//   version: 1.0.0
	// paths:
	//   /pets:
	//     get:
	//       operationId: getPets
	//       summary: Return all pets.
	//       parameters:
	//         - name: status
	//           in: query
	//           schema:
	//             type: string
	//       responses:
	//         '200':
	//           description: ok response
	//           content:
	//             application/json:
	//               schema:
	//                 type: array
	//                 items:
	//                   $ref: '#/components/schemas/Pet'
	//         'default':
	//           description: error response
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Error'
	//     post:
	//       operationId: postPets
	//       summary: Create a pet.
	//       parameters:
	//         - name: pretty
	//           in: query
	//           description: If true, return pretty-printed JSON.
	//           schema:
	//             type: boolean
	//             default: true
	//       requestBody:
	//         required: true
	//         content:
	//           application/json:
	//             schema:
	//               type: object
	//               properties:
	//                 name:
	//                   type: string
	//       responses:
	//         '201':
	//           description: ok response
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Pet'
	//         'default':
	//           description: error response
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Error'
	//   /pets/{id}:
	//     get:
	//       operationId: getPetsId
	//       summary: Fetch info about a pet.
	//       parameters:
	//         - name: id
	//           in: path
	//           required: true
	//           schema:
	//             type: integer
	//             format: int64
	//       responses:
	//         '200':
	//           description: ok response
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Pet'
	//         'default':
	//           description: error response
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/Error'
	// components:
	//   schemas:
	//     Error:
	//       type: object
	//       properties:
	//         code:
	//           type: integer
	//           format: int32
	//         message:
	//           type: string
	//     Pet:
	//       type: object
	//       properties:
	//         id:
	//           type: integer
	//           format: int64
	//         name:
	//           type: string
	//         tag:
	//           type: string
}
