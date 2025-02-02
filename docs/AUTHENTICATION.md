# Gleece authentication mechanism

Gleece handles security on both sides: the declaration of required security and scope, which is defined by the Security annotation, and the processing of the check itself through a supplied check function that will be called once a request arrives.

Let's dive into each part of the security mechanism.

## Security definition

Before any use, the security schema needs to be defined.

Set the security schema in the OpenAPI standard in the `openAPIGeneratorConfig->securitySchemes` in the `gleece.config.json` configuration.

For example
```json
...
"securitySchemes": [
			{
				"description": "API Key for accessing the API",
				"name": "securitySchemaName",
				"fieldName": "x-header-name",
				"type": "apiKey",
				"in": "header"
			},
			{
				"description": "API Key for accessing the API - 2",
				"name": "securitySchemaName2",
				"fieldName": "x-header-name-2",
				"type": "apiKey",
				"in": "header"
			}
		],
...
```

Properties are OpenAPI v3.x schema.

Later on, in the declarations, only those defined security schemas from the `name` property can be used.

## Security declaration

For each route, you can define a security annotation, which will specify the required authentication schema and scope for that route.

The security can be defined from the configuration, from the controller's class, and on each route function.

The lower the declaration level, it will OVERRIDE any decelerated security above it.

So you can define a default security in the config, override it and set a default security per controller class, and of course, you can override it and set security per route method.

#### Declare security in configuration
Set the default security in the `gleece.config.json` file.

Add your security to `openAPIGeneratorConfig->defaultSecurity` array.

For example:
```json
"defaultSecurity": [
			{
				"securityMethod": [
					{
						"name": "securitySchemaName",
						"scopes": [
							"read",
						]
					}
				]
			}
		],
```

#### Declare security in a controller

Add `Security` annotation to the struct/class declaration

For example:
```go
// @Route(/my-route)
// @Security(securitySchemaName, { scopes: ["read", "write"] })
type MyController struct {
	external.GleeceController
}
```
#### Declare security in a route

Add `Security` annotation to the function/route declaration

```go
// @Method(GET)
// @Route(/my-route-method)
// @Security(securitySchemaName, { scopes: ["admin"] })
func (ec *MyController) MyRouteMethod() (error) {
	return nil
}
```

At each level (configuration, class, and method), you can define more than one security set, and there will be an OR relationship between them. This means if one of them passes, the authentication is considered PASSED, even if other security checks failed.

For example:
```go
// @Method(GET)
// @Route(/my-route-method)
// @Security(securitySchemaName1, { scopes: ["admin"] })
// @Security(securitySchemaName2, { scopes: ["admin"] })
func (ec *MyController) MyRouteMethod() (error) {
	return nil
}
```

If `securitySchemaName2` check passes, even if `securitySchemaName1` check fails, the authorization will be considered as PASSED.

> We also highly recommend turning on the `routesConfig->authorizationConfig->enforceSecurityOnAllRoutes` flag to `true`. This will ensure at build time that there is no route without any security.


## Security Check

Each Gleece configuration **must** include a reference to the package where the security check is implemented in `routesConfig->authorizationConfig->authFileFullPackageName`.

For example:
```json
...
"authFileFullPackageName": "github.com/gopher-fleece/gleece/auth",
...
```

Implement your own security check in your package of choice.

Create a function with this signature, according to your router engine choice:

For `gin`:
```go
import (
	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/external"
)
func GleeceRequestAuthorization(ctx *gin.Context, check external.SecurityCheck) *external.SecurityError {
    return nil
}
```

For `echo`
```go
import (
	"github.com/gopher-fleece/gleece/external"
	"github.com/labstack/echo/v4"
)

func GleeceRequestAuthorization(ctx *gin.Context, check external.SecurityCheck) *external.SecurityError {
    return nil
}
```

For Gorilla's `Mux`
```go
import (
	"net/http"

	"github.com/gopher-fleece/gleece/external"
)

func GleeceRequestAuthorization(r *http.Request, check external.SecurityCheck) *external.SecurityError {
    return nil
}
```

The function will be called from the generated routes once a request arrives, with route's context and the defined security and scopes.

Implement your own check logic.

If it passes, return `nil`; otherwise, return `SecurityError`.

In case of more than one security definition on a route, the function will be called once per security, each time with the security schema name and scopes of that declaration. Gleece will throw an error only if ALL checks fail (as written, there is an OR relationship between security annotation declarations).

By default, the error will be formatted as [rfc7807](https://datatracker.ietf.org/doc/html/rfc7807) and you can override it and provide your own payload that will be responded with.

For example
```go
func GleeceRequestAuthorization(ctx *gin.Context, check external.SecurityCheck) *external.SecurityError {
    return &external.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: external.StatusUnauthorized,
	}
}
```
This will be returned with a 401 status code and the payload will be formatted as rfc7807.

And this:
```go
func GleeceRequestAuthorization(ctx *gin.Context, check external.SecurityCheck) *external.SecurityError {
    return &external.SecurityError{
			Message:    "Failed to authorize",
			StatusCode: external.StatusForbidden,
           	CustomError: &external.CustomError{
				Payload: struct {
					Message     string `json:"message"`
					Description string `json:"description"`
				}{
					Message:     "Custom error message",
					Description: "Custom error description",
				},
			},
	}
	
	return nil
}
```

Will be returned with a 403 status code and payload will be:

```json
{
    "message" : "Custom error message",
    "description": "Custom error description"
}
```