package visitors

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"golang.org/x/tools/go/packages"
)

// StructVisitor builds declaration-side StructMeta objects.
// It does not mutate the graph; graph insertion is the caller's job.
type StructVisitor struct {
	BaseVisitor

	// typeUsageVisitor must be wired so the StructVisitor can parse field types.
	typeUsageVisitor *TypeUsageVisitor
}

// NewStructVisitor constructs a StructVisitor.
func NewStructVisitor(context *VisitContext) (*StructVisitor, error) {
	visitor := &StructVisitor{}
	if err := visitor.initialize(context); err != nil {
		return nil, err
	}

	return visitor, nil
}

func (v *StructVisitor) setTypeUsageVisitor(visitor *TypeUsageVisitor) {
	v.typeUsageVisitor = visitor
}

// VisitStructType builds metadata.StructMeta for a declaration TypeSpec.
// It does not mutate the graph; the caller is responsible for inserting the struct node.
func (v *StructVisitor) VisitStructType(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	typeSpec *ast.TypeSpec,
) (metadata.StructMeta, graphs.SymbolKey, error) {

	v.enterFmt("VisitStructType %s", typeSpec.Name.Name)
	defer v.exit()

	// First, check whether this is a 'special' and short-circuit if it is.
	specialSymKey := getSymKeyForSpecial(file, typeSpec)
	if specialSymKey != nil {
		return metadata.StructMeta{}, *specialSymKey, nil
	}

	// Obtain file version for this declaration - fail on error to avoid inconsistent state.
	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return metadata.StructMeta{}, graphs.SymbolKey{},
			fmt.Errorf("could not obtain file version for struct %s: %w",
				typeSpec.Name.Name, fvErr)
	}

	// Ensure the type is a struct
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return metadata.StructMeta{}, graphs.SymbolKey{},
			fmt.Errorf("type spec %s is not a struct", typeSpec.Name.Name)
	}

	typeParamEnv, typeParamDecls, err := v.buildDeclTypeParamEnv(pkg, file, typeSpec.TypeParams)
	if err != nil {
		return metadata.StructMeta{}, graphs.SymbolKey{}, err
	}

	// Build FieldMeta entries (no graph mutations)
	allFieldsMeta := make([]metadata.FieldMeta, 0)
	for _, field := range structType.Fields.List {
		fieldMetadata, err := v.getFieldMeta(pkg, file, fileVersion, field, typeParamEnv)
		if err != nil {
			return metadata.StructMeta{}, graphs.SymbolKey{}, fmt.Errorf(
				"failed to obtain metadata for field '%s' - %v",
				strings.Join(gast.GetFieldNames(field), ", "),
				err,
			)
		}
		allFieldsMeta = append(allFieldsMeta, fieldMetadata...)

	}

	structMeta := metadata.StructMeta{
		SymNodeMeta: metadata.SymNodeMeta{
			Name:        typeSpec.Name.Name,
			Node:        typeSpec,
			SymbolKind:  common.SymKindStruct,
			PkgPath:     pkg.PkgPath,
			Annotations: nil,
			FVersion:    fileVersion,
			Range:       common.ResolveNodeRange(pkg.Fset, typeSpec),
		},
		TypeParams: typeParamDecls,
		Fields:     allFieldsMeta,
	}

	structSymKey, err := v.graphStructAndFields(structMeta)

	return structMeta, structSymKey, err
}

func (v *StructVisitor) getFieldMeta(
	pkg *packages.Package,
	file *ast.File,
	fileVersion *gast.FileVersion,
	field *ast.Field,
	typeParamEnv map[string]int,
) ([]metadata.FieldMeta, error) {
	// build a usage-side TypeUsageMeta for the field type (delegated)
	typeUsage, err := v.typeUsageVisitor.VisitExpr(pkg, file, field.Type, typeParamEnv)
	if err != nil {
		return []metadata.FieldMeta{}, err
	}

	names := gast.GetFieldNames(field)
	isEmbedded := gast.IsEmbeddedOrAnonymousField(field)

	meta := []metadata.FieldMeta{}
	for _, name := range names {
		fieldMeta := metadata.FieldMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        name,
				Node:        field,
				SymbolKind:  common.SymKindField,
				PkgPath:     pkg.PkgPath,
				Annotations: nil,
				FVersion:    fileVersion,
				Range:       common.ResolveNodeRange(pkg.Fset, field),
			},
			Type:       typeUsage,
			IsEmbedded: isEmbedded,
		}
		meta = append(meta, fieldMeta)
	}

	return meta, nil
}

func (v *StructVisitor) graphStructAndFields(structMeta metadata.StructMeta) (graphs.SymbolKey, error) {
	// Insert struct node into graph (graph builder will de-dupe).
	createStructReq := symboldg.CreateStructNode{
		Data:        structMeta,
		Annotations: structMeta.Annotations,
	}

	structNode, err := v.context.GraphBuilder.AddStruct(createStructReq)
	if err != nil {
		return graphs.SymbolKey{}, err
	}

	// Now insert each field node (so field nodes exist, and edges can be created).
	// GraphBuilder.AddField will create the field node and add EdgeKindType -> typeRef.
	for _, fieldMeta := range structMeta.Fields {
		createFieldReq := symboldg.CreateFieldNode{
			Data:        fieldMeta,
			Annotations: fieldMeta.Annotations,
		}
		if _, err := v.context.GraphBuilder.AddField(createFieldReq); err != nil {
			return graphs.SymbolKey{}, err
		}
	}

	return structNode.Id, nil

}

// buildDeclTypeParamEnv builds a name->index env and a []TypeParamDecl.
// It assigns indexes first, then parses constraints (if present) using the
// typeUsageVisitor and the env. Short and deterministic.
func (v *StructVisitor) buildDeclTypeParamEnv(
	pkg *packages.Package,
	file *ast.File,
	params *ast.FieldList,
) (map[string]int, []metadata.TypeParamDecl, error) {

	env := map[string]int{}
	decls := []metadata.TypeParamDecl{}
	if params == nil {
		return env, decls, nil
	}

	// First pass: assign indexes deterministically
	idx := 0
	for _, fld := range params.List {
		for _, name := range fld.Names {
			env[name.Name] = idx
			decls = append(decls, metadata.TypeParamDecl{
				Name:  name.Name,
				Index: idx,
			})
			idx++
		}
	}

	// Second pass: parse constraints (if any) and attach to decls
	idx = 0
	for _, fld := range params.List {
		var constraintRef metadata.TypeRef
		if fld.Type != nil {
			// parse constraint using same env (constraints can refer to earlier params)
			tuMeta, err := v.typeUsageVisitor.VisitExpr(pkg, file, fld.Type, env)
			if err != nil {
				return nil, nil, err
			}
			constraintRef = tuMeta.Root
		}
		// assign constraint to every name in this field (usually single)
		for range fld.Names {
			decls[idx].Constraint = constraintRef
			idx++
		}
	}

	return env, decls, nil
}

// getSymKeyForSpecial gets a SymbolKey for the given spec, if it's a 'special', otherwise returns nil
func getSymKeyForSpecial(file *ast.File, spec *ast.TypeSpec) *graphs.SymbolKey {
	if file == nil || file.Name == nil || spec == nil || spec.Name == nil {
		// A nameless entity cannot be a 'special'
		return nil
	}

	fName := file.Name.Name
	specName := spec.Name.Name

	if specName == "Time" && fName == "time" || specName == "Context" && fName == "context" {
		return common.Ptr(graphs.NewNonUniverseBuiltInSymbolKey(fmt.Sprintf("%s.%s", fName, specName)))
	}

	// Not a standard special
	return nil
}
