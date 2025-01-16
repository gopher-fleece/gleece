# Gleece  


[![gleece](https://github.com/haimkastner/gleece/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/haimkastner/gleece/actions/workflows/build.yml)
[![Latest Release](https://img.shields.io/github/v/release/haimkastner/gleece)](https://github.com/haimkastner/gleece/releases)
[![GitHub stars](https://img.shields.io/github/stars/haimkastner/gleece.svg?style=social&label=Stars)](https://github.com/haimkastner/gleece/stargazers) 
[![License](https://img.shields.io/github/license/haimkastner/gleece.svg?style=social)](https://github.com/haimkastner/gleece/blob/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/haimkastner/gleece.svg)](https://pkg.go.dev/github.com/haimkastner/gleece)
[![Coverage Status](https://coveralls.io/repos/github/haimkastner/gleece/badge.svg?branch=main)](https://coveralls.io/github/haimkastner/gleece?branch=main)
🎉 **Gleece** - Bringing joy and ease to API development in Go! 🚀  
---

## 🌟 Philosophy  
Developing APIs doesn’t have to be a chore - it should be simple, efficient, and enjoyable. 💡✨  

Gone are the days of manually writing repetitive boilerplate code or struggling to keep your API documentation in sync with your implementation. 🚫🛠️ With Gleece, you can:  
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
	"github.com/haimkastner/gleece/ctrl" // Importing GleeceController
)

// @Tag My API
// @Route /base-route
// @Description This is a description of that "tag"
type UserController struct {
	ctrl.GleeceController // Embedding the GleeceController
}

type Info struct {
	// @Description The address
	Address string `validate:"required"`
	// @Description The number of the house (must be at least 1)
	HouseNumber int `validate:"gte=1"`
}


// @Description This is a route under
// @Method POST
// @Route /user
// @Query name The name
// @Body info The info of the user
// @Header(x-origin) origin The origin of the user
// @ResponseDescription The ID of the newly created user
func (ec *UserController) CreateNewUser(name string, info Info, origin string) (string, error) {
	// Do the logic....
	userId := uuid.New()
	return userId.String(), nil
}
```
### What’s Happening Here?  
1. **Annotations**:  
   - Gleece uses annotations (like `@Tag`, `@Route`, and `@Description`) to automatically generate OpenAPI documentation.  

2. **Validation Handled by Gleece**:  
   - Input validation is simplified by Gleece using [go-playground/validator](https://github.com/go-playground/validator) format.  
   - You define validation rules directly on your struct fields:  
     - `validate:"required"` ensures the `Address` field is mandatory.  
     - `validate:"gte=1"` ensures the `HouseNumber` field has a value of at least 1.  
   - Gleece processes these validation rules automatically during request handling.  

3. **Controllers**:  
   - Simply embed the `GleeceController` (imported from `github.com/haimkastner/gleece/ctrl`) into your own controllers to gain its functionality.  

4. **Automation**:  
   - No manual steps required—your OpenAPI spec is ready to go!  

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
    "github.com/haimkastner/gleece/generated_routes" // Import the generated routes file
)

func main() {
    // Create a Gin router
    router := gin.Default()

    // Register Gleece routes
    generated_routes.RegisterRoutes(router)

    // Start the server
    router.Run(":8080")
}
```
---

## 🚧 Disclaimer  
⚠️ **Work in Progress** 🚨  
Gleece is currently an under-development project. 🛠️ We’re working hard to make it amazing.

We’d love your feedback and contributions as we shape Gleece together! 🤝✨  

Stay tuned for updates, and feel free to open issues or pull requests to help us improve! 🌟  

---

## 📜 License  
Gleece is licensed under the **MIT License**. 📄 You are free to use, modify, and distribute it with attribution. See the `LICENSE` file for details.  

---

🌟 **Let’s make API development gleam with Gleece!** 🌟  

