package definitions

// The name of the RFC7807 error struct
const Rfc7807ErrorName = "Rfc7807Error"

// The full package path of the RFC7807 error struct
const Rfc7807ErrorFullPackage = "github.com/gopher-fleece/runtime"

// Defines Gleece's supported routing engines
var SupportedRoutingEngineStrings = []string{
	string(RoutingEngineGin),
	string(RoutingEngineEcho),
	string(RoutingEngineMux),
	string(RoutingEngineFiber),
	string(RoutingEngineChi),
}
