package definitions

import "fmt"

var validHttpVerbs = map[string]struct{}{
	string(HttpGet):     {},
	string(HttpPost):    {},
	string(HttpPut):     {},
	string(HttpDelete):  {},
	string(HttpPatch):   {},
	string(HttpOptions): {},
	string(HttpHead):    {},
	string(HttpTrace):   {},
	string(HttpConnect): {},
}

func IsValidHttpVerb(verb string) bool {
	_, exists := validHttpVerbs[verb]
	return exists
}

func EnsureValidHttpVerb(verb string) HttpVerb {
	if IsValidHttpVerb(verb) {
		return HttpVerb(verb)
	}
	panic(fmt.Sprintf("'%s' is not a valid HTTP verb", verb))
}
