package definitions

const Rfc7807ErrorName = "Rfc7807Error"
const Rfc7807ErrorFullPackage = "github.com/gopher-fleece/runtime"

var SupportedRoutingEngineStrings = []string{
	string(RoutingEngineGin),
	string(RoutingEngineEcho),
	string(RoutingEngineMux),
	string(RoutingEngineFiber),
	string(RoutingEngineChi),
}
