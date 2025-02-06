# Gin & Gleece Integration
If you are using the [Gin](https://github.com/gin-gonic/gin) framework for your HTTP routes, you can easily integrate Gleece by following these steps:

1. **Configure Gin as the engine**
   - In the Gleece configuration (usually `gleece.config.json`) make sure the `routesConfig->engine` is `gin`.

2. **Configure security function**
   - In the Gleece configuration set the full package path `routesConfig->authorizationConfig->authFileFullPackageName` (e.g. `github.com/gopher-fleece/gleece/security`).

3. **Generate Routes File**:  
   - Gleece will generate a routes file from your annotated controllers. For example, it might generate `generated_routes.go`.

4. **Import and Register Routes**:  
   - In your `main.go` file, import the generated routes file and call the `RegisterRoutes` function to register the routes with Gin.

Here's an example:

### Server
```go
package main

import (
    "github.com/gin-gonic/gin"
    "<package>/routes" // Import the generated routes file
)

func main() {
    // Create a Gin router
    router := gin.Default()

    // Register Gleece routes
    routes.RegisterRoutes(router)

    // Start the server
    router.Run(":8080")
}
```

#### Authentication function
```go
package security

import (
	"github.com/gin-gonic/gin"
	"github.com/gopher-fleece/gleece/runtime"
)

func GleeceRequestAuthorization(ctx *gin.Context, check runtime.SecurityCheck) *runtime.SecurityError {
	return nil
}
```
