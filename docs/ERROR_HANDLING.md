# Gleece Error Handling

Each route function must declare a return type of `error`. If there is no response payload, `error` will be the only return type. If there is a payload response, `error` will be the second and last return type.

> The returned error, whether it's the default or custom, must NOT be a pointer.

When an error is returned as non-nil, the operation will be considered failed, and the error will be formatted according to [rfc7807](https://datatracker.ietf.org/doc/html/rfc7807).

For example, this route function logic:

```go
// @Method(GET)
// @Route(/default-error)
func (ec *MyController) DefaultError() error {
	return fmt.Errorf("default error")
}
```

Will be formatted to:

```json
{

    "type":"Internal Server Error",
    "title":"",
    "detail":"Encountered an error during operation 'DefaultError'",
    "status":500,
    "instance":"/gleece/controller/error/DefaultError",
    "extensions":{
        "error":"default error"
    }

}
```

## Custom HTTP Error

To override the default error format, define your own error struct and set whatever properties you like:


For example:
```go
type CustomError struct {
	error
	Message string `json:"message"`
}

// @Method(GET)
// @Route(/custom-error)
func (mc *MyController) CustomError() CustomError {
	return CustomError{
		Message: "custom error",
	}
}
```

Will be formatted to:

```json
{
    "message" : "custom error"
}
```

## Custom HTTP error response code
By default, any error is returned with a `500` HTTP error code. To set a different code, use the `SetStatus` API.

For example:
```go
// @Method(GET)
// @Route(/default-error)
func (mc *MyController) DefaultError() error {
    mc.SetStatus(runtime.StatusNotImplemented)
	return fmt.Errorf("default error")
}
```

Will return a `501` HTTP response code.