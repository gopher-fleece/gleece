# Gleece Middlewares

Gleece provides middleware support for request processing.

Middleware events can be triggered at three points:
- Before the operation/function execution: `beforeOperation`
- After successful operation completion: `afterOperationSuccess`
- After operation failure: `onOperationError`

Each middleware receives the router's context (specific to the engine in use). For `onOperationError` events, the middleware also receives the error instance.

The middleware returns a boolean value that determines whether to continue execution (`true`) or abort it (`false`). Typically, execution is aborted when the response is handled within the middleware function.

## Implementing Middleware

```go
// For Gin
func MyMiddleware(ctx *gin.Context) bool {
	return true
}

// For Echo
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

And similarly for errors middlewares:
```go
// For Gin
func MyErrorMiddleware(ctx *gin.Context, err error) bool {
	return true
}

// For Echo
func MyErrorMiddleware(c echo.Context, err error) bool {
	return true
}

// For Gorilla Mux & Chi
func MyErrorMiddleware(w http.ResponseWriter, r *http.Request, err error) bool {
	return true
}

// For Fiber
func MyErrorMiddleware(ctx *fiber.Ctx, err error) bool {
	return true
}
```

## Registering Middleware

Use the generated code's `RegisterMiddleware` / `RegisterErrorMiddleware` API to register your middleware functions.

Example:
```go
package main

import (
    "net/http"

    gleeceRoutes "<package>"

    "github.com/gopher-fleece/runtime"
    "github.com/gin-gonic/gin"
)

func LogBeforeOperationMiddleware(ctx *gin.Context) bool {
	println("Method:", ctx.Request.Method, "Path:", ctx.Request.URL.Path, "arrived")
	return true
}

func LogAfterOperationSuccessMiddleware(ctx *gin.Context) bool {
	println("Method:", ctx.Request.Method, "Path:", ctx.Request.URL.Path, "completed")
	return true
}

func LogOnErrorMiddleware(ctx *gin.Context, err error) bool {
	println("Method:", ctx.Request.Method, "Path:", ctx.Request.URL.Path, "failed with error:", err.Error())
	return true
}

func main() {
    // Create a default Gin router
    router := gin.Default()

    // Register the middlewares
    gleeceRoutes.RegisterMiddleware(runtime.BeforeOperation, LogBeforeOperationMiddleware)
    gleeceRoutes.RegisterMiddleware(runtime.AfterOperationSuccess, LogAfterOperationSuccessMiddleware)
    gleeceRoutes.RegisterErrorMiddleware(runtime.OnOperationError, LogOnErrorMiddleware)

    // Register the routes from the generated code
    gleeceRoutes.RegisterRoutes(router)

    // Start the server on port 8080
    router.Run("127.0.0.1:8080")
}
```

You can register any number of middleware functions. They will be executed in the order of registration.

When a middleware returns `false`, it halts the execution chain, preventing any remaining middleware from running.
