package typeref

import (
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

// NamedTypeRef: reference to an existing declared type via graphs.SymbolKey.
// TypeArgs are concrete type arguments if this usage is an instantiation (e.g. MyType[int]).
type NamedTypeRef struct {
	Key      graphs.SymbolKey
	TypeArgs []metadata.TypeRef
}

func (n *NamedTypeRef) Kind() metadata.TypeRefKind { return metadata.TypeRefKindNamed }

func (n *NamedTypeRef) CanonicalString() string {
	base := canonicalSymKey(n.Key)
	if len(n.TypeArgs) == 0 {
		return base
	}
	argStrings := make([]string, 0, len(n.TypeArgs))
	for _, a := range n.TypeArgs {
		argStrings = append(argStrings, a.CanonicalString())
	}
	return fmt.Sprintf("%s[%s]", base, strings.Join(argStrings, ","))
}

func (n *NamedTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	// if Key present (declared/universe), use it
	if !n.Key.Equals(graphs.SymbolKey{}) {
		// Instantiation case: combine base key with type args (if any).
		if len(n.TypeArgs) == 0 {
			return n.Key, nil
		}

		// Build arg keys first.
		argKeys := make([]graphs.SymbolKey, 0, len(n.TypeArgs))
		for _, arg := range n.TypeArgs {
			key, err := arg.ToSymKey(fileVersion)
			if err != nil {
				return graphs.SymbolKey{}, err
			}
			argKeys = append(argKeys, key)
		}
		return graphs.NewInstSymbolKey(n.Key, argKeys), nil
	}

	// If no base Key present but type args exist, we cannot produce a stable instantiation
	if len(n.TypeArgs) > 0 {
		return graphs.SymbolKey{}, fmt.Errorf("cannot instantiate named type without base Key")
	}

	// No Key and no TypeArgs: this is unexpected.
	return graphs.SymbolKey{}, fmt.Errorf("named type ref missing Key")
}

func (f *NamedTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
