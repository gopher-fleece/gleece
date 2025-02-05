# Gleece Middlewares

Gleece supports middlewares.

The event middleware can be triggered:
- Before running the operation/function: `beforeOperation`
- After the operation returns without error: `afterOperationSuccess`
- After the operation returns with error: `onError`

The middleware receives the router's context (depending on the engine in use), and in the case of `onError`, it also receives the error instance.

The middleware returns a boolean indicating whether to continue the execution or to abort it (usually when the response is handled inside the middleware function).

## Implement Middleware

```go
// For Gin
func MyMiddleware(ctx *gin.Context) bool {
	return true
}

// For echo
func MyMiddleware(ctx echo.Context) bool {
	return true
}

// For Fiber
func MyMiddleware(ctx *fiber.Ctx) bool {
	return true
}

// For Gorilla Mux & Chi
func MyMiddleware(w http.ResponseWriter, r *http.Request) bool {
	return true
}
```

And similarly for `onError` middlewares:
```go
// For Gin
func MyErrorMiddleware(ctx *gin.Context, err error) bool {
	return true
}

// For echo
func MyErrorMiddleware(c echo.Context, err error) bool {
	return true
}

// For Gorilla Mux
func MyErrorMiddleware(w http.ResponseWriter, r *http.Request, err error) bool {
	return true
}

// For Fiber
func MyMiddleware(ctx *fiber.Ctx, err error) bool {
	return true
}
```

## Declare Middleware

In the `gleece.config.json`, set the `routesConfig->middlewares` with an array of middlewares.

Each middleware should contain the package from where to import, the middleware function name, and when to execute it.

For example:
```json
...
"middlewares": [
		{
			"fullPackageName": "github.com/gopher-fleece/gleece/middlewares",
			"execution": "beforeOperation",
			"functionName": "MiddlewareBeforeOperation"
		},
		{
			"fullPackageName": "github.com/gopher-fleece/gleece/middlewares",
			"execution": "afterOperationSuccess",
			"functionName": "MiddlewareAfterOperationSuccess"
		},
		{
			"fullPackageName": "github.com/gopher-fleece/gleece/middlewares",
			"execution": "onError",
			"functionName": "MiddlewareOnError"
		}
	]
...
```

There is an unlimited number of middlewares, and they are executed in the order specified in the configuration.

Aborting execution (returning `false`) will stop the execution of the next (if any) middlewares as well.
