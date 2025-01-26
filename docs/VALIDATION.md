# Gleece Validation

Input validation is simplified by Gleece using [go-playground/validator](https://github.com/go-playground/validator) v10 format.  

Gleece pass them to the `go-playground/validator` engine during request processing, expose them in the OpenAPI v3 specification (if it's supported in the spec) and returns 422 in case of not passing validation.

> Explorer the full options for validation in [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator/v10#section-readme) documentation.

## Validate for struct fields

The validation is read from the ordinary convention of `go-playground` validator.

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

## Validate for Rest Params (query, header etc.)

The validation is read from the annotation `validate` option.


```go
/ @Description Create a new user
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

- validate: "required" in @Path ensures the path parameter is mandatory
- validate: "required,email" in @Query ensures:
  - The email query parameter is mandatory
  - The value must be a valid email format

> In REST params, if the value is non-pointer, the parameter will be considered mandatory regardless of the `validate` content.

# Custom Validators

Gleece supports providing customized validators in a way very similar to how `go-playground/validator` supports it.

> Note that custom validators will be ignored in the specification.

Write your own validator function by implementing Gleece's `external.ValidationFunc` interface.

```go
package validators

import (
	"unicode"

	"github.com/gopher-fleece/gleece/external"
)

// Custom validation function to check if a string starts with a letter
func ValidateStartsWithLetter(fl external.ValidationFieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return false
	}
	firstChar := rune(field[0])
	return unicode.IsLetter(firstChar)
}
```

Append and define the newly created function to the `routesConfig.customValidators` array in the `gleece.config.json` configuration file.

```json
{
  "routesConfig": {
   ...
    "customValidators": [
			{
				"validateTagName": "validate_starts_with_letter",
				"functionName": "ValidateStartsWithLetter",
				"fullPackageName": "<the full package path>/validators"
			}
		],
    ...
  },
```

Once done, the `validate_starts_with_letter` validation is available to use across all API validations.

