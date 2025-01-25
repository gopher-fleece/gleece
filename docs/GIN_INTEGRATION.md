# Gin & Gleece Integration
If you are using the Gin framework for your HTTP routes, you can easily integrate Gleece by following these steps:

1. **Generate Routes File**:  
   - Gleece will generate a routes file from your annotated controllers. For example, it might generate `generated_routes.go`.

2. **Import and Register Routes**:  
   - In your `main.go` file, import the generated routes file and call the `RegisterRoutes` function to register the routes with Gin.

Here's an example:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gopher-fleece/gleece/routes" // Import the generated routes file
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