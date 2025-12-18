package metadata

// Represents a generic parameter declaration
// e.g.,
//
//	TValue
//
// in
//
//	SomeStruct[TValue]
type TypeParamDeclMeta struct {
	// The parameter's name, as determined by the symbol key, e.g.,
	// 	"typeparam:TValue#0"
	Name string

	// The parameter's index at the declaration,
	// Example - In
	//
	// 	SomeStruct[TA, TB]
	//
	// TA has index 0 and TB has index 1
	Index int
}
