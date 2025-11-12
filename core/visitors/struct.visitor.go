package visitors

import (
	"go/ast"
	"strings"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
)

// StructVisitor builds declaration-side StructMeta objects.
// It does not mutate the graph; graph insertion is the caller's job.
type StructVisitor struct {
	BaseVisitor

	// typeUsageVisitor must be wired so the StructVisitor can parse field types.
	typeUsageVisitor *TypeUsageVisitor

	nonMaterializingTypeUsageVisitor *TypeUsageVisitor
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

func (v *StructVisitor) setNonMaterializingTypeUsageVisitor(visitor *TypeUsageVisitor) {
	v.nonMaterializingTypeUsageVisitor = visitor
}

// VisitStructType builds metadata.StructMeta for a declaration TypeSpec.
// It does not mutate the graph; the caller is responsible for inserting the struct node.
func (v *StructVisitor) VisitStructType(
	pkg *packages.Package,
	file *ast.File,
	genDecl *ast.GenDecl,
	typeSpec *ast.TypeSpec,
) (metadata.StructMeta, graphs.SymbolKey, error) {
	tsName := "unknown"
	if typeSpec != nil && typeSpec.Name != nil {
		tsName = typeSpec.Name.Name
	}

	v.enterFmt("Visiting '%s'", tsName)
	defer v.exit()

	// First, check whether this is a 'special' and short-circuit if it is.
	var specName string
	if typeSpec != nil && typeSpec.Name != nil {
		specName = typeSpec.Name.Name
	}

	specialSymKey := v.typeUsageVisitor.tryGetSymKeyForSpecial(pkg, typeSpec, specName)
	if specialSymKey != nil {
		return metadata.StructMeta{}, *specialSymKey, nil
	}

	// Obtain file version for this declaration - fail on error to avoid inconsistent state.
	fileVersion, fvErr := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if fvErr != nil || fileVersion == nil {
		return metadata.StructMeta{}, graphs.SymbolKey{},
			v.getFrozenError(
				"could not obtain file version for struct %s: %w",
				typeSpec.Name.Name,
				fvErr,
			)
	}

	// Ensure the type is a struct
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return metadata.StructMeta{}, graphs.SymbolKey{}, v.getFrozenError("type spec %s is not a struct", typeSpec.Name.Name)
	}

	typeParamEnv, typeParamDecls, err := v.buildDeclTypeParamEnv(pkg, file, typeSpec.TypeParams)
	if err != nil {
		return metadata.StructMeta{}, graphs.SymbolKey{}, err
	}

	holder, err := v.getAnnotations(typeSpec.Doc, genDecl)
	if err != nil {
		return metadata.StructMeta{}, graphs.SymbolKey{}, v.getFrozenError("failed to obtain notifications for struct '%s' - %v", tsName, err)
	}

	// Build FieldMeta entries (no graph mutations)
	allFieldsMeta := make([]metadata.FieldMeta, 0)
	for _, field := range structType.Fields.List {
		fieldMetadata, err := v.getFieldMeta(pkg, file, fileVersion, field, typeParamEnv)
		if err != nil {
			return metadata.StructMeta{}, graphs.SymbolKey{}, v.getFrozenError(
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
			Annotations: holder,
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
	names := gast.GetFieldNames(field)

	// build a usage-side TypeUsageMeta for the field type (delegated)
	typeUsage, err := v.typeUsageVisitor.VisitExpr(pkg, file, field.Type, typeParamEnv)
	if err != nil {
		return nil, v.getFrozenError("failed to visit expression for field/s [%v] - %v", names, err)
	}

	isEmbedded := gast.IsEmbeddedOrAnonymousField(field)

	holder, err := v.getAnnotations(field.Doc, nil)
	if err != nil {
		return nil, v.getFrozenError("failed to obtain annotations for field/s [%v] - %v", names, err)
	}

	createMeta := func(name string) metadata.FieldMeta {
		return metadata.FieldMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        name,
				Node:        field,
				SymbolKind:  common.SymKindField,
				PkgPath:     pkg.PkgPath,
				Annotations: holder,
				FVersion:    fileVersion,
				Range:       common.ResolveNodeRange(pkg.Fset, field),
			},
			Type:       typeUsage,
			IsEmbedded: isEmbedded,
		}
	}

	if len(names) <= 0 {
		return []metadata.FieldMeta{createMeta(typeUsage.Name)}, nil
	}

	meta := []metadata.FieldMeta{}
	for _, name := range names {
		meta = append(meta, createMeta(name))
	}

	return meta, nil
}

func (v *StructVisitor) graphStructAndFields(structMeta metadata.StructMeta) (graphs.SymbolKey, error) {
	// Insert struct node into graph (graph builder will de-dupe).
	createStructReq := symboldg.CreateStructNode{
		Data:        structMeta,
		Annotations: structMeta.Annotations,
	}

	structNode, err := v.context.Graph.AddStruct(createStructReq)
	if err != nil {
		return graphs.SymbolKey{}, err
	}

	if err := v.context.MetadataCache.AddStruct(&structMeta); err != nil {
		logger.Warn("Struct visitor failed to cache struct '%s' - %v", structMeta.Name, err)
	}

	// Now insert each field node (so field nodes exist, and edges can be created).
	// GraphBuilder.AddField will create the field node and add EdgeKindType -> typeRef.
	for _, fieldMeta := range structMeta.Fields {
		createFieldReq := symboldg.CreateFieldNode{
			Data:        fieldMeta,
			Annotations: fieldMeta.Annotations,
		}
		if _, err := v.context.Graph.AddField(createFieldReq); err != nil {
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
	for _, field := range params.List {
		var constraintRef metadata.TypeRef
		if field.Type != nil {
			// Parse constraint using same env (constraints can refer to earlier params)
			// Note we're using a distinct, non-materializing usage visitor here to avoid polluting the HIR.
			// The rationale for not having this as a mutable flag is to preserve forward compatibility with threading.
			typeUsage, err := v.nonMaterializingTypeUsageVisitor.VisitExpr(pkg, file, field.Type, env)
			if err != nil {
				return nil, nil, err
			}
			constraintRef = typeUsage.Root
		}
		// assign constraint to every name in this field (usually single)
		for range field.Names {
			decls[idx].Constraint = constraintRef
			idx++
		}
	}

	return env, decls, nil

}
