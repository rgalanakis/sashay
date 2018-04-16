package sashay_test

import (
	"strings"
	"testing"

	"time"

	"bytes"

	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rgalanakis/sashay"
)

func TestSwagger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SwaggerGen Suite")
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

	It("can write to a custom buffer", func() {
		b := bytes.NewBuffer([]byte("hello there!\n"))
		sw.WriteYAML(b)
		Expect(b.String()).To(ContainSubstring(`hello there!
openapi: 3.0.0
info:
`))
	})
})
