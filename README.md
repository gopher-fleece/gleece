**Gleece** - Bringing joy and ease to API development in Go! 🚀   

<!-- Source code health & info -->
[![gleece](https://github.com/gopher-fleece/gleece/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/gopher-fleece/gleece/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gopher-fleece/gleece)](https://goreportcard.com/report/gopher-fleece/gleece)
[![Coverage Status](https://coveralls.io/repos/github/gopher-fleece/gleece/badge.svg?branch=main)](https://coveralls.io/github/gopher-fleece/gleece?branch=main)
[![Go Version](https://img.shields.io/github/go-mod/go-version/gopher-fleece/gleece)](https://github.com/gopher-fleece/gleece/blob/main/go.mod)

<!-- Packages, Releases etc -->
[![VSCode Extension](https://img.shields.io/visual-studio-marketplace/v/haim-kastner.gleece-extension?label=VSCode%20Extension)](https://marketplace.visualstudio.com/items?itemName=haim-kastner.gleece-extension)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen.svg)](./docs/STEPBYSTEP.md)
[![Latest Release](https://img.shields.io/github/v/release/gopher-fleece/gleece)](https://github.com/gopher-fleece/gleece/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/gopher-fleece/gleece.svg)](https://pkg.go.dev/github.com/gopher-fleece/gleece)

<!-- Supported standards -->
[![OpenAPI 3.0](https://img.shields.io/badge/OpenAPI-3.0.0-green.svg)](https://spec.openapis.org/oas/v3.0.0)
[![OpenAPI 3.1](https://img.shields.io/badge/OpenAPI-3.1.0-green.svg)](https://spec.openapis.org/oas/v3.1.0)

<!-- Supported frameworks -->
[![Gin Support](https://img.shields.io/badge/Gin-Supported-blue)](https://gin-gonic.com/)
[![Echo Support](https://img.shields.io/badge/Echo-Supported-blue)](https://echo.labstack.com/)
[![Gorilla Mux Support](https://img.shields.io/badge/Gorilla_Mux-Supported-blue)](https://github.com/gorilla/mux)
[![Chi Support](https://img.shields.io/badge/Chi-Supported-blue)](https://github.com/go-chi/chi)
[![Fiber Support](https://img.shields.io/badge/Fiber-Supported-blue)](https://github.com/gofiber/fiber)

<!-- Social -->
[![GitHub stars](https://img.shields.io/github/stars/gopher-fleece/gleece.svg?style=social&label=Stars)](https://github.com/gopher-fleece/gleece/stargazers) 
[![License](https://img.shields.io/github/license/gopher-fleece/gleece.svg?style=social)](https://github.com/gopher-fleece/gleece/blob/master/LICENSE)


---

## Philosophy  
Developing APIs doesn’t have to be a chore - it should be simple, efficient, and enjoyable.  

Gone are the days of manually writing repetitive boilerplate code or struggling to keep your API documentation in sync with your implementation.

With Gleece, you can:  
- ✨ **Simplify** your API development process.  
- 📄 Automatically **generate OpenAPI v3.0.0 / v3.1.0** specification directly from your code.  
- 🎯 Ensure your APIs are always **well-documented and consistent**.  
- ✅ **Validate input data** effortlessly to keep your APIs robust and secure.
- 🛡 **Security first** approach, easy authorization with supplied check function.
- ⚡️ Choose Your Framework - seamlessly works with **Gin, Echo, Gorilla Mux, Chi, & Fiber** Rest frameworks.

Gleece aims to make Go developers’ lives easier by seamlessly integrating API routes, validation, and documentation into a single cohesive workflow.

## 💫 Look & Feel  

Here’s a practical snippet of how Gleece simplifies your API development:  

![Screenshot](https://raw.githubusercontent.com/gophar-fleece/.github/main/docs/screenshots/usage-example.png)

All other aspects, including HTTP routing generation, authorization enforcement, payload validation, error handling, and OpenAPI v3 specification generation, are handled by Gleece.

## 🪄 How It Works

There's NO magic or hidden underlying hacks here! 😊

Once your API functions are ready, use the Gleece CLI to generate routes according to your chosen engine.

These generated routes work just like any manually written route and need to be registered to your engine's router instance through your application code.

You can continue to use your engine's native calls, middlewares, and other features as before - Gleece generated code acts only as a complementary plugin to your router.


## 📚 Documentation

- [Step By Step Guide](./docs/STEPBYSTEP.md)
- [Annotations & Options](./docs/ANNOTATIONS.md)
- [Authentication](./docs/AUTHENTICATION.md)
- [Validations](./docs/VALIDATION.md) 
- [Error handling](./docs/ERROR_HANDLING.md)
- [Middlewares](./docs/MIDDLEWARES.md)
- [Configuration](./docs/CONFIG.md)
- [Advanced](./docs/ADVANCED.md)

## 🌐 Integrating with Golang Rest Routers 

- [Gin](./docs/routers/GIN_INTEGRATION.md)
- [Echo](./docs/routers/ECHO_INTEGRATION.md)
- [Gorilla Mux](./docs/routers/MUX_INTEGRATION.md)
- [Chi](./docs/routers/CHI_INTEGRATION.md)
- [Fiber](./docs/routers/FIBER_INTEGRATION.md)


For a complete example project using Gleece, check out the [Gleece Example Project](https://github.com/gopher-fleece/gleecexample#readme). This project demonstrates how to set up and use Gleece in a real-world scenario, providing you with a practical reference to get started quickly.

## 🎨 VSCode Extension

To enhance your development experience with Gleece, we provide an official VSCode extension that provides intelligent annotation highlighting and improved code visibility.

For more information and capabilities see the [Gleece VSCode Extension](https://github.com/gopher-fleece/gleece-vscode-extension#readme).

To install it search `Gleece` in the "Extension" tab or go to the [VSCode Marketplace](https://marketplace.visualstudio.com/items?itemName=haim-kastner.gleece-extension).

## 💡 Our Initiative

We believe that API development should be code-first, focusing on the developer experience while maintaining industry standards. Coming from the TypeScript ecosystem, we were inspired by frameworks like [TSOA](https://github.com/lukeautry/tsoa) that handle everything from routing and validation to OpenAPI generation - all from a single source of truth: your code.

Read more about our initiative and development philosophy in my [Gleece Project](https://blog.castnet.club/en/blog/gleece-project) blog post.

## ⚠️ Disclaimer
Gleece is currently an under-development project.  We’re working hard to make it amazing.

We’d love your feedback and contributions as we shape Gleece!

---

## 📜 License  
Gleece is licensed under the [MIT LICENSE](./LICENSE). 

---

