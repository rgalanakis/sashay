package sashay_test

import (
	"bytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rgalanakis/sashay"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSashay(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sashay Suite")
}

type User struct {
	Result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}

type ErrorModel struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

var _ = Describe("Sashay", func() {

	var (
		sw *sashay.Sashay
	)

	BeforeEach(func() {
		sw = sashay.New(
			"SwaggerGenAPI",
			"Demonstrate auto-generating Swagger",
			"0.1.9",
		)
	})

	It("generates the info section", func() {
		Expect(sw.BuildYAML()).To(ContainSubstring(`openapi: 3.0.0
info:
  title: SwaggerGenAPI
  description: Demonstrate auto-generating Swagger
  version: 0.1.9
paths:
`))
	})

	It("can set more fields in the info section", func() {
		sw.SetTermsOfService("http://example.com/terms/").
			SetContact("API Support", "http://www.example.com/support", "support@example.com").
			SetLicense("Apache 2.0", "https://www.apache.org/licenses/LICENSE-2.0.html").
			AddTag("tagA", "its a tag")
		Expect(sw.BuildYAML()).To(ContainSubstring(`openapi: 3.0.0
info:
  title: SwaggerGenAPI
  description: Demonstrate auto-generating Swagger
  termsOfService: http://example.com/terms/
  contact:
    name: API Support
    url: http://www.example.com/support
    email: support@example.com
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html
  version: 0.1.9
tags:
  - name: tagA
    description: its a tag
`))
	})

	It("generates a server section", func() {
		sw.AddServer("https://api.example.com/v1", "Production server.")
		Expect(sw.BuildYAML()).To(ContainSubstring(`servers:
  - url: https://api.example.com/v1
    description: Production server.
`))
	})

	It("has no server section if that are no servers", func() {
		Expect(sw.BuildYAML()).To(Not(ContainSubstring(`servers:`)))
	})

	It("can set global security types", func() {
		sw.AddBasicAuthSecurity()
		sw.AddJWTSecurity()
		sw.AddAPIKeySecurity("header", "X-MY-APIKEY")
		Expect(sw.BuildYAML()).To(ContainSubstring(`components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic
    bearerAuth:
      type: http
      bearerFormat: JWT
      scheme: bearer
    apiKeyAuth:
      type: apiKey
      in: header
      name: X-MY-APIKEY
security:
  - basicAuth: []
  - bearerAuth: []
  - apiKeyAuth: []
`))
	})

	It("generates paths for routes with no parameters", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"Returns a list of users.",
			nil,
			[]User{},
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    get:
      operationId: getUsers
      summary: Returns a list of users.
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
`))
	})

	It("keeps routes and methods in the same order", func() {
		newOp := func(method, path string) sashay.Operation {
			return sashay.NewOperation(
				method, path, "", nil, nil, nil)
		}
		ops := []sashay.Operation{
			newOp("POST", "/zzz"),
			newOp("POST", "/aaa"),
			newOp("GET", "/aaa"),
			newOp("GET", "/zzz"),
			newOp("DELETE", "/aaa"),
			newOp("PUT", "/aaa"),
			newOp("PATCH", "/aaa"),
		}
		makeSwagger := func() *sashay.Sashay {
			sw := sashay.New("title", "desc", "0.0.1")
			list := rand.Perm(len(ops))
			for _, i := range list {
				sw.Add(ops[i])
			}
			return sw
		}
		for i := 0; i < 20; i++ {
			Expect(makeSwagger().BuildYAML()).To(ContainSubstring(`paths:
  /aaa:
    get:
      operationId: getAaa
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
    post:
      operationId: postAaa
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
    put:
      operationId: putAaa
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
    patch:
      operationId: patchAaa
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
    delete:
      operationId: deleteAaa
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
  /zzz:
    get:
      operationId: getZzz
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
    post:
      operationId: postZzz
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
`))
		}
	})

	It("can uses swagger.Responses for response fields", func() {
		type TeapotResponse struct {
			Probability float64 `json:"prob"`
		}
		type TeapotError struct {
			Strength float64 `json:"strength"`
		}
		sw.Add(sashay.NewOperation(
			"GET",
			"/is_teapot",
			"Error if the server is a teapot.",
			nil,
			sashay.Responses{sashay.NewResponse(203, "I may not be a teapot", TeapotResponse{})},
			sashay.Responses{sashay.NewResponse(418, "Yes, I am sure a teapot!", TeapotError{})},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /is_teapot:
    get:
      operationId: getIsTeapot
      summary: Error if the server is a teapot.
      responses:
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
`))
	})

	It("can uses swagger.Response for response fields", func() {
		type TeapotResponse struct {
			Probability float64 `json:"prob"`
		}
		type TeapotError struct {
			Strength float64 `json:"strength"`
		}
		sw.Add(sashay.NewOperation(
			"GET",
			"/is_teapot",
			"Error if the server is a teapot.",
			nil,
			sashay.NewResponse(203, "I may not be a teapot", TeapotResponse{}),
			sashay.NewResponse(418, "Yes, I am sure a teapot!", TeapotError{}),
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /is_teapot:
    get:
      operationId: getIsTeapot
      summary: Error if the server is a teapot.
      responses:
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
`))
	})

	It("interprets string responses as text/plain", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/plain",
			"plain",
			nil,
			sashay.NewResponse(200, "desc", ""),
			""),
		)
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /plain:
    get:
      operationId: getPlain
      summary: plain
      responses:
        '200':
          description: desc
          content:
            text/plain:
              schema:
                type: string
        'default':
          description: error response
          content:
            text/plain:
              schema:
                type: string
`))
	})

	It("interprets empty structs as plain objects", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/plain",
			"plain",
			nil,
			struct{}{},
			""),
		)
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /plain:
    get:
      operationId: getPlain
      summary: plain
      responses:
        '200':
          description: ok response
          content:
            application/json:
              schema:
                type: object
        'default':
`))
	})

	It("generates paths with descriptions and tags", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"Returns a list of users.",
			nil,
			nil,
			nil).
			WithDescription("CommonMark description.").
			AddTags("tagA", "tagB"))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    get:
      tags: ["tagA", "tagB"]
      operationId: getUsers
      summary: Returns a list of users.
      description: CommonMark description.
`))
	})

	It("generates paths for routes with annotated path and query parameters", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users/:id",
			"Returns the ID'd user.",
			struct {
				ID     int  `path:"id" validate:"min=1" description:"CommonMark description."`
				Pretty bool `query:"pretty" default:"true"`
			}{},
			User{},
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users/{id}:
    get:
      operationId: getUsersId
      summary: Returns the ID'd user.
      parameters:
        - name: id
          in: path
          required: true
          description: CommonMark description.
          schema:
            type: integer
            format: int64
        - name: pretty
          in: query
          schema:
            type: boolean
            default: true
      responses:
        '200':
          description: ok response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        'default':
          description: error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`))
	})

	It("can use an alternative global content type", func() {
		sw.DefaultContentType = "application/myapp+json+v1"
		sw.Add(sashay.NewOperation(
			"POST",
			"/users/:id",
			"Update the user",
			struct {
				ID   string `path:"id"`
				Name string `json:"name"`
			}{},
			User{},
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users/{id}:
    post:
      operationId: postUsersId
      summary: Update the user
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/myapp+json+v1:
            schema:
              type: object
              properties:
                name:
                  type: string
      responses:
        '201':
          description: ok response
          content:
            application/myapp+json+v1:
              schema:
                $ref: '#/components/schemas/User'
        'default':
          description: error response
          content:
            application/myapp+json+v1:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`))
	})

	It("generates schemas for return and error types", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"",
			nil,
			[]User{},
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`components:
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
`))
	})

	It("generates schemas for types with simple arrays", func() {
		type Tagger struct {
			Tags []string `json:"tags"`
		}
		sw.Add(sashay.NewOperation(
			"GET",
			"/tags",
			"",
			nil,
			Tagger{},
			nil,
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`components:
  schemas:
    Tagger:
      type: object
      properties:
        tags:
          type: array
          items:
            type: string
`))
	})

	It("generates schemas for named structs under anonymous structs", func() {
		type FieldB struct {
			FieldC string `json:"fieldC"`
		}
		type FieldA struct {
			Wrapper struct {
				FieldB FieldB `json:"fieldB"`
			} `json:"wrapper"`
		}
		type Response struct {
			Result struct {
				FieldA FieldA `json:"fieldA"`
			} `json:"result"`
		}

		sw.Add(sashay.NewOperation(
			"GET",
			"/tags",
			"",
			nil,
			Response{},
			nil,
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`components:
  schemas:
    FieldA:
      type: object
      properties:
        wrapper:
          type: object
          properties:
            fieldB:
              $ref: '#/components/schemas/FieldB'
    FieldB:
      type: object
      properties:
        fieldC:
          type: string
    Response:
      type: object
      properties:
        result:
          type: object
          properties:
            fieldA:
              $ref: '#/components/schemas/FieldA'
`))
	})

	It("generates schemas for embedded structs", func() {
		type Inside struct {
			nothere     int
			InsideField string `json:"insideField"`
		}
		type ExportedBase struct {
			hidden         bool
			ExportedResult struct {
				ExportedInside Inside `json:"exportedInside"`
			} `json:"exportedResult"`
		}
		type unexportedBase struct {
			unexported       string
			UnexportedResult struct {
				UnexportedInside Inside `json:"unexportedInside"`
			} `json:"unexportedResult"`
		}
		type Response struct {
			ExportedBase
			unexportedBase
		}

		sw.Add(sashay.NewOperation(
			"GET",
			"/tags",
			"",
			nil,
			Response{},
			nil,
		))
		Expect(sw.BuildYAML()).To(HaveSuffix(`components:
  schemas:
    Inside:
      type: object
      properties:
        insideField:
          type: string
    Response:
      type: object
      properties:
        exportedResult:
          type: object
          properties:
            exportedInside:
              $ref: '#/components/schemas/Inside'
        unexportedResult:
          type: object
          properties:
            unexportedInside:
              $ref: '#/components/schemas/Inside'
`))
	})

	It("generates schemas for POSTs with request bodies", func() {
		sw.Add(sashay.NewOperation(
			"POST",
			"/users",
			"Creates a new user.",
			struct {
				Name   string `json:"name"`
				Pretty bool   `query:"pretty"`
			}{},
			User{},
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    post:
      operationId: postUsers
      summary: Creates a new user.
      parameters:
        - name: pretty
          in: query
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
        '201':
          description: ok response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        'default':
          description: error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`))
	})

	It("does not include requestBody for POST/PUT with no parameters", func() {
		sw.Add(sashay.NewOperation(
			"POST",
			"/checkin",
			"",
			nil,
			nil,
			nil,
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /checkin:
    post:
      operationId: postCheckin
      responses:
`))
	})

	It("uses a 204 success if an endpoint has no response struct", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/ping",
			"Check health.",
			nil,
			nil,
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /ping:
    get:
      operationId: getPing
      summary: Check health.
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorModel'
`))
	})

	It("can handle parameters that are slices of objects", func() {
		sw.Add(sashay.NewOperation(
			"POST",
			"/users",
			"Create a user.",
			struct {
				Name string `json:"name"`
				Tags []struct {
					TagName string `json:"tagName"`
				} `json:"tags"`
			}{},
			nil,
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
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
                  type: string
                tags:
                  type: array
                  items:
                    type: object
                    properties:
                      tagName:
                        type: string
`))
	})

	It("handles all supported simple types", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"Get users",
			struct {
				Boolean bool    `query:"pbool"`
				Float32 float32 `query:"pfloat32"`
				Float64 float64 `query:"pfloat64"`
				Int64   int64   `query:"pint64"`
				Int32   int32   `query:"pint32"`
			}{},
			nil,
			ErrorModel{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    get:
      operationId: getUsers
      summary: Get users
      parameters:
        - name: pbool
          in: query
          schema:
            type: boolean
        - name: pfloat32
          in: query
          schema:
            type: number
            format: float
        - name: pfloat64
          in: query
          schema:
            type: number
            format: double
        - name: pint64
          in: query
          schema:
            type: integer
            format: int64
        - name: pint32
          in: query
          schema:
            type: integer
            format: int32
`))
	})

	It("does not duplicate schemas", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"",
			nil,
			nil,
			ErrorModel{},
		))
		sw.Add(sashay.NewOperation(
			"GET",
			"/other_users",
			"",
			nil,
			nil,
			[]ErrorModel{},
		))
		yaml := sw.BuildYAML()
		cnt := strings.Count(yaml, "ErrorModel:")
		if cnt != 1 {
			Fail("Duplicate ErrorModel definitions in components:\n" + yaml)
		}
	})

	It("creates refs for struct fields that are themselves exported structs", func() {
		type Teacher struct {
			FullName string `json:"fullName"`
		}
		type Class struct {
			Subject string  `json:"subject"`
			Teacher Teacher `json:"teacher"`
		}
		type School struct {
			Mascot  string  `json:"mascot"`
			Classes []Class `json:"classes"`
		}
		sw.Add(sashay.NewOperation(
			"GET",
			"/schools",
			"Get schools.",
			nil,
			School{},
			nil,
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`components:
  schemas:
    Class:
      type: object
      properties:
        subject:
          type: string
        teacher:
          $ref: '#/components/schemas/Teacher'
    School:
      type: object
      properties:
        mascot:
          type: string
        classes:
          type: array
          items:
            $ref: '#/components/schemas/Class'
    Teacher:
      type: object
      properties:
        fullName:
          type: string
`))
	})

	It("does not try to use unexported struct fields", func() {
		type MeTime struct {
			impl int
			X    int `json:"x"`
		}
		type Moment struct {
			Time MeTime `json:"time"`
		}
		sw.Add(sashay.NewOperation(
			"GET",
			"/now",
			"",
			nil,
			Moment{},
			nil,
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`components:
  schemas:
    MeTime:
      type: object
      properties:
        x:
          type: integer
          format: int64
    Moment:
      type: object
      properties:
        time:
          $ref: '#/components/schemas/MeTime'
`))
	})

	It("does not include empty parameters or properties", func() {
		type Empty struct {
			impl int
		}
		type Wrapper struct {
			E Empty `json:"e"`
		}
		sw.Add(sashay.NewOperation(
			"GET",
			"/empty",
			"",
			struct{}{},
			Wrapper{},
			nil,
		))
		yaml := sw.BuildYAML()
		Expect(yaml).To(ContainSubstring(`operationId: getEmpty
      responses:
`))
		Expect(yaml).To(ContainSubstring(`components:
  schemas:
    Empty:
      type: object
    Wrapper:
      type: object
      properties:
        e:
          $ref: '#/components/schemas/Empty'
`))
	})

	It("maps Time fields to strings data types", func() {
		type Response struct {
			Time time.Time `json:"time"`
		}
		sw.Add(sashay.NewOperation(
			"POST",
			"/stuff",
			"Updates stuff.",
			struct {
				PTime time.Time `json:"ptime"`
			}{},
			Response{},
			nil,
		))
		yaml := sw.BuildYAML()
		Expect(yaml).To(ContainSubstring(`requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                ptime:
                  type: string
                  format: date-time
`))
		Expect(yaml).To(ContainSubstring(`Response:
      type: object
      properties:
        time:
          type: string
          format: date-time
`))
		Expect(yaml).To(Not(ContainSubstring("/Time"))) // No $ref link
	})

	It("can use custom data type definitions", func() {
		type Custom struct {
			Field string `json:"field"`
		}
		type Response struct {
			Custom Custom `json:"custom"`
		}
		sw.DefineDataType(Custom{}, sashay.SimpleDataTyper("boolean", ""))
		sw.Add(sashay.NewOperation(
			"POST",
			"/stuff",
			"Updates stuff.",
			struct {
				PCustom Custom `json:"pcustom"`
			}{},
			Response{},
			nil,
		))
		yaml := sw.BuildYAML()
		Expect(yaml).To(ContainSubstring(`requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                pcustom:
                  type: boolean
`))
		Expect(yaml).To(ContainSubstring(`Response:
      type: object
      properties:
        custom:
          type: boolean
`))
		Expect(yaml).To(Not(ContainSubstring("/Custom"))) // No $ref link
	})

	It("can override builtin data types", func() {
		sw.DefineDataType("", sashay.BuiltinDataTyperFor("", func(_ sashay.Field, of sashay.ObjectFields) {
			of["format"] = "hello"
		}))
		sw.Add(sashay.NewOperation(
			"POST",
			"/stuff",
			"Updates stuff.",
			struct {
				String string `json:"string"`
			}{},
			nil,
			nil,
		))
		yaml := sw.BuildYAML()
		Expect(yaml).To(ContainSubstring(`requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                string:
                  type: string
                  format: hello
`))
	})

	It("handles custom data types that derive simple kinds", func() {
		type MyInt int
		type myString string
		sw.Add(sashay.NewOperation(
			"POST",
			"/stuff",
			"Updates stuff.",
			struct {
				MyInt    MyInt    `json:"myInt"`
				MyString myString `json:"myString"`
			}{},
			nil,
			nil,
		))
		yaml := sw.BuildYAML()
		Expect(yaml).To(ContainSubstring(`requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                myInt:
                  type: integer
                  format: int64
                myString:
                  type: string
`))
	})

	It("panics if given an unsupported type", func() {
		type Custom struct{}
		sw.Add(sashay.NewOperation(
			"POST",
			"/stuff",
			"Updates stuff.",
			struct {
				Custom ***Custom `json:"custom"`
			}{},
			nil,
			nil,
		))
		Expect(func() {
			sw.BuildYAML()
		}).To(Panic())
	})

	It("can write to a custom buffer", func() {
		b := bytes.NewBuffer([]byte("hello there!\n"))
		sw.WriteYAML(b)
		Expect(b.String()).To(ContainSubstring(`hello there!
openapi: 3.0.0
info:
`))
	})

	It("treats pointer fields in parameters, body parameters, and responses the same as value fields", func() {
		type addressParam struct {
			Address1 string `json:"address1"`
			State    string `json:"state"`
		}
		type User struct {
			Name *string `json:"name"`
		}
		type Response struct {
			Users *[]User `json:"users"`
			User  *User   `json:"user"`
		}
		sw.Add(sashay.NewOperation(
			"POST",
			"/users/:id",
			"Update a user.",
			struct {
				ID        *int            `path:"id"`
				Pretty    *bool           `query:"pretty"`
				Name      *string         `json:"name"`
				Addresses *[]addressParam `json:"addresses"`
			}{},
			Response{},
			nil,
		))
		yaml := sw.BuildYAML()
		Expect(yaml).To(HaveSuffix(`paths:
  /users/{id}:
    post:
      operationId: postUsersId
      summary: Update a user.
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
        - name: pretty
          in: query
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
                addresses:
                  type: array
                  items:
                    type: object
                    properties:
                      address1:
                        type: string
                      state:
                        type: string
      responses:
        '201':
          description: ok response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Response'
        'default':
          description: error response
components:
  schemas:
    Response:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/User'
        user:
          $ref: '#/components/schemas/User'
    User:
      type: object
      properties:
        name:
          type: string
`))
	})

	It("expands fields of parameters (do not use $ref)", func() {
		type Address struct {
			Address1 string `json:"address1"`
			State    string `json:"state"`
		}
		type User struct {
			Name         string    `json:"name"`
			Address      Address   `json:"address"`
			OldAddresses []Address `json:"oldAddresses"`
		}
		sw.Add(sashay.NewOperation(
			"POST",
			"/users",
			"Create a user.",
			User{},
			nil,
			nil,
		))
		yaml := sw.BuildYAML()
		Expect(yaml).To(HaveSuffix(`paths:
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
                  type: string
                address:
                  type: object
                  properties:
                    address1:
                      type: string
                    state:
                      type: string
                oldAddresses:
                  type: array
                  items:
                    type: object
                    properties:
                      address1:
                        type: string
                      state:
                        type: string
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
`))
	})

	It("can create a new registry by filtering/mapping operations", func() {
		sw.DefaultContentType = "application/xml"
		sw.AddServer("server.com", "the server")
		sw.Add(sashay.NewOperation("GET", "/internal/users", "", nil, nil, nil))
		sw.Add(sashay.NewOperation("GET", "/users", "", nil, nil, nil))
		result := sashay.SelectMap(sw, func(op sashay.Operation) *sashay.Operation {
			if strings.Contains(op.Path, "/internal") {
				return nil
			}
			op.Path = "/v3" + op.Path
			return &op
		})
		Expect(result.BuildYAML()).To(ContainSubstring(`openapi: 3.0.0
info:
  title: SwaggerGenAPI
  description: Demonstrate auto-generating Swagger
  version: 0.1.9
servers:
  - url: server.com
    description: the server
paths:
  /v3/users:
    get:
      operationId: getV3Users
      responses:
        '204':
          description: The operation completed successfully.
        'default':
          description: error response
`))
	})

	It("writes to a file", func() {
		f, err := ioutil.TempFile("", "sashay")
		Expect(err).To(Not(HaveOccurred()))
		defer os.Remove(f.Name())
		sw.WriteYAMLFile(f.Name())

		contents, err := ioutil.ReadFile(f.Name())
		Expect(string(contents)).To(ContainSubstring("SwaggerGenAPI"))
	})

	It("can handle maps", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"",
			map[string]interface{}{},
			map[string]interface{}{},
			map[string]interface{}{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    get:
      operationId: getUsers
      requestBody:
        content:
          */*:
            schema:
              type: object
      responses:
        '200':
          description: ok response
          content:
            application/json:
              schema:
                type: object
        'default':
          description: error response
          content:
            application/json:
              schema:
                type: object`))
	})
	It("can handle interface slices", func() {
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"",
			[]interface{}{},
			[]interface{}{},
			[]interface{}{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    get:
      operationId: getUsers
      requestBody:
        content:
          */*:
            schema:
              type: array
      responses:
        '200':
          description: ok response
          content:
            application/json:
              schema:
                type: array
                items:
        'default':
          description: error response
          content:
            application/json:
              schema:
                type: array
                items:`))
	})
	It("can handle subtypes of maps", func() {
		type submap map[string]interface{}
		sw.DefineDataType(submap{}, sashay.SimpleDataTyper("object", ""))
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"",
			submap{},
			submap{},
			submap{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    get:
      operationId: getUsers
      requestBody:
        content:
          */*:
            schema:
              type: object
      responses:
        '200':
          description: ok response
          content:
            application/json:
              schema:
                type: object
        'default':
          description: error response
          content:
            application/json:
              schema:
                type: object`))
	})
	It("can handle nested generic objects", func() {
		type t struct {
			Map      map[string]interface{}   `json:"map"`
			Slice    []interface{}            `json:"slice"`
			SliceMap []map[string]interface{} `json:"slicemap"`
		}
		sw.Add(sashay.NewOperation(
			"GET",
			"/users",
			"",
			t{},
			t{},
			t{},
		))
		Expect(sw.BuildYAML()).To(ContainSubstring(`paths:
  /users:
    get:
      operationId: getUsers
      responses:
        '200':
          description: ok response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/t'
        'default':
          description: error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/t'
components:
  schemas:
    t:
      type: object
      properties:
        map:
          type: object
        slice:
          type: array
          items:
        slicemap:
          type: array
          items:
            type: object`))
	})
})
