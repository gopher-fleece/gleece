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