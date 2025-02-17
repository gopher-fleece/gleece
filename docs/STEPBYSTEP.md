## Gleece Step by Step Instructions

Assuming you already have a REST server, a simple server might look like this:

Example in `Gin`:

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func main() {
    // Create a default Gin router
    router := gin.Default()

    // Define a route for GET /hello
    router.GET("/hello", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Hello World!",
        })
    })

    // Start the server on port 8080
    router.Run("127.0.0.1:8080")
}
```

## Importing Gleece

Add Gleece's [runtime](https://github.com/gopher-fleece/runtime) package to the app's codebase:

```bash
go get github.com/gopher-fleece/runtime
```

And install the `Gleece` CLI:

```bash
go get github.com/gopher-fleece/gleece
go install github.com/gopher-fleece/gleece
```


## Basic Gleece Configuration

Create a basic `Gleece` configuration file named `gleece.config.json`:
```json
{
	"commonConfig": {
		"controllerGlobs": [
			"./*.go",
			"./**/*.go"
		]
	},
	"routesConfig": {
		"engine": "gin",
		"outputPath": "./dist/gleece.go",
		"outputFilePerms": "0644",
		"authorizationConfig": {
			"authFileFullPackageName": "",
			"enforceSecurityOnAllRoutes": true
		}
	},
	"openAPIGeneratorConfig": {
		"openapi" : "3.0.0",
		"info": {
			"title": "Sample API",
			"description": "This is a sample API",
			"termsOfService": "http://example.com/terms/",
			"version": "1.0.0"
		},
		"baseUrl": "https://api.example.com",
		"securitySchemes": [
			{
				"description": "API Key for accessing the API",
				"name": "securitySchemaName",
				"fieldName": "x-header-name",
				"type": "apiKey",
				"in": "header"
			}
		],
		"defaultSecurity": {
			"name": "securitySchemaName",
			"scopes": [
				"read",
				"write"
			]
		},
		"specGeneratorConfig": {
			"outputPath": "./dist/swagger.json"
		}
	}
}
```

See the full [configuration](./CONFIG.md) documentation for detailed explanation of all available properties.

## Authentication Function

Create a module and Go file for the authentication function, assuming it is `auth/security.go`. Set the package path where your security check function is located in the `routesConfig->authorizationConfig->authFileFullPackageName`.

Inside the file, paste the following code. Modify the logic in `GleeceRequestAuthorization` to fit your needs:

```go
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/runtime"
)

func GleeceRequestAuthorization(ctx *gin.Context, check runtime.SecurityCheck) *runtime.SecurityError {

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "change that condition...." {
		return &runtime.SecurityError{
			StatusCode: 403,
			Message:    "You are not authorized to read that API",
		}
	}
	return nil
}
```

Note: This example is for the `Gin` engine. See [Integrating with Golang Rest Routers](#Integrating-with-Golang-Rest-Routers) for instructions on integrating authorization with other engines.

## Creating Controllers

Create the controller `controllers/user.ctrl.go`.

First, create a struct and embed the `GleeceController` in it. Then add a route method to the struct with the Gleece annotations:

```go
package controllers

import (
	"github.com/gopher-fleece/runtime"
)

// UsersController
// @Tag(Users) Users
// @Route(/users)
// @Description The Users API
type UsersController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Description Get user data
// @Method(GET)
// @Route(/{id})
// @Path(id) The id of the user to get
// @Response(200) The user's information
// @ErrorResponse(404) The user not found
// @ErrorResponse(500) The error when process failed
// @Security(securitySchemaName, { scopes: ["read:users" ] }) Consumer should pass this security schema
func (ec *UsersController) GetUser(id string) (string, error) {
	return "", nil
}
```

Every route function must declare at least one return type: `error`. For operations without a response payload, `error` will be the only return type. For operations that return data, the response payload (`string`, `struct` etc.) must be the first return value, followed by error as the second return value.


## Running the Gleece Generator

Now you are ready to the Gleece generator.

Run the Gleece generator in your terminal:
```bash
gleece
```

By default, it reads the `gleece.config.json` configuration file.
It generates routes based on the engine specified in `routesConfig->engine`.
It also generates the OpenAPI specification according to the version set in `openAPIGeneratorConfig->openapi`.

In this example configuration, it generates `gin` routes and OpenAPI `v3.0.0` specifications, with both outputs in the `dist` directory.

## Integrating Generated Routes

Import the newly created routes into your `main.go` module and register the Gin instance to the generated code:

```go
package main

import (
    "net/http"

    gleeceRoutes "<package>/routes"

    "github.com/gin-gonic/gin"
)

func main() {
    // Create a default Gin router
    router := gin.Default()

    // Define a route for GET /hello
    router.GET("/hello", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Hello World!",
        })
    })

    // Register the routes from the generated code
    gleeceRoutes.RegisterRoutes(router)

    // Start the server on port 8080
    router.Run("127.0.0.1:8080")
}
```

And... start the application.

Happy Coding! 🤓

---

## Further Reading:
- [Annotations & Options](./ANNOTATIONS.md)
- [Authentication](./AUTHENTICATION.md)
- [Validations](./VALIDATION.md) 
- [Error handling](./ERROR_HANDLING.md)
- [Middlewares](./MIDDLEWARES.md)
- [Configuration](./CONFIG.md)
- [Advanced](./ADVANCED.md)

## Integrating with Golang Rest Routers 

- [Gin](./routers/GIN_INTEGRATION.md)
- [Echo](./routers/ECHO_INTEGRATION.md)
- [Gorilla Mux](./routers/MUX_INTEGRATION.md)
- [Chi](./routers/CHI_INTEGRATION.md)
- [Fiber](./routers/FIBER_INTEGRATION.md)
