# Gorilla Mux & Gleece Integration
If you are using the [Gorilla Mux](https://github.com/gorilla/mux) framework for your HTTP routes, you can easily integrate Gleece by following these steps:

1. **Configure Mux as the engine**
   - In the Gleece configuration (usually `gleece.config.json`) set the `routesConfig->engine` to `mux`.

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

    "github.com/gorilla/mux"
    "<package>/routes" // Import the generated routes file
)

func main() {
    // Create a Mux router
    router := mux.NewRouter()

    // Register Gleece routes
    routes.RegisterRoutes(router)

    // Start the server
    http.ListenAndServe(":8080", router)
}
```

#### Authentication function
```go
package security

import (
	"net/http"

	"github.com/gopher-fleece/runtime"
)

func GleeceRequestAuthorization(r *http.Request, check runtime.SecurityCheck) *runtime.SecurityError {
	return nil
}
```