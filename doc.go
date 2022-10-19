/*
Package sashay allows you to generate OpenAPI 3.0 (Swagger) files using executable Go code,
including the same types you already use for parameter declaration and serialization.

You don't have to worry about creating extensive Swagger-specific comments
or editing a Swagger file by hand.
You can get a good enough Swagger document with very little work,
using the code you already have!

- Use your existing serializable Go structs to document what an endpoint returns.
Really, Sashay will figure out the OpenAPI contents using reflection.

- Declare your parameters using Go structs. If you are binding and validating using structs in your endpoint handlers,
you can use the same structs for Sashay.

- You can extend Sashay to handle your own types and struct tags,
such as if you use custom time/date types,
or want to parse validation struct tags into something you can place in your OpenAPI doc.

Creating a nicer OpenAPI 3.0 document from your existing code is generally a matter of adding
a bit of annotation to struct tags or using some Sashay types around your API's types.

See https://swagger.io/specification/ for more information about the OpenAPI 3.0 spec.

# Sashay Tutorial

There are generally three parts to defining and generating Swagger docs using Sashay:

- Define the sashay.Sashay registry which holds all information that will be in the document.
This is usually a singleton for an entire service, or passed to all route registration.

- Define new sashay.Operation instances where you have your handlers,
adding them to the registry using the Add method as you go.

- Generate the YAML string using the WriteYAML method.

In the following sections, we will go through the steps to build something very similar to
the "Pet Store API" Swagger example. This is the default example API at https:/editor.swagger.io/#/.

The "Pet Store API" OpenAPI 3.0 YAML file the service is based on is here:
https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore-expanded.yaml

There is code for a "Pet Store API" Go server here:
https://github.com/swagger-api/swagger-codegen/blob/master/samples/server/petstore/go-api-server/go/routers.go
Note that this is for their Swagger 2.0 definition, which is much larger than the 3.0 definition.

Our code will be based off that Pet Store code. There are many ways to structure a Go service;
the Pet Store example is only one such structure, with centralized routing and general HTTP handlers.
Sashay, being a library, can fit into any application setup.
It just needs to get the right calls, which should be clear by the end, as the API has very few moving parts.

# Tutorial Step 1- Define Service Level Settings

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

We can use the following code to create a *sashay.Sashay object that will generate that YAML.
This can be a stateful singleton, placed somewhere accessible to all handlers and routers,
like some common or config file.
Later in our example, we create the instance in our main function,
and pass it to the router:

	sa := sashay.New(
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

The way this code maps to the YAML should be pretty self-explanatory.
For more information on any of these, you can refer to the OpenAPI documentation,
as it maps cleanly.

This code uses "apiKey" security, via AddAPIKeySecurity. The sashay.Sashay object also has
AddBasicAuthSecurity and AddJWTSecurity methods available.

# Tutorial Step 2- Define Operations

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
and the sashay.Sashay registry.
Note that the sashay.Operation object has the method and path necessary to
register routes in pretty much every framework.
In this code, we have a custom Route struct that marries the Operation along with an http.HandlerFunc.

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
			),
			GetPets,
		},
	}

# Tutorial Step 3- Generate the OpenAPI File

Finally, there is the server startup code, usually in some sort of main() function.
This code initializes a new sashay.Sashay instance, registers routes,
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

	func PetStoreSwagger() *sashay.Sashay {
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

# Sashay Detail- Basic Parameters

The sashay.Operation object supports defining an endpoint's parameters.
Because parameter settings can be quite detailed,
this package will parse some parameter settings from struct tags.
Let's look at the Parameters field in the following sashay.Operation definition:

	sashay.NewOperation(
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
	getUsersOp := sashay.NewOperation(
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

Note that Sashay never uses $ref for parameters (resources in POST/PUT request bodies).
Even if the same type is used for a request and a response,
it'll be expanded in the requestBody section and a $ref in the response section.
This may change in the future.

# Sashay Detail- Request Bodies

Struct types can also be used in parameters.
Usually, these will be nested structs for request bodies:

	sashay.NewOperation(
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

# Sashay Detail- Representing Custom Types

Note that out of the box, Sashay will treat simple custom types (like `type MyString string`)
as their underlying simple type, and will walk any custom structs.

However, sometimes you want to use Go struct types that are represented as data types in Swagger.
Times are an exampmle of this: time.Time is a Go struct type,
but we want to represent it with a string data type in Swagger (type: string, format: date-time).
For example, let's say "month" is a common concept in our API, so we represent it with a type:

	type Month struct {
		Year int
		Month int
	}

	type Params struct {
		Month Month `query:"month"`
	}

When we have a struct field with a type of MyTime, we would normally get a schema of:

	type: object
	properties:
	  time:
	    type: object
	    properties:
	      year:
	        type: integer
	      month:
	        type: integer

However, what we actually want is something like this:

	type: object
	properties:
	  time:
	    type: string
	    format: YYYY-MM

We can define a mapping between custom types and a "data type transformer" to do this.
For example, to get the desired Swagger we would use a SimpleDataTyper transformer:

	sa.DefineDataType(Month{}, SimpleDataTyper("string", "YYYY-MM"))

DefineDataType takes in an instance of a value to map into a data type
and the DataTyper transformer function.
SimpleDataTyper uses the given type and format strings.

Sashay includes other built-in DataTypers:

- DefaultDataTyper() will parse the "default" struct tag and write it into the "default" field.

- ChainDataTyper calls one DataTyper after another.
The most common usage is to use this around SimpleDataTyper and DefaultDataTyper,
but feel free to get creative.

- BuiltinDataTyperFor returns the default DataTyper behavior for a type.
This is useful when you want to extend the behavior for a built-in type,
but not entirely replace it (we use it below, for a custom string data typer behavior).

The DataTyper function can get more creative, too.
For example, it can parse struct fields to inform what should write into the Swagger file.
Consider a "unit of time" type that can be used for any unit, rather than custom month, day, etc types:

	type UnitOfTime struct {
		time time.Time
		unit string
	}

And using it for parameters looks like:

	type Params struct {
		Month UnitOfTime `query:"month" timeunit:"month"`
	}

We could use a DataTyper that reads the "timeunit" struct tag,
and specifies the "format" field based on that:

	sw.DefineDataType(UnitOfTime{}, func(f sashay.Field, of sashay.ObjectFields) {
		of["type"] = "string"
		if timeunit := f.StructField.Tag.Get("timeunit"); timeunit != "" {
			switch timeunit {
			case "date":
				of["format"] = "date"
			case "month":
				of["format"] = "YYYY-MM"
			}
		}
	})

# Sashay Detail- Other Advanced DataTyper Usage

We can use DefineDataType to customize all sorts of behavior.
One common usage is parsing tags to specify other information about a field, like we did with "timeunit" above.
Perhaps we want to parse an "enum" tag that specifies valid values for a string field:

	extractEnum := func(f sashay.Field, of sashay.ObjectFields) {
		of["type"] = "string"
		if enum := f.StructField.Tag.Get("enum"); enum != "" {
			values := strings.Split(enum, "|")
			of["enum"] = fmt.Sprintf("['%s']", strings.Join(values, "', '"))
		}
	}
	sw.DefineDataType("", sashay.BuiltinDataTyperFor("", extractEnum))

Now, when we have a string with the "enum" struct tag, we will get the "enum" field in our YAML:

	type Params struct {
		Status string `json:"status" enum:"on|off"`
	}

	schema:
	  type: object
	  properties:
	    status:
	      type: string
	      enum: ['on', 'off']

The goal of Sashay is, you may recall, to reuse as much of your existing code as possible,
and to build off it rather than require a bunch of custom annotation or documentation.
In practice, this often means pulling this sort of data out of "validation" struct tags,
rather than custom struct tags like "enum" or "timeunit", but the idea is the same.

For an example of this in action, and a good basis for hooking your own validation needs up to Sashay,
see validator_data_typer_test.go. It includes a fully-functional example using go-validator style struct tags
to inform data type fields.

# Sashay Detail- Responses

The other part of sashay.Operation that may require some customization are usually responses.
Sashay tries to be smart and enforce some conventions:

- Successful POSTs returns a 201.
- All other successful methods return a 200.
- All operations get a 'default' error response.

For example, let's look at the Go code to fetch an array of users
(we can use an empty User slice, or a custom Users slice type would work fine).

	sashay.NewOperation(
		"GET",
		"/users",
		"",
		nil,
		[]User{},
		ErrorModel{},
	)

The 200 response is an array that points to references of the User schema,
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

	sashay.NewOperation(
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

Finally, there are a couple special cases for responses:

- If a response is a string type, rather than a struct,
it is assumed to be of content type text/plain.

- If a response is an empty struct (`struct{}{}`), use application/json with no schema.

# Sashay Detail- Pointer Fields

Sashay treats value and pointer fields the same.
In other words, *bool and bool will use the same data type/schema.
When you register a data type (refer to DefineDataType),
the same DataTyper is used for pointer fields of that type.

The primary use case for pointer fields in Go is to represent optional fields.
There's nothing much for Sashay to do with that information,
because both parameters and object fields are optional/not-required in Swagger by default.
For example, in parameters, "required: false" is the default.
And for schemas (request bodies, responses), the "nullable: true" attribute
is quite semantically different than the "optional" meant by a Go pointer field.

In the future, Sashay may support more more extensive specification around required fields,
but not right now.
*/
package sashay
