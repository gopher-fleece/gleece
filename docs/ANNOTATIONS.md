# Gleece Annotations Documentation

This document provides a comprehensive guide to all annotations and options supported by Gleece for controllers and schemas.

## Table of Contents
- [Paths Annotations](#paths-annotations)
  - [Class-Level Annotations - Controller](#class-level-annotations---controller)
  - [Function-Level Annotations - Route](#function-level-annotations---route)
- [Schema Annotations](#schema-annotations)
  - [Struct Annotations - Schema](#struct-annotations---schema)
  - [Field Annotations - Property](#field-annotations---property)
- [Advanced Options](#advanced-options)
  - [Query Options](#query-options)
  - [Path Options](#path-options)
  - [Header Options](#header-options)
  - [Body Options](#body-options)
  - [Security Options](#security-options)

# Paths Annotations

## Class-Level Annotations - Controller

These annotations are used at the controller class level.

| Annotation    | Purpose                                          | Parameter          | JSON5 Options Ref                          | Required | comment |
|---------------|--------------------------------------------------|--------------------|--------------------------------------------|----------|---------|
| `@Tag`        | Defines the OpenAPI tag for grouping routes      | tagName            | -                                          | Yes      |         |
| `@Route`      | Defines the base route path for all methods      | path               | -                                          | Yes      |         |
| `@Security`   | Set the default security for controller's routes | securitySchemeName | [Security Options](#security-options)      | No       | Supports multiple annotation and will be count as logical OR between them |
| `@Description`| Provides a detailed description of the controller| -                  | -                                          | Yes      |         | 
| `@Deprecated` | Marks the entire controller as deprecated        | -                  | -                                          | No       |         |

## Function-Level Annotations - Route
| Annotation      | Purpose                                                    | Parameter          | JSON5 Options Ref                     | Required | Comment |
|-----------------|------------------------------------------------------------|--------------------|---------------------------------------|----------|---------|
| `@Method`       | Defines the HTTP method for the route                      | httpVerb           | -                                     | Yes      |         | 
| `@Route`        | Defines the route path for the method                      | path               | -                                     | Yes      |         |
| `@Query`        | Defines a query parameter for the endpoint                 | paramName          | [Query Options](#query-options)       | No       | Support description |
| `@Header`       | Defines an header parameter for the endpoint               | paramName          | [Header Options](#header-options)     | No       | Support description |
| `@Path`         | Defines a path parameter for the endpoint                  | paramName          | [Path Options](#path-options)         | No       | Support description |
| `@Body`         | Defines a body parameter for the endpoint                  | paramName          | [Body Options](#body-options)         | No       | Support description |
| `@Security`     | Defines a security for the route                           | securitySchemeName | [Security Options](#security-options) | No       | Supports multiple annotation and will be count as logical OR between them |
| `@Response`     | Specifies the success response code                        | statusCode         | -                                     | No       | Description can include HTML formatting |
| `@ErrorResponse`| Defines possible error response                            | statusCode         | -                                     | No       | Description can include HTML formatting |
| `@Description`  | Provides a detailed description of the endpoint            | -                  | -                                     | No       |         | 
| `@Deprecated`   | Marks the endpoint as deprecated                           | -                  | -                                     | No       |         |
| `@Hidden`       | Hides the endpoint from API documentation                  | -                  | -                                     | No       |         |

# Schema Annotations

## Struct Annotations - Schema
| Annotation      | Purpose                                                    | Parameter     | JSON5 Options Ref                 | Required | Comment |
|-----------------|------------------------------------------------------------|--------------|------------------------------------|----------|---------|
| `@Description`  | Provides a detailed description of the schema              | -            | -                                  | No       |         | 
| `@Deprecated`   | Marks the schema as deprecated                             | -            | -                                  | No       |         |
| `@Hidden`       | Hides the schema from API documentation                    | -            | -                                  | No       |         |

## Field Annotations - Property
| Annotation      | Purpose                                                    | Parameter     | JSON5 Options Ref                 | Required | Comment |
|-----------------|------------------------------------------------------------|--------------|------------------------------------|----------|---------|
| `@Description`  | Provides a detailed description of the property            | -            | -                                  | No       |         | 
| `@Deprecated`   | Marks the property as deprecated                           | -            | -                                  | No       |         |
| `@Hidden`       | Hides the property from API documentation                  | -            | -                                  | No       |         |

> A struct's property supported the standard json and validate in the field tag e.g. `json:"fieldName" validate:"required"`

# Advanced Options

### Query Options

| Field Name    | Description                                     | Default                   |
|---------------|-------------------------------------------------|---------------------------|
| `name`        | The name of the query in the HTTP request       | Function parameter name   |
| `validate`    | Standard go-playground validation string        | -                         |

### Path Options

| Field Name    | Description                                     | Default                   |
|---------------|-------------------------------------------------|---------------------------|
| `name`        | The name of the path element in the HTTP request| Function parameter name   |
| `validate`    | Standard go-playground validation string        | -                         |

### Header Options

| Field Name    | Description                                    | Default                   |
|---------------|------------------------------------------------|---------------------------|
| `name`        | The name of the header in the HTTP request     | Function parameter name   |
| `validate`    | Standard go-playground validation string       | -                         |

### Body Options

| Field Name    | Description                                    | Default                   |
|---------------|------------------------------------------------|---------------------------|
| `validate`    | Standard go-playground validation string       | -                         |

### Security Options

| Field Name    | Description                                    | Default                   |
|---------------|------------------------------------------------|---------------------------|
| `scopes`      | The scopes e.g. `[read:users, write:users]`    | -                         |