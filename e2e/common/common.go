package common

type RunningMode string

const (
	// Run test on fully featured routes only (where all features are activated. g.e. middleware, templates execution & override, etc)
	RunOnFullyFeaturedRoutes RunningMode = "RunOnFullyFeaturedRoutes"
	// Run test on vanilla routes only same as default but without any extra features
	RunOnVanillaRoutes RunningMode = "RunOnVanillaRoutes"
	// Run test on all routes (fully featured and vanilla)
	RunOnAllRoutes RunningMode = "RunOnAllRoutes"
)

type RouterTest struct {
	Name                string
	Path                string
	Method              string
	Body                any
	Query               map[string]string
	QueryArray          map[string][]string
	Headers             map[string]string
	Form                map[string]string
	ExpectedStatus      int
	ExpectedBody        string
	ExpectedBodyContain string
	ExpendedHeaders     map[string]string
	RunningMode         *RunningMode
}

type RouterTestResult struct {
	Code    int
	Body    string
	Headers map[string]string
}
