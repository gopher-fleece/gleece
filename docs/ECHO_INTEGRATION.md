# ECHO & Gleece Integration
If you are using the echo v4 framework for your HTTP routes, you can easily integrate Gleece by following these steps:

1. **Configure Echo as the engine**
   - In the Gleece configuration (usually `gleece.config.json`) set the `routesConfig->engine` to `echo`.

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
    "github.com/labstack/echo/v4"
    "github.com/gopher-fleece/gleece/routes" // Import the generated routes file
)

func main() {
    // Create a Echo router
    e := echo.New()

    // Register Gleece routes
    routes.RegisterRoutes(e)

    // Start the server
    e.Start(":8080")
}
```

#### Authentication function
```go
package security

import (
	"github.com/gopher-fleece/gleece/external"
   "github.com/labstack/echo/v4"
)

func GleeceRequestAuthorization(ctx echo.Context, check external.SecurityCheck) *external.SecurityError {
	return nil
}
```