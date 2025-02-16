# Gleece Validation

Input validation is simplified by Gleece using [go-playground/validator](https://github.com/go-playground/validator) v10 format.  

Gleece passes validation rules to the `go-playground/validator` engine during request processing, exposes them in the OpenAPI v3 specification (where supported), and returns a 422 status code if validation fails.

> Explore the full validation options in the [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator/v10#section-readme) documentation.

## Validating Struct Fields

Validation rules follow the standard convention of the `go-playground` validator.

```go
// @Description User's domicile
type Domicile struct {
	Address string `json:"address" validate:"required"`
	// @Description The number of the house (must be at least 1)
	HouseNumber int `json:"houseNumber" validate:"gte=1"`
}
```

- `validate:"required"` ensures the `Address` field is mandatory.  
- `validate:"gte=1"` ensures the `HouseNumber` field has a value of at least 1.  

## Validating REST Parameters (Query, Header, etc.)

Validation rules are defined using the `validate` annotation option.

```go
// @Description Create a new user
// @Method(POST)
// @Route(/user/{user_name})
// @Path(name, { name: "user_name", validate: "require" }) The user's name
// @Query(email, { validate: "required,email" }) The user's email
// @Body(domicile) The user's domicile info
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

- `validate: "required"` in `@Path` ensures the path parameter is mandatory
- `validate: "required,email"` in `@Query` ensures:
  - The email query parameter is mandatory
  - The value must be in a valid email format

> For REST parameters, non-pointer values are always considered mandatory, regardless of the `validate` content. Path parameters are always mandatory as per the OpenAPI specification (a missing path results in a `404` error).

# Custom Validators

Gleece supports custom validators in a manner similar to `go-playground/validator`.

> Note: Custom validators only affect runtime validation and are not reflected in the OpenAPI specification.

To create a custom validator, implement Gleece's `runtime.ValidationFunc` interface and register it using the generated code API's `RegisterCustomValidator` function.

```go
package main

import (
	"unicode"
    "net/http"

    gleeceRoutes "<package>"

	"github.com/gopher-fleece/runtime"
    "github.com/gin-gonic/gin"
)

// Custom validation function to check if a string starts with a letter
func ValidateStartsWithLetter(fl runtime.ValidationFieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return false
	}
	firstChar := rune(field[0])
	return unicode.IsLetter(firstChar)
}

func main() {
    // Create a default Gin router
    router := gin.Default()

	// Register the `ValidateStartsWithLetter` custom validator
	gleeceRoutes.RegisterCustomValidator("validate_starts_with_letter", ValidateStartsWithLetter)

    // Register the routes from the generated code
    gleeceRoutes.RegisterRoutes(router)

    // Start the server on port 8080
    router.Run("127.0.0.1:8080")
}
```

After registration, the `validate_starts_with_letter` validator becomes active across all API requests.