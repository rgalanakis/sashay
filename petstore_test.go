package sashay_test

import (
	"fmt"
	"github.com/rgalanakis/sashay"
	"io/ioutil"
	"net/http"
	"os"
)

// Stand-in for whatever HTTP framework you are using
type FrameworkRouter struct{}

func (FrameworkRouter) ServeHTTP(http.ResponseWriter, *http.Request)     {}
func (FrameworkRouter) AddRoute(method, path string, h http.HandlerFunc) {}

// main.go
func PetStoreSwagger() *sashay.Sashay {
	return sashay.New(
		"Swagger Petstore",
		"A sample API that uses a petstore as an example to demonstrate features in the OpenAPI 3.0 specification",
		"1.0.0").
		AddAPIKeySecurity("header", "api_key").
		SetTermsOfService("http://swagger.io/terms/").
		SetContact("", "", "apiteam@swagger.io").
		SetLicense("Apache 2.0", "http://www.apache.org/licenses/LICENSE-2.0.html").
		AddServer("http://petstore.swagger.io/api", "Public API server").
		AddTag("pet", "Everything about your Pets").
		AddTag("store", "Access to Petstore orders").
		AddTag("user", "Operations about user")
}

// noinspection GoUnusedExportedFunction
func StartServer() {
	sw := PetStoreSwagger()
	router := &FrameworkRouter{}
	RegisterRoutes(router, sw)
	if len(os.Args) > 0 && os.Args[0] == "-swagger" {
		yaml := sw.BuildYAML()
		ioutil.WriteFile("swagger.yml", []byte(yaml), 0644)
		os.Exit(0)
	}
	http.ListenAndServe(":8080", router)
}

// handlers.go file
// noinspection GoUnusedParameter
func GetPets(http.ResponseWriter, *http.Request) {
	// Your code here
}

var CreatePet http.HandlerFunc
var GetPet http.HandlerFunc
var DeletePet http.HandlerFunc

// models.go file
type Pet struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// routes.go file
type Route struct {
	operation sashay.Operation
	handler   http.HandlerFunc
}

func RegisterRoutes(router *FrameworkRouter, sw *sashay.Sashay) {
	for _, route := range routes {
		sw.Add(route.operation)
		router.AddRoute(route.operation.Method, route.operation.Path, route.handler)
	}
}

var routes = []Route{
	{
		sashay.NewOperation(
			"GET",
			"/pets",
			"Returns all pets from the system that the user has access to",
			struct {
				Tags  []string `query:"tags" description:"tags to filter by"`
				Limit int32    `query:"limit" description:"maximum number of results to return"`
			}{},
			[]Pet{},
			Error{},
		), GetPets,
	}, {
		sashay.NewOperation(
			"POST",
			"/pets",
			"Creates a new pet in the store.  Duplicates are allowed",
			struct {
				Name string `json:"name"`
				Tag  string `json:"tag"`
			}{},
			sashay.NewResponse(200, "pet response", Pet{}), // Normally a 201
			Error{},
		), CreatePet,
	}, {
		sashay.NewOperation(
			"GET",
			"/pets/:id",
			"Returns a user based on a single ID, if the user does not have access to the pet",
			struct {
				ID int `path:"id" description:"ID of pet to fetch"`
			}{},
			Pet{},
			Error{},
		), GetPet,
	}, {
		sashay.NewOperation(
			"DELETE",
			"/pets/:id",
			"deletes a single pet based on the ID supplied",
			struct {
				ID int `path:"id" description:"ID of pet to delete"`
			}{},
			nil, // 204 response
			Error{},
		), DeletePet,
	},
}

func Example_petstore() {
	sw := PetStoreSwagger()
	RegisterRoutes(&FrameworkRouter{}, sw)
	fmt.Println(sw.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: Swagger Petstore
	//   description: A sample API that uses a petstore as an example to demonstrate features in the OpenAPI 3.0 specification
	//   termsOfService: http://swagger.io/terms/
	//   contact:
	//     email: apiteam@swagger.io
	//   license:
	//     name: Apache 2.0
	//     url: http://www.apache.org/licenses/LICENSE-2.0.html
	//   version: 1.0.0
	// tags:
	//   - name: pet
	//     description: Everything about your Pets
	//   - name: store
	//     description: Access to Petstore orders
	//   - name: user
	//     description: Operations about user
	// servers:
	//   - url: http://petstore.swagger.io/api
	//     description: Public API server
	// paths:
	//   /pets:
	//     get:
	//       operationId: getPets
	//       summary: Returns all pets from the system that the user has access to
	//       parameters:
	//         - name: tags
	//           in: query
	//           description: tags to filter by
	//           schema:
	//             type: array
	//             items:
	//               type: string
	//         - name: limit
	//           in: query
	//           description: maximum number of results to return
	//           schema:
	//             type: integer
	//             format: int32
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
	//       summary: Creates a new pet in the store.  Duplicates are allowed
	//       requestBody:
	//         required: true
	//         content:
	//           application/json:
	//             schema:
	//               type: object
	//               properties:
	//                 name:
	//                   type: string
	//                 tag:
	//                   type: string
	//       responses:
	//         '200':
	//           description: pet response
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
	//       summary: Returns a user based on a single ID, if the user does not have access to the pet
	//       parameters:
	//         - name: id
	//           in: path
	//           required: true
	//           description: ID of pet to fetch
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
	//     delete:
	//       operationId: deletePetsId
	//       summary: deletes a single pet based on the ID supplied
	//       parameters:
	//         - name: id
	//           in: path
	//           required: true
	//           description: ID of pet to delete
	//           schema:
	//             type: integer
	//             format: int64
	//       responses:
	//         '204':
	//           description: The operation completed successfully.
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
	//   securitySchemes:
	//     apiKeyAuth:
	//       type: apiKey
	//       in: header
	//       name: api_key
	// security:
	//   - apiKeyAuth: []
}
