# Gleece Config

The Gleece configuration file, usually named `gleece.config.json`, contains all the mandatory and available configurations and options to customize the code and specification generator.

Below are all the options with comments explaining each property:

```json5
{
    "commonConfig": { // Mandatory
        "controllerGlobs": [ // Mandatory with minimum one path - These paths are the directories/files where Gleece will search for controllers and types/structs. Defining controllers/types outside these paths will result in them being ignored or cause build errors. In this example, it will search in the entire codebase.
            "./*.go",
            "./**/*.go"
        ]
    },
    "routesConfig": { // Mandatory
        "engine": "gin", // Mandatory - The router engine to generate routes for. Available options: "gin", "echo", "mux", "chi", "fiber"
        "packageName": "<package name>", // Optional - Set the package name of the generated routes. Default is "routes"
        "outputPath": "./routes/gleece.routes.go", // Mandatory - The path of the generated routes
        "outputFilePerms": "0644", // Optional - Set Linux permissions for the generated routes go file
        "authorizationConfig": { // Mandatory
            "authFileFullPackageName": "<package>", // Mandatory - The package where the authorization function "GleeceRequestAuthorization" is implemented (e.g. github.com/gopher-fleece/gleece/auth)
            "enforceSecurityOnAllRoutes": true // Optional - Enforce during generation time that every route has at least one direct/inherited security annotation
        },
        "templateOverrides": { // Optional - Override selected engine templates. Key is the template name and value is the path of the custom template
            "ResponseHeaders": "./gin.custom.response.headers.hbs"
        }
    },
    "openAPIGeneratorConfig": { // Mandatory
        "openapi": "3.0.0", // Mandatory - The OpenAPI specification version to generate. Available options: "3.0.0", "3.1.0"
        "info": { // Mandatory - Metadata that will be added to the generated specification
            "title": "Sample API", // Mandatory
            "description": "This is a sample API", // Optional
            "termsOfService": "http://example.com/terms/", // Optional
            "contact": { // Optional
                "name": "API Support", // Optional
                "url": "http://www.example.com/support", // Optional
                "email": "support@example.com" // Mandatory
            },
            "license": { // Optional
                "name": "Apache 2.0", // Optional
                "url": "http://www.apache.org/licenses/LICENSE-2.0.html" // Mandatory
            },
            "version": "1.0.0" // Mandatory - The API project version (it's the project's API version, not OpenAPI spec version)
        },
        "baseUrl": "https://api.example.com", // Mandatory 
        "securitySchemes": [ // Optional - The security schema in the API, will be exposed to the OpenAPI specification AND the "name" will be used in the routes Security annotation
            {
                "description": "API Key for accessing the API", // Mandatory
                "name": "securitySchemaName", // Mandatory - The name of the Security schema, this field will be used in the relevant Security annotations
                "fieldName": "x-header-name", // Mandatory - The name of the field in the HTTP request
                "type": "apiKey", // Mandatory - The type of authorization, see OpenAPI spec for options
                "in": "header" // Mandatory - Where the authentication data will be held - see OpenAPI spec for options
            }
        ],
        "defaultSecurity": { // Optional - The default security to set for routes without a security annotation on the route or the controller. Used as a fallback
            "name": "securitySchemaName", // Mandatory - A security schema name (defined in securitySchemes) to be used as default
            "scopes": [ // Mandatory - Collection of the scopes to be used as default
                "read:all"
            ]
        },
        "specGeneratorConfig": { // Mandatory
            "outputPath": "./dist/openapi.json" // Mandatory - The path of the generated OpenAPI specification file
        }
    }
}
```
