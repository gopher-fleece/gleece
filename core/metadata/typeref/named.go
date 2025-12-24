package typeref

import (
	"fmt"
	"strings"

	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
)

// NamedTypeRef: reference to an existing declared type via graphs.SymbolKey.
// TypeArgs are concrete type arguments if this usage is an instantiation (e.g. MyType[int]).
type NamedTypeRef struct {
	Key      graphs.SymbolKey
	TypeArgs []metadata.TypeRef
}

// NewNamedTypeRef creates a new named type reference.
// Serves mostly as a single point of reference for easier code lookups
func NewNamedTypeRef(key *graphs.SymbolKey, typeArgs []metadata.TypeRef) NamedTypeRef {
	ref := NamedTypeRef{TypeArgs: typeArgs}
	if key != nil {
		ref.Key = *key
	}
	return ref
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

func (n *NamedTypeRef) SimpleTypeString() string {
	if len(n.TypeArgs) == 0 {
		return n.Key.Name
	}

	argStrings := make([]string, 0, len(n.TypeArgs))
	for _, a := range n.TypeArgs {
		argStrings = append(argStrings, a.SimpleTypeString())
	}

	return fmt.Sprintf("%s[%s]", n.Key.Name, strings.Join(argStrings, ","))
}

func (n *NamedTypeRef) CacheLookupKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	return n.ToSymKey(fileVersion)
}

func (n *NamedTypeRef) ToSymKey(fileVersion *gast.FileVersion) (graphs.SymbolKey, error) {
	// if Key present (declared/universe), use it
	if !n.Key.Empty() {
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

	// No Key and no TypeArgs - this is unexpected.
	return graphs.SymbolKey{}, fmt.Errorf("named type ref missing Key")
}

func (f *NamedTypeRef) Flatten() []metadata.TypeRef {
	return flatten(f)
}
