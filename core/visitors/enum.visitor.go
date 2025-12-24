package visitors

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
)

// EnumVisitor parses enum-like TypeSpecs and consts and materializes Enum nodes.
type EnumVisitor struct {
	BaseVisitor
}

// NewEnumVisitor constructs an EnumVisitor.
func NewEnumVisitor(ctx *VisitContext) (*EnumVisitor, error) {
	v := &EnumVisitor{}
	if err := v.initialize(ctx); err != nil {
		return nil, err
	}
	return v, nil
}

// VisitEnumType analyzes a type spec that is expected to be an enum-like alias (e.g. type X string).
// It returns the constructed EnumMeta and the graph SymbolKey for the created node.
func (v *EnumVisitor) VisitEnumType(
	pkg *packages.Package,
	file *ast.File,
	fileVersion *gast.FileVersion,
	specGenDecl *ast.GenDecl,
	spec *ast.TypeSpec,
) (metadata.EnumMeta, graphs.SymbolKey, error) {
	specName := gast.GetIdentNameOrFallback(spec.Name, "N/A")

	v.enterFmt("Visiting enum %s.%s", pkg.PkgPath, specName)
	defer v.exit()

	// Cache check
	symKey := graphs.NewSymbolKey(spec, fileVersion)
	if cached := v.context.MetadataCache.GetEnum(symKey); cached != nil {
		return *cached, symKey, nil
	}

	// Resolve the named type object (types.TypeName) for this declaration
	typeName, err := gast.GetTypeNameOrError(pkg, spec.Name.Name)
	if err != nil {
		return metadata.EnumMeta{}, graphs.SymbolKey{}, err
	}

	// Extract enum metadata (value kind + values)
	enumMeta, err := v.extractEnumAliasType(pkg, fileVersion, specGenDecl, spec, typeName)
	if err != nil {
		return metadata.EnumMeta{}, graphs.SymbolKey{}, err
	}

	// Cache it
	v.context.MetadataCache.AddEnum(&enumMeta)

	// Insert into graph
	create := symboldg.CreateEnumNode{
		Data:        enumMeta,
		Annotations: enumMeta.Annotations,
	}
	enumNode, err := v.context.Graph.AddEnum(create)
	if err != nil {
		return metadata.EnumMeta{}, graphs.SymbolKey{}, err
	}

	return enumMeta, enumNode.Id, nil
}

// extractEnumAliasType builds metadata.EnumMeta from a types.TypeName (must be alias to a basic type).
func (v *EnumVisitor) extractEnumAliasType(
	enumPkg *packages.Package,
	enumFVersion *gast.FileVersion,
	specGenDecl *ast.GenDecl,
	spec *ast.TypeSpec,
	typeName *types.TypeName,
) (metadata.EnumMeta, error) {
	v.enterFmt("Extracting metadata for enum %s.%s", enumPkg.PkgPath, typeName.Name())
	defer v.exit()

	// Underlying must be a basic type (string/int/...)
	basic, ok := typeName.Type().Underlying().(*types.Basic)
	if !ok {
		return metadata.EnumMeta{}, NewUnexpectedEntityError(
			common.SymKindEnum,
			common.SymKindUnknown,
			"type '%s' is not a basic type and therefore not an enum",
			typeName.Name(),
		)
	}

	kind, err := metadata.NewEnumValueKind(basic.Kind())
	if err != nil {
		return metadata.EnumMeta{}, NewUnexpectedEntityError(
			common.SymKindEnum,
			common.SymKindUnknown,
			"type '%s' has an underlying basic kind '%v' that is not enum-compatible and is therefore not an enum",
			typeName.Name(),
			basic.Kind(),
		)
	}

	// Annotations (from TypeSpec doc or parent GenDecl)
	enumAnnotations, err := v.getAnnotations(spec.Doc, specGenDecl)
	if err != nil {
		return metadata.EnumMeta{}, err
	}

	// Gather values from package scope
	enumValues, err := v.getEnumValueDefinitions(enumFVersion, enumPkg, typeName, basic)
	if err != nil {
		return metadata.EnumMeta{}, err
	}

	if len(enumValues) <= 0 {
		return metadata.EnumMeta{}, NewUnexpectedEntityError(
			common.SymKindEnum,
			common.SymKindUnknown,
			"type '%s' does not appear to have 'values' in the same file and is therefore not an enum",
			typeName.Name(),
		)
	}

	enum := metadata.EnumMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        typeName.Name(),
			Node:        spec,
			SymbolKind:  common.SymKindEnum,
			PkgPath:     enumPkg.PkgPath,
			FVersion:    enumFVersion,
			Annotations: enumAnnotations,
			Range:       common.ResolveNodeRange(enumPkg.Fset, spec),
		},
		ValueKind: kind,
		Values:    enumValues,
	}

	return enum, nil
}

// getEnumValueDefinitions scans the package scope and returns all consts of the given named enum type.
func (v *EnumVisitor) getEnumValueDefinitions(
	enumFVersion *gast.FileVersion,
	enumPkg *packages.Package,
	enumTypeName *types.TypeName,
	enumBasic *types.Basic,
) ([]metadata.EnumValueDefinition, error) {

	v.enterFmt("Obtaining enum value definitions for %s.%s", enumPkg.PkgPath, enumTypeName.Name())
	defer v.exit()

	out := []metadata.EnumValueDefinition{}

	scope := enumPkg.Types.Scope()
	if scope == nil {
		return out, nil
	}

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		constObj, ok := obj.(*types.Const)
		if !ok {
			continue
		}

		// To avoid clobbering consts that happen to have the same type under aliases/enums,
		// we check the actual underlying type name as well.
		if curObjAlias, isCurObjAlias := constObj.Type().(*types.Alias); isCurObjAlias {
			if enumTypeName.Name() != curObjAlias.Obj().Name() {
				continue
			}
		}

		// ensure the const's type exactly matches the enum type
		if !types.Identical(enumTypeName.Type(), constObj.Type()) {
			continue
		}

		// extract a stable value using your helper (returns a stringable info)
		val := gast.ExtractConstValue(enumBasic.Kind(), constObj)
		if val == nil {
			// skip consts we cannot represent
			continue
		}

		// try to find the AST node (ValueSpec) for this const if possible
		var constNode ast.Node
		if v.context != nil && v.context.ArbitrationProvider != nil {
			constNode = gast.FindConstSpecNode(enumPkg, constObj.Name())
		}

		// attempt to extract annotations for individual enum value
		var holder *annotations.AnnotationHolder
		if constNode != nil {
			comments := gast.GetCommentsFromNode(constNode, v.context.ArbitrationProvider.Pkg().FSet())
			if h, err := annotations.NewAnnotationHolder(comments, annotations.CommentSourceProperty); err == nil {
				holder = &h
			} else {
				return nil, err
			}
		}

		ev := metadata.EnumValueDefinition{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        constObj.Name(),
				Node:        constNode,
				SymbolKind:  common.SymKindEnumValue,
				PkgPath:     enumPkg.PkgPath,
				Annotations: holder,
				FVersion:    enumFVersion,
				Range:       common.ResolveNodeRange(enumPkg.Fset, constNode),
			},
			Value: val,
		}

		out = append(out, ev)
	}

	return out, nil
}
