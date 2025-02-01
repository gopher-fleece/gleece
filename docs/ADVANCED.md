# Advanced Abilities for Controllers

### Set HTTP Response Code in Run-time

When a controller function returns with a nil error, the operation is considered successful.

If the function has no response payload, the status code will be `204`. If it contains a payload, the status code will be `200`, as per HTTP specifications.

If the function returns an error, the default error code will be `500`.

However, it is possible to set a custom response code using the `SetStatus` API.

```go
// @Method(GET)
// @Route(/my-route)
func (mc *MyController) MyRoute() (string, error) {
	mc.SetStatus(external.StatusPartialContent)
	return "", nil
}
```

### Set HTTP Response Header in Run-time

To set HTTP response headers, use the `SetHeader` API.

```go
// @Method(GET)
// @Route(/my-route)
func (mc *MyController) MyRoute() (string, error) {
	mc.SetHeader("X-my-Header", "some string")
	mc.SetHeader("X-my-Header-2", "some string 2")
	return "", nil
}
```

### Access HTTP Request Context

Sometimes, especially in edge cases, there is a need to access the full HTTP request context.

This might be necessary to perform extra dynamic operations on the request, support features not yet implemented and integrated into Gleece, or other specific needs.

Whatever the reason, you can access the router context via the `GetContext` API. The type is `any`, and you can cast it to your router's specific context type.

For example, for `gin`:
```go
// @Method(GET)
// @Route(/my-route)
func (mc *MyController) MyRoute() (string, error) {
	context := mc.GetContext()
	ginContext, ok := context.(*gin.Context)
    // Do the advanced logic with the ginContext....
	return "", nil
}
```

# Template Overriding

Gleece supports template overriding.

To override a template, navigate to the template you want to override based on the engine specified in the `routesConfig->engine` configuration. For example, the `gin` templates are located here:  
[https://github.com/gopher-fleece/gleece/generator/templates/gin](https://github.com/gopher-fleece/gleece/tree/main/generator/templates/gin)

To inject a custom template, add the template name as a key and the path to the custom template as its value in the `routesConfig->templateOverrides` map.

Your configuration might look like this:

```json
...
"routesConfig": {
    "engine": "gin",
    "outputPath": "./routes/gin.e2e.gleece.go",
    "outputFilePerms": "0644",
    "authorizationConfig": {
        "authFileFullPackageName": "github.com/gopher-fleece/gleece/e2e/gin/auth",
        "enforceSecurityOnAllRoutes": true
    },
    "templateOverrides": {
        "ResponseHeaders" : "./assets/gin.custom.response.headers.hbs"
    }
},
...
```

The main template name is `Routes` (for all engines), and each engine’s partial template names are listed in the `Partials` map within the engine embed file. For the `gin` engine, see: [https://github.com/gopher-fleece/gleece/generator/templates/gin/embeds.go](https://github.com/gopher-fleece/gleece/blob/main/generator/templates/gin/embeds.go)  

