# Chi & Gleece Integration
If you are using the [Chi v5](https://github.com/go-chi/chi) framework for your HTTP routes, you can easily integrate Gleece by following these steps:

1. **Configure Chi as the engine**
   - In the Gleece configuration (usually `gleece.config.json`) set the `routesConfig->engine` to `chi`.

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
    "net/http"

    "github.com/go-chi/chi/v5"
    "<package>/routes" // Import the generated routes file
)

func main() {
    // Create a new Chi router
    r := chi.NewRouter()

    // Register Gleece routes
    routes.RegisterRoutes(r)

    // Start the server on port 8080
    http.ListenAndServe(":8080", r)
}
```

#### Authentication function
```go
package security

import (
	"net/http"

	"github.com/gopher-fleece/gleece/runtime"
)

func GleeceRequestAuthorization(r *http.Request, check runtime.SecurityCheck) *runtime.SecurityError {
	return nil
}
```