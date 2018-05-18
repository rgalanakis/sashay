package sashay

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

// Operation is the definition for an endpoint (method and path).
// See https://swagger.io/specification/#operationObject
type Operation struct {
	// Method is a string like GET, POST, etc.
	Method string
	// Path is the path for the endpoint. Parameters should have a leading colon, like /users/:id.
	Path string
	// Summary is the summary for the endpoint. Please provide it.
	Summary string
	// Description is an optional longer endpoint description with Markdown support.
	Description string
	// Params is a zero'd instance of parameters for the endpoint.
	// If there are no params, use nil.
	Params interface{}
	// ReturnOk is a zero'ed instance of the struct used for successful responses from the endpoint.
	// If nil, assume a 204 success and use no body.
	ReturnOk interface{}
	// ReturnOk is a zero'ed instance of the struct used for an error response from the endpoint.
	// Since all endpoints should return the same error response shape,
	// we use thue 'default' Swagger response field. We can add custom error code mapping in the future.
	ReturnErr interface{}
	// Tags is a slice of string tags for the operation.
	// Tags can be used for logical grouping of operations by resources or any other qualifier.
	Tags []string
}

// WithDescription sets the description on the receiver and returns a modified instance.
func (op Operation) WithDescription(desc string) Operation {
	op.Description = desc
	return op
}

// AddTags appends new tags to the receiver and returns a modified instance.
func (op Operation) AddTags(tags ...string) Operation {
	op.Tags = append(op.Tags, tags...)
	return op
}

func (op Operation) toInternalOperation() internalOperation {
	return internalOperation{
		op,
		NewMethod(op.Method),
		NewPath(op.Path),
		NewOperationID(op),
		op.Summary,
		op.Description,
		NewField(op.Params),
		op.responses(),
		op.Tags,
	}
}

func (op Operation) responses() Responses {
	responses := make(Responses, 0, 2)

	switch returnOk := op.ReturnOk.(type) {
	case Responses:
		responses = append(responses, returnOk...)
	case Response:
		responses = append(responses, returnOk)
	default:
		method := NewMethod(op.Method)
		// Default success code is 204 if there is no ReturnOK (empty body),
		// 201 for POST, 200 otherwise.
		var code int
		var desc string
		if op.ReturnOk == nil {
			code = 204
			desc = "The operation completed successfully."
		} else if method == "post" {
			code = 201
			desc = "ok response"
		} else {
			code = 200
			desc = "ok response"
		}
		responses = append(responses, NewResponse(code, desc, op.ReturnOk))
	}

	switch returnErr := op.ReturnErr.(type) {
	case Responses:
		responses = append(responses, returnErr...)
	case Response:
		responses = append(responses, returnErr)
	default:
		responses = append(responses, NewResponse(-1, "error response", op.ReturnErr))
	}

	return responses
}

// NewOperation returns a new Operation instance with the given parameters.
func NewOperation(method, path, summary string, params, returnOK, returnErr interface{}) Operation {
	return Operation{
		Method:    method,
		Path:      path,
		Summary:   summary,
		Params:    params,
		ReturnOk:  returnOK,
		ReturnErr: returnErr,
	}
}

// Response defines a Swagger response, roughly corresponding to
// https://swagger.io/specification/#responseObject
// (see also https://swagger.io/docs/specification/describing-responses/)
// Clients can use Responses and Response when they need to override the default behavior.
type Response struct {
	Code        string
	Description string
	Field       Field
}

// NewResponse returns a new Response initialized with the given code and description.
// code is an HTTP status code, or -1 for "default".
// shape should be the return object, like what is passed as NewOperation's
// returnOK or returnErr argument (something like User{} or ErrorResponse{}).
func NewResponse(code int, description string, shape interface{}) Response {
	var strcode string
	if code == -1 {
		strcode = "default"
	} else {
		strcode = strconv.Itoa(code)
	}
	return Response{strcode, description, NewField(shape)}
}

// Responses is a slice of Response objects.
type Responses []Response

// Method represents an HTTP method string ("get", "post", etc.).
type Method string

// NewMethod returns a Method with the right casing/form from an HTTP verb string.
// "get" => "get"
// "POST" => "post"
func NewMethod(s string) Method {
	return Method(strings.ToLower(s))
}

// Path represents a Swagger path, like "/users/{id}/pets".
type Path string

// NewPath returns a Path with the right form from a Go-style route.
// "/users/:id" => "/users/{id}"
func NewPath(s string) Path {
	return Path(swaggerPathReplace.ReplaceAllString(s, "/{$1}"))
}

var swaggerPathReplace = regexp.MustCompile("/:([A-Za-z0-9]+)")

// OperationID represents a Swagger operationId string.
type OperationID string

// NewOperationID returns an OperationID that is unique for the method and endpoint.
func NewOperationID(op Operation) OperationID {
	bu := bytes.NewBuffer(nil)
	bu.WriteString(strings.ToLower(op.Method))

	path := op.Path
	path = strings.Replace(path, "/", "_", -1)
	path = strings.Replace(path, "-", "_", -1)
	path = operationIDPathClean.ReplaceAllString(path, "")
	path = strings.Trim(path, "_")

	for _, piece := range strings.Split(path, "_") {
		bu.WriteString(strings.ToUpper(piece[0:1]))
		bu.WriteString(piece[1:])
	}

	return OperationID(bu.String())
}

var operationIDPathClean = regexp.MustCompile("[^A-Za-z0-9_]")

// internalOperation wraps stuff in Field and Responses so we don't have to do it inline,
// and can use consistent interfaces in our internal code.
type internalOperation struct {
	Original    Operation
	Method      Method
	Path        Path
	OperationID OperationID
	Summary     string
	Description string
	Params      Field
	Responses   Responses
	Tags        []string
}

// True if a requestBody section is needed for the object.
// POST and PUT operations should get this section if any params are defined,
// otherwise it should be false (GET, DELETE etc should never use request bodies).
func (o internalOperation) useRequestBody() bool {
	return (o.Method == "post" || o.Method == "put") && !o.Params.Nil()
}
