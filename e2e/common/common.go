package common

type RouterTest struct {
	Name                string
	Path                string
	Method              string
	Body                any
	Query               map[string]string
	Headers             map[string]string
	Form                map[string]string
	ExpectedStatus      int
	ExpectedBody        string
	ExpectedBodyContain string
	ExpendedHeaders     map[string]string
}

type RouterTestResult struct {
	Code    int
	Body    string
	Headers map[string]string
}
