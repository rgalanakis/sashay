[![Build Status](https://travis-ci.org/rgalanakis/sashay.svg?branch=master)](https://travis-ci.org/rgalanakis/sashay)
[![codecov](https://codecov.io/gh/rgalanakis/sashay/branch/master/graph/badge.svg)](https://codecov.io/gh/rgalanakis/sashay)

[![GoDoc](https://godoc.org/github.com/rgalanakis/sashay?status.svg)](http://godoc.org/github.com/rgalanakis/sashay)
[![license](http://img.shields.io/badge/license-MIT-orange.svg)](https://raw.githubusercontent.com/rgalanakis/sashay/master/LICENSE)

# Sashay

Sashay allows you to generate OpenAPI 3.0 (Swagger) files using executable Go code,
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
For example, given the following code:

```go
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
```

You can generate the following YAML:

```yaml
openapi: 3.0.0
info:
  title: PetStore API
  description: Manage your pet store with our API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: getPets
      summary: Return all pets.
      parameters:
        - name: status
          in: query
          schema:
            type: string
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
    post:
      operationId: postPets
      summary: Create a pet.
      parameters:
        - name: pretty
          in: query
          description: If true, return pretty-printed JSON.
          schema:
            type: boolean
            default: true
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
                $ref: '#/components/schemas/Pet'
        'default':
          description: error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /pets/{id}:
    get:
      operationId: getPetsId
      summary: Fetch info about a pet.
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: ok response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
        'default':
          description: error response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
    Pet:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
```

See the [documentation at godoc.org] for more information and tutorials about how to use Sashay.
See [https://swagger.io/specification/](https://swagger.io/specification/) for more info about the spec.
