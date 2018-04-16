/*
Package sashay generates OpenAPI 3.0 (Swagger) documentation for a service.
See https://swagger.io/specification/ for more info about the spec.

swagger allows you to document your Go APIs using executable Go code,
including the same types you use for parameter declaration and serialization.
You can get a "good enough" Swagger document with very little work,
using the code you already have! Creating a nicer Swagger document is generally a matter
of adding a bit of annotation to struct tags or using some Swagger types around your API's types.

There are generally three parts to defining and generating Swagger docs using this package:

- Define the swagger.Swagger registry which holds all information that will be in the document.
  This is usually a singleton for an entire service, or passed to all route registration.
- Define new swagger.Operation instances where you have your handlers,
  adding them to the global registry using swagger.Swagger#Add as you go.
- Generate the docs with swagger.Swagger#WriteYAML.

In the following sections, we will go through the steps to build something very similar to
the "Pet Store API" Swagger example. This is the default example API at https:/editor.swagger.io/#/.

The "Pet Store API" OpenAPI 3.0 YAML file the service is based on is here:
https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore-expanded.yaml

There is code for a "Pet Store API" Go server here:
https://github.com/swagger-api/swagger-codegen/blob/master/samples/server/petstore/go-api-server/go/routers.go
Note that this is for their Swagger 2.0 definition, which is much larger than the 3.0 definition.

Our code will be based off that Pet Store code. There are many ways to structure a Go service;
the Pet Store example is only one such structure, with centralized routing and general HTTP handlers.
swagger package, being a library, can fit into any application setup.
It just needs to get the right calls, which should be clear by the end, as the API has very few moving parts.


Using the Swagger object to define top-level settings

In our example petstore.yaml file, we have the following settings that apply to the service,
rather than any specific paths, operations, or resources:

	openapi: 3.0.0
	info:
	  title: Swagger Petstore
	  description: A sample API that uses a petstore as an example to demonstrate features in the OpenAPI 3.0 specification
	  termsOfService: http://swagger.io/terms/
	  contact:
	    email: apiteam@swagger.io
	  license:
	    name: Apache 2.0
	    url: http://www.apache.org/licenses/LICENSE-2.0.html
	  version: 1.0.0
	tags:
	  - name: pet
	    description: Everything about your Pets
	  - name: store
	    description: Access to Petstore orders
	  - name: user
	    description: Operations about user
	servers:
	  - url: http://petstore.swagger.io/api
	    description: Public API server
	security:
	  - apiKeyAuth: []

We can use the following swagger code to create a *swagger.Swagger object that will generate that YAML.
This can be a stateful singleton, placed somewhere accessible to all handlers and routers,
like some common or config file.
In our example, we create it in our main/StartServer function, and pass it to the router:

	// main.go
	func PetStoreSwagger() *swagger.Swagger {
		return swagger.New(
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

The way this code maps to the YAML should be pretty self-explanatory.
For more information on any of these, you can refer to the OpenAPI documentation,
as it maps cleanly.

This code uses "apiKey" security, via AddAPIKeySecurity. The swagger.Swagger object also has
AddBasicAuthSecurity and AddJWTSecurity methods available.


Defining operations

An "operation" in OpenAPI 3.0 is a description for a path/route and method.
For example, here is the GET /pets endpoint Swagger YAML:

	paths:
	  /pets:
	    get:
	      operationId: getPets
	      summary: Returns all pets from the system that the user has access to
	      parameters:
	        - name: tags
	          in: query
	          description: tags to filter by
	          schema:
	            type: array
	            items:
	              type: string
	        - name: limit
	          in: query
	          description: maximum number of results to return
	          schema:
	            type: integer
	            format: int32
	      responses:
	        '200':
	          description: ok response
	          content:
	            application/json:
	              schema:
	                type: array
	                items:
	                  $ref: '#/components/schemas/Pet'
	        'default':
	          description: error response
	          content:
	            application/json:
	              schema:
	                $ref: '#/components/schemas/Error'

Let's go through the Go code required for that YAML.

First there is the code for the models.
These are probably not endpoint-specific, but shared for the entire application.
There is nothing swagger-related to this code; it already exists for the service:

	type Pet struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Tag  string `json:"tag"`
	}
	type Error struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	}

Next there is the actual route handler. This also has nothing Swagger-specific.
It is code that already exists for your service.

	func GetPets (http.ResponseWriter, *http.Request) {
		// Your code here
	}

Finally, we get to the route definitions/registration.
This, too, is something that needs to happen for any service.
The changes here have to do with registering a route adding it both to your HTTP framework's router,
and the swagger.Swagger registry.
Note that the swagger.Operation object has the method and path necessary to
register routes in pretty much every framework.
In this code, we have a custom Route struct that marries the Operation along with an http.HandlerFunc.

	type Route struct {
		operation swagger.Operation
		handler   http.HandlerFunc
	}

	func RegisterRoutes(router *FrameworkRouter, sw *swagger.Swagger) {
		for _, route := range routes {
			sw.Add(route.operation)
			router.AddRoute(route.operation.Method, route.operation.Path, route.handler)
		}
	}

	var routes = []Route{
		{
			swagger.NewOperation(
				"GET",
				"/pets",
				"Returns all pets from the system that the user has access to",
				struct {
					Tags  []string `query:"tags" description:"tags to filter by"`
					Limit int32    `query:"limit" description:"maximum number of results to return"`
				}{},
				[]Pet{},
				Error{},
			),
			GetPets,
		},
	}


Generating the Swagger file

Finally, there is the server startup code, usually in some sort of main() function.
This code initializes a new swagger.Swagger instance, registers routes,
and writes to a yaml file if the program is run with a -swagger argument.
This code in particular is going to be different depending on your conventions;
the following code is just an idea to show how this all fits together.

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

	func PetStoreSwagger() *swagger.Swagger {
		return swagger.New(
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

That's all there is to it. You can see a fuller example in the petstore_test.go file,
which contains the preceding code but with more routes.


Parameters

The swagger.Operation object supports defining an endpoint's parameters.
Because parameter settings can be quite detailed,
this package will parse some parameter settings from struct tags.
Let's look at the Parameters field in the following swagger.Operation definition:

	swagger.NewOperation(
		"POST",
		"/users/:id",
		"Update the user.",
		struct {
			ID int `path:"id" validate:"min=1"`
			Pretty bool `query:"pretty" description:"If true, return pretty-printed JSON." default:"true"`
			NoResponse bool `header:"X-NO-RESPONSE" description:"If true, return a 204 rather than the updated User."`
			Name string `json:"name"`
		}{},
		nil,
		nil,
	)

The struct tags of "path", "header", and "query" define the name of the parameter in the path/header/query.
Using the "json" tag indicates the parameter is included in the request body.
This Operation generates the following YAML:

	paths:
	  /users/{id}:
		post:
		  operationId: postUsersId
		  summary: Update the user.
		  parameters:
			- name: id
			  in: path
			  required: true
			  schema:
				type: integer
				format: int64
			- name: pretty
			  in: query
			  description: If true, return pretty-printed JSON.
			  schema:
				type: boolean
				default: true
			- name: X-NO-RESPONSE
			  in: header
			  description: If true, return a 204 rather than the updated User.
			  schema:
				type: boolean
		  requestBody:
			required: true
			content:
			  application/json:
				schema:
				  type: object
				  properties:
					name:
					  type: string
		  responses:
			'204':
			  description: The operation completed successfully.
			'default':
			  description: error response

The parameter struct definitions are nice, but the best feature is that they are actually executable Go code
that you can use for the parameter validation and binding in your own endpoints!
In practice, your Operation definitions will look something like this:

	type getUsersParams struct {
		Status string `query:"status" validate:"eq=active|eq=deleted"`
	}
	getUsersOp := swagger.NewOperation(
		"GET",
		"/users",
		"Get users",
		getUsersParams{},
		[]User{},
		ErrorModel{},
	)
	getUsersHandler := func(c echo.Context) error {
		params := getUsersParams{}
		if err := c.Bind(&params); err != nil {
			return err
		}
		if err := c.Validate(params); err != nil {
			return err
		}
		var users []User
		// Logic to get users
		return c.JSON(200, users)
	}

	sw.Add(getUsersOp)
	router.Add(getUsersOp.Method, getUsersOp.Path, getUsersHandler)

The actual getUsersHandler code uses the same struct to describe itself as it does in code.
The same is true for response types- the schema is built from the real objects, with the json struct tags,
not separate documentation.


Advanced Parameter usage

Struct types can also be used in parameters.
Usually, these will be nested structs for request bodies:

	swagger.NewOperation(
		"POST",
		"/users",
		"Create a user.",
		struct {
			Name struct {
				First string `json:"first"`
				Last string `json:"last"`
			} `json:"name"`
		}{},
		nil,
		nil,
	)

You can see the requestBody YAML it generates:

	paths:
	  /users:
		post:
		  operationId: postUsers
		  summary: Create a user.
		  requestBody:
			required: true
			content:
			  application/json:
				schema:
				  type: object
				  properties:
					name:
					  type: object
					  properties:
						first:
						  type: string
						last:
						  type: string

However, sometimes you want to use struct types that are represented as data types.
Times are an exampmle of this- time.Time is a Go struct type,
but we want to represent it with a string data type in Swagger.
We can define custom "data types" to do this ("data type" is a Swagger term).
For example, here is what maps time.Time to a {type: string, format: date-time} in generated Swagger:

	sw.DefineDataType(
		time.Time{},
		ChainDataTyper(
			SimpleDataTyper("string", "date-time"),
			DefaultDataTyper()))

swagger.Swagger#DefineDataType takes in an instance of a value to map into a data type
and the DataTyper transformer function.
DefaultDataTyper() pulls the string from the 'default' struct tag when this value occurs on a struct field,
and SimpleDataTyper uses the given type and format strings. ChainDataTyper calls one DataTyper after another.

For example, perhaps you have a type that represents a time.Time instance and some arbitrary unit.
We can define a custom DataTyper that will look for a particular enum tag,
and use that to inform the format string:

	type UnitOfTime struct {
		time time.Time
		unit string
	}
	sw.DefineDataType(UnitOfTime{}, func(tvp swagger.Field) swagger.ObjectFields {
		of := swagger.ObjectFields{"type": "string"}
		if timeunit := tvp.StructField.Tag.Get("timeunit"); timeunit != "" {
			switch timeunit {
			case "date":
				of["format"] = "date"
			case "month":
				of["format"] = "YYYY-MM"
			}
		}
		return of
	})

We can also override the default data type behavior,
such as if we want to look at an enum tag for possible string values:

	sw.DefineDataType("", func(tvp swagger.Field) swagger.ObjectFields {
		of := swagger.ObjectFields{"type": "string"}
		if enum := tvp.StructField.Tag.Get("enum"); enum != "" {
			values := strings.Split(enum, "|")
			of["enum"] = fmt.Sprintf("['%s']", strings.Join(values, "', '"))
		}
		return of
	})

We can put it together in the following Operation params:

	struct {
		StartMonth UnitOfTime `json:"startMonth" timeunit:"month"`
		EndDay UnitOfTime `json:"endDay" timeunit:"date"`
		Status string `json:"status" enum:"on|off"`
	}

That will generate the following requestBody YAML.
Note the custom "format" strings for startMonth and endDay,
and the "enum" values for status:

	requestBody:
	  required: true
	  content:
		application/json:
		  schema:
			type: object
			properties:
			  startMonth:
				type: string
				format: YYYY-MM
			  endDay:
				type: string
				format: date
			  status:
				type: string
				enum: ['on', 'off']

The goal, again, is to reuse as much of your existing code as possible,
and to build off it rather than require a bunch of custom annotation or documentation.
In practice, this often means pulling this sort of data out of "validation" struct tags,
rather than custom struct tags, but the idea is the same.


Responses

The other part of Operations that may require some customization are usually responses.
The swagger package tries to be smart and enforce some conventions:

- Successful POSTs returns a 201.
- All other successful methods return a 200.
- All operations get a 'default' error response.

For example, let's look at the Go code to fetch an array of users.
Note that slice types are handled properly:

	swagger.NewOperation(
		"GET",
		"/users",
		"",
		nil,
		[]User{},
		ErrorModel{},
	)

And the corresponding YAML.
Note in particular that the 200 response is an array that points to references of the User schema,
and the User and Error Model are defined in components/schemas:

	paths:
	  /users:
		get:
		  operationId: getUsers
		  responses:
			'200':
			  description: ok response
			  content:
				application/json:
				  schema:
					type: array
					items:
					  $ref: '#/components/schemas/User'
			'default':
			  description: error response
			  content:
				application/json:
				  schema:
					$ref: '#/components/schemas/ErrorModel'
	components:
	  schemas:
		ErrorModel:
		  type: object
		  properties:
			error:
			  type: object
			  properties:
				message:
				  type: string
				code:
				  type: integer
				  format: int64
		User:
		  type: object
		  properties:
			result:
			  type: object
			  properties:
				id:
				  type: integer
				  format: int64
				name:
				  type: string

However, sometimes you need more advanced response information.
In particular, you may want to document specific error conditions or return type shapes.
You can use the swagger.Response or swagger.Responses object for this:

	swagger.NewOperation(
		"GET",
		"/is_teapot",
		"Error if the server is a teapot.",
		nil,
		swagger.Responses{
			swagger.NewResponse(200, "Not a teapot.", TeapotResponse{}),
			swagger.NewResponse(203, "I may not be a teapot", TeapotResponse{}),
		},
		swagger.NewResponse(418, "Yes, I am sure a teapot!", TeapotError{}),
	)

Note the calls to NewResponse, and the Responses slice.
In this way, the default codes can be overwritten, and multiple responses can be specified:

	paths:
	  /is_teapot:
		get:
		  operationId: getIsTeapot
		  summary: Error if the server is a teapot.
		  responses:
			'200':
			  description: Not a teapot.
			  content:
				application/json:
				  schema:
					$ref: '#/components/schemas/TeapotResponse'
			'203':
			  description: I may not be a teapot
			  content:
				application/json:
				  schema:
					$ref: '#/components/schemas/TeapotResponse'
			'418':
			  description: Yes, I am sure a teapot!
			  content:
				application/json:
				  schema:
					$ref: '#/components/schemas/TeapotError'
*/
package sashay
