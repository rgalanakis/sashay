package sashay_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rgalanakis/sashay"
)

func ExampleSashay_basicParameters() {
	sw := sashay.New("t", "d", "v")
	op := sashay.NewOperation(
		"POST",
		"/users/:id",
		"Update the user.",
		struct {
			ID         int    `path:"id" validate:"min=1"`
			Pretty     bool   `query:"pretty" description:"If true, return pretty-printed JSON." default:"true"`
			NoResponse bool   `header:"X-NO-RESPONSE" description:"If true, return a 204 rather than the updated User."`
			Name       string `json:"name"`
		}{},
		nil,
		nil,
	)
	sw.Add(op)
	fmt.Println(sw.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: t
	//   description: d
	//   version: v
	// paths:
	//   /users/{id}:
	//     post:
	//       operationId: postUsersId
	//       summary: Update the user.
	//       parameters:
	//         - name: id
	//           in: path
	//           required: true
	//           schema:
	//             type: integer
	//             format: int64
	//         - name: pretty
	//           in: query
	//           description: If true, return pretty-printed JSON.
	//           schema:
	//             type: boolean
	//             default: true
	//         - name: X-NO-RESPONSE
	//           in: header
	//           description: If true, return a 204 rather than the updated User.
	//           schema:
	//             type: boolean
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
	//         '204':
	//           description: The operation completed successfully.
	//         'default':
	//           description: error response
}

func validate(_ interface{}) error { return nil }

func ExampleSashay_usableParams() {
	sw := sashay.New("t", "d", "v")

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
	getUsersHandler := func(w http.ResponseWriter, r *http.Request) {
		params := getUsersParams{Status: r.URL.Query().Get("status")}
		if err := validate(params); err != nil {
			w.WriteHeader(500)
		} else {
			var users []User
			// Logic to get users
			bytes, _ := json.Marshal(users)
			w.Write(bytes)
			w.WriteHeader(200)
		}
	}

	sw.Add(getUsersOp)
	http.HandleFunc(getUsersOp.Path, getUsersHandler)
}

func ExampleSashay_nestedParams() {
	sw := sashay.New("t", "d", "v")
	op := sashay.NewOperation(
		"POST",
		"/users",
		"Create a user.",
		struct {
			Name struct {
				First string `json:"first"`
				Last  string `json:"last"`
			} `json:"name"`
		}{},
		nil,
		nil,
	)
	sw.Add(op)
	fmt.Println(sw.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: t
	//   description: d
	//   version: v
	// paths:
	//   /users:
	//     post:
	//       operationId: postUsers
	//       summary: Create a user.
	//       requestBody:
	//         required: true
	//         content:
	//           application/json:
	//             schema:
	//               type: object
	//               properties:
	//                 name:
	//                   type: object
	//                   properties:
	//                     first:
	//                       type: string
	//                     last:
	//                       type: string
	//       responses:
	//         '204':
	//           description: The operation completed successfully.
	//         'default':
	//           description: error response
}

func ExampleSashay_basicResponse() {
	sw := sashay.New("t", "d", "v")
	op := sashay.NewOperation(
		"GET",
		"/users",
		"",
		nil,
		[]User{},
		ErrorModel{},
	)
	sw.Add(op)
	fmt.Println(sw.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: t
	//   description: d
	//   version: v
	// paths:
	//   /users:
	//     get:
	//       operationId: getUsers
	//       responses:
	//         '200':
	//           description: ok response
	//           content:
	//             application/json:
	//               schema:
	//                 type: array
	//                 items:
	//                   $ref: '#/components/schemas/User'
	//         'default':
	//           description: error response
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/ErrorModel'
	// components:
	//   schemas:
	//     ErrorModel:
	//       type: object
	//       properties:
	//         error:
	//           type: object
	//           properties:
	//             message:
	//               type: string
	//             code:
	//               type: integer
	//               format: int64
	//     User:
	//       type: object
	//       properties:
	//         result:
	//           type: object
	//           properties:
	//             id:
	//               type: integer
	//               format: int64
	//             name:
	//               type: string
}

func ExampleSashay_advancedResponses() {
	sw := sashay.New("t", "d", "v")

	type TeapotResponse struct {
		Probability float64 `json:"prob"`
	}
	type TeapotError struct {
		Strength float64 `json:"strength"`
	}
	op := sashay.NewOperation(
		"GET",
		"/is_teapot",
		"Error if the server is a teapot.",
		nil,
		sashay.Responses{
			sashay.NewResponse(200, "Not a teapot.", TeapotResponse{}),
			sashay.NewResponse(203, "I may not be a teapot", TeapotResponse{}),
		},
		sashay.NewResponse(418, "Yes, I am sure a teapot!", TeapotError{}),
	)

	sw.Add(op)
	fmt.Println(sw.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: t
	//   description: d
	//   version: v
	// paths:
	//   /is_teapot:
	//     get:
	//       operationId: getIsTeapot
	//       summary: Error if the server is a teapot.
	//       responses:
	//         '200':
	//           description: Not a teapot.
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/TeapotResponse'
	//         '203':
	//           description: I may not be a teapot
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/TeapotResponse'
	//         '418':
	//           description: Yes, I am sure a teapot!
	//           content:
	//             application/json:
	//               schema:
	//                 $ref: '#/components/schemas/TeapotError'
	// components:
	//   schemas:
	//     TeapotError:
	//       type: object
	//       properties:
	//         strength:
	//           type: number
	//           format: double
	//     TeapotResponse:
	//       type: object
	//       properties:
	//         prob:
	//           type: number
	//           format: double
}

func ExampleSashay_customDataType() {
	sw := sashay.New("t", "d", "v")

	type UnitOfTime struct {
		time time.Time
		unit string
	}

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

	extractEnum := func(f sashay.Field, of sashay.ObjectFields) {
		of["type"] = "string"
		if enum := f.StructField.Tag.Get("enum"); enum != "" {
			values := strings.Split(enum, "|")
			of["enum"] = fmt.Sprintf("['%s']", strings.Join(values, "', '"))
		}
	}
	sw.DefineDataType("", sashay.BuiltinDataTyperFor("", extractEnum))

	sw.Add(sashay.NewOperation(
		"POST",
		"/stuff",
		"Update stuff.",
		struct {
			StartMonth UnitOfTime `json:"startMonth" timeunit:"month"`
			EndDay     UnitOfTime `json:"endDay" timeunit:"date"`
			Status     string     `json:"status" enum:"on|off"`
		}{},
		nil,
		nil,
	))
	fmt.Println(sw.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: t
	//   description: d
	//   version: v
	// paths:
	//   /stuff:
	//     post:
	//       operationId: postStuff
	//       summary: Update stuff.
	//       requestBody:
	//         required: true
	//         content:
	//           application/json:
	//             schema:
	//               type: object
	//               properties:
	//                 startMonth:
	//                   type: string
	//                   format: YYYY-MM
	//                 endDay:
	//                   type: string
	//                   format: date
	//                 status:
	//                   type: string
	//                   enum: ['on', 'off']
	//       responses:
	//         '204':
	//           description: The operation completed successfully.
	//         'default':
	//           description: error response

}

func ExampleSelectMap() {
	sw := sashay.New("t", "d", "v")
	// We can remove "/internal" routes
	sw.Add(sashay.NewOperation("GET", "/internal/users", "", nil, nil, nil))
	// And lowercase all paths
	sw.Add(sashay.NewOperation("GET", "/USeRS", "", nil, nil, nil))
	result := sashay.SelectMap(sw, func(op sashay.Operation) *sashay.Operation {
		if strings.Contains(op.Path, "/internal") {
			return nil
		}
		op.Path = strings.ToLower(op.Path)
		return &op
	})
	fmt.Println(result.BuildYAML())
	// Output:
	// openapi: 3.0.0
	// info:
	//   title: t
	//   description: d
	//   version: v
	// paths:
	//   /users:
	//     get:
	//       operationId: getUsers
	//       responses:
	//         '204':
	//           description: The operation completed successfully.
	//         'default':
	//           description: error response
}
