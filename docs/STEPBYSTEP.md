## Step by Step Gin REST Server Example

Assuming you already have a REST server using Gin, a simple server might look like this:
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

## Adding Gleece Package

To add the `Gleece` external package, run the following command:

```bash
go get github.com/gopher-fleece/gleece/external
```

Install the `Gleece` CLI:

```bash
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
		"info": {
			"title": "Sample API",
			"description": "This is a sample API",
			"termsOfService": "http://example.com/terms/",
			"contact": {
				"name": "API Support",
				"url": "http://www.example.com/support",
				"email": "support@example.com"
			},
			"license": {
				"name": "Apache 2.0",
				"url": "http://www.apache.org/licenses/LICENSE-2.0.html"
			},
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
		"defaultSecurity": [
			{
				"securityMethod": [
					{
						"name": "securitySchemaName",
						"scopes": [
							"read",
							"write"
						]
					}
				]
			}
		],
		"specGeneratorConfig": {
			"outputPath": "./dist/swagger.json"
		}
	}
}
```

## Authentication Function

Create a module and Go file for the authentication function, assuming it is `auth/security.go`. Set the package path where your security check function is located in the `routesConfig->authorizationConfig->authFileFullPackageName`.

Inside the file, paste the following code. Modify the logic in `GleeceRequestAuthorization` to fit your needs:

```go
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/external"
)

func GleeceRequestAuthorization(ctx *gin.Context, check external.SecurityCheck) *external.SecurityError {

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "change that condition...." {
		return &external.SecurityError{
			StatusCode: 403,
			Message:    "You are not authorized to read that API",
		}
	}
	return nil
}
```

## Creating Controllers

Create the controller `controllers/user.ctrl.go`.

First, create a struct and embed the `GleeceController` in it. Then add a route method to the struct with the Gleece annotations:

```go
package controllers

import (
	"github.com/gopher-fleece/gleece/external"
)

// UsersController
// @Tag(Users) Users
// @Route(/users)
// @Description The Users API
type UsersController struct {
	external.GleeceController // Embedding the GleeceController to inherit its methods
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

## Running the Gleece Generator

Now you are ready to the Gleece generator.

Run the Gleece generator in your terminal:
```bash
gleece
```

By default, it will read the `gleece.config.json` configuration and generate the Gin routes and OpenAPI3 specification file into the `dist` directory.

## Integrating Generated Routes

Import the newly created routes into your `main.go` module and register the Gin instance to the generated code:

```go
package main

import (
    "net/http"

    gleeceRoutes "example-prog/dist"

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

Now you can run your app and see that everything works like a charm :)
---

Further Reading:
- [Annotations & Options](./ANNOTATIONS.md)
- [Custom Validations](./CUSTOM_VALIDATION.md) 
- [Error handling](./ERROR_HANDLING.md)
- [Error handling](./ERROR_HANDLING.md)
