**Gleece** - Bringing joy and ease to API development in Go! 🚀   


[![gleece](https://github.com/gopher-fleece/gleece/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/gopher-fleece/gleece/actions/workflows/build.yml)
[![Latest Release](https://img.shields.io/github/v/release/gopher-fleece/gleece)](https://github.com/gopher-fleece/gleece/releases)
[![Coverage Status](https://coveralls.io/repos/github/gopher-fleece/gleece/badge.svg?branch=main)](https://coveralls.io/github/gopher-fleece/gleece?branch=main)

[![GitHub stars](https://img.shields.io/github/stars/gopher-fleece/gleece.svg?style=social&label=Stars)](https://github.com/gopher-fleece/gleece/stargazers) 
[![License](https://img.shields.io/github/license/gopher-fleece/gleece.svg?style=social)](https://github.com/gopher-fleece/gleece/blob/master/LICENSE)

[![Go Reference](https://pkg.go.dev/badge/github.com/gopher-fleece/gleece.svg)](https://pkg.go.dev/github.com/gopher-fleece/gleece)


---

## Philosophy  
Developing APIs doesn’t have to be a chore - it should be simple, efficient, and enjoyable.  

Gone are the days of manually writing repetitive boilerplate code or struggling to keep your API documentation in sync with your implementation.

With Gleece, you can:  
- 🔧 **Simplify** your API development process.  
- 📜 Automatically **generate OpenAPI specs** directly from your code.  
- 🎯 Ensure your APIs are always **well-documented and consistent**.  
- ✅ **Validate input data** effortlessly to keep your APIs robust and secure.  

Gleece aims to make Go developers’ lives easier by seamlessly integrating API routes, validation, and documentation into a single cohesive workflow.

## 🚀 Usage Example  

Here’s a practical example of how Gleece simplifies your API development:  


```go
package api

import (
	"github.com/google/uuid"
	"github.com/gopher-fleece/gleece/ctrl" // Importing GleeceController
)

// @Tag(User Management)
// @Route(/users-management)
// @Description The User Management API
type UserController struct {
	ctrl.GleeceController // Embedding the GleeceController
}

// @Description User's domicile
type Domicile struct {
	// @Description The address
	Address string `json:"address" validate:"required"`
	// @Description The number of the house (must be at least 1)
	HouseNumber int `json:"houseNumber" validate:"gte=1"`
}


// @Description Create a new user
// @Method(POST)
// @Route(/user/{user_name})
// @Path(name, { name: "user_name", validate: "require" }) The user's name
// @Query(email, { validate: "required,email" }) The user's email
// @Body(domicile, { validate: "required" }) The user's domicile info
// @Header(origin, { name: "x-origin" }) The request origin
// @Header(trace) The trace info
// @Response(200) The ID of the newly created user
// @ErrorResponse(500) The error when process failed
// @Security(ApiKeyAuth, { scopes: ["read:users", "write:users"] })
func (ec *UserController) CreateNewUser(email string, name string, domicile Domicile, origin string, trace string) (string, error) {
	// Do the logic....
	userId := uuid.New()
	return userId.String(), nil
}
```
### What’s Happening Here?  

1. **Annotations**:  
   - Gleece uses annotations to automatically generate routes & OpenAPI documentation. The format for the annotations is:

   ```go
   // {{ annotation }} {{ ( {{theBasicRequiredParam}}, { json5 attributes } ) }} {{ description }}
   ```

   - **Annotation**: This is always required and specifies the type of annotation, such as `@Tag`, `@Route`, `@Method`, etc.
   - **theBasicRequiredParam**: This parameter is required depending on the type of annotation. For example, `@Route` requires a route path, `@Method` requires an HTTP method, etc.
   - **json5 attributes**: These are optional attributes in JSON5 format that provide additional information for the annotation. For example, in `@Header(origin, { name: x-origin })`, `{ name: x-origin }` is a JSON5 attribute specifying the name of the header in the http request, while the `origin` is the name of the parameter in the given function.
   - **Description**: This is an optional description that provides further details about the annotation. It helps in generating more descriptive documentation.

   See full supported [annotations](./docs/ANNOTATIONS.md) documentation


2. **Validation Handled by Gleece**:  
   - Input validation is simplified by Gleece using [go-playground/validator](https://github.com/go-playground/validator) format.  
   - You define validation rules directly on your struct fields:  
     - `validate:"required"` ensures the `Address` field is mandatory.  
     - `validate:"gte=1"` ensures the `HouseNumber` field has a value of at least 1.  
   - Gleece processes these validation rules automatically during request handling and returns 422 in case of not passing validation.  

3. **Controllers**:  
   - Simply embed the `GleeceController` (imported from `github.com/gopher-fleece/gleece/ctrl`) into your own controllers to gain its functionality.  

4. **Automation**:  
   - No manual steps required — your OpenAPI spec is ready to go!  

## 🌐 Integrating with Gin  

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
---

## 🚧 Disclaimer  
⚠️⚠️⚠️ **Work in Progress**  
Gleece is currently an under-development project.  We’re working hard to make it amazing.

We’d love your feedback and contributions as we shape Gleece together!

Stay tuned for updates, and feel free to open issues or pull requests to help us improve!  

---

## 📜 License  
Gleece is licensed under the **MIT License**. 📄 You are free to use, modify, and distribute it with attribution. See the [`LICENSE`](./LICENSE) file for details.  

---

