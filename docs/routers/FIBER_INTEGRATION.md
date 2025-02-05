# Fiber & Gleece Integration
If you are using the [fiber v2](https://github.com/gofiber/fiber) framework for your HTTP routes, you can easily integrate Gleece by following these steps:

1. **Configure Fiber as the engine**
   - In the Gleece configuration (usually `gleece.config.json`) set the `routesConfig->engine` to `fiber`.

2. **Configure security function**
   - In the Gleece configuration set the full package path `routesConfig->authorizationConfig->authFileFullPackageName` (e.g. `github.com/gopher-fleece/gleece/security`).

3. **Generate Routes File**:  
   - Gleece will generate a routes file from your annotated controllers. For example, it might generate `generated_routes.go`.

4. **Import and Register Routes**:  
   - In your `main.go` file, import the generated routes file and call the `RegisterRoutes` function to register the routes with Echo.


Here's an example:

#### Server
```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gopher-fleece/gleece/routes" // Import the generated routes file
)

func main() {
    // Create a Fiber app
    app := fiber.New()

    // Register Gleece routes
    routes.RegisterRoutes(app)

    // Start the server
    app.Listen(":8080")
}
```

#### Authentication Function

```go
package security

import (
    "github.com/gopher-fleece/gleece/external"
    "github.com/gofiber/fiber/v2"
)

func GleeceRequestAuthorization(ctx *fiber.Ctx, check external.SecurityCheck) *external.SecurityError {
    return nil
}
```