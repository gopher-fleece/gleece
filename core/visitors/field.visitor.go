package visitors

import (
	"errors"
	"fmt"
	"go/ast"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	"golang.org/x/tools/go/packages"
)

type FieldVisitor struct {
	BaseVisitor

	typeUsageVisitor *TypeUsageVisitor
}

// NewFieldVisitor Instantiates a new type visitor.
func NewFieldVisitor(context *VisitContext) (*FieldVisitor, error) {
	visitor := FieldVisitor{}
	err := visitor.initialize(context)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a new FieldVisitor")
	}

	return &visitor, err
}

func (v *FieldVisitor) setTypeUsageVisitor(visitor *TypeUsageVisitor) {
	v.typeUsageVisitor = visitor
}

// VisitField returns one metadata.FieldMeta per declared "name" in the AST field node.
// The caller **must** pass the desired symbol kind (Parameter, RetType, Field).
// For anonymous/embedded fields we always return a single FieldMeta with Name == "".
// isEmbedded is set only when the AST declares an embedded field (len(field.Names)==0).
func (v *FieldVisitor) VisitField(
	pkg *packages.Package,
	file *ast.File,
	field *ast.Field,
	kind common.SymKind,
) ([]metadata.FieldMeta, error) {
	if field == nil {
		return nil, errors.New("nil field was provided to FieldVisitor.VisitField")
	}

	fileName := gast.GetAstFileNameOrFallback(file, nil)
	v.enterFmt("VisitField in %s — kind=%s expr=%T", fileName, kind, field.Type)
	defer v.exit()

	annotationsHolder, err := v.getAnnotations(field.Doc, nil)
	if err != nil {
		return nil, v.frozenError(err)
	}

	// canonical name list: ensure at least one entry for anonymous/embedded fields and unnamed returns
	names := gast.GetFieldNames(field)
	if len(names) == 0 {
		names = []string{""}
	}

	// delegate to type-usage visitor
	typeUsage, err := v.typeUsageVisitor.VisitExpr(pkg, file, field.Type, nil)
	if err != nil {
		return nil, v.frozenError(fmt.Errorf("type parsing failed: %w", err))
	}

	fileVersion, err := v.context.MetadataCache.GetFileVersion(file, pkg.Fset)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain FileVersion for '%s' - %v", fileName, err)
	}

	return v.buildFieldsMeta(
		pkg,
		field,
		names,
		kind,
		annotationsHolder,
		typeUsage,
		fileVersion,
	)
}

func (v *FieldVisitor) buildFieldsMeta(
	pkg *packages.Package,
	field *ast.Field,
	names []string,
	kind common.SymKind,
	annotationsHolder *annotations.AnnotationHolder,
	typeUsage metadata.TypeUsageMeta,
	fileVersion *gast.FileVersion,
) ([]metadata.FieldMeta, error) {
	// *ast.Field here can be a parameter, return value, a named field or an embedded field.
	isEmbedded := kind == common.SymKindField && gast.IsEmbeddedOrAnonymousField(field)

	out := make([]metadata.FieldMeta, 0, len(names))
	for _, name := range names {
		fieldMeta := metadata.FieldMeta{
			SymNodeMeta: metadata.SymNodeMeta{
				Name:        name,
				Node:        field,
				SymbolKind:  kind,
				PkgPath:     pkg.PkgPath,
				Annotations: annotationsHolder,
				FVersion:    fileVersion,
				Range:       common.ResolveNodeRange(pkg.Fset, field),
			},
			Type:       typeUsage,
			IsEmbedded: isEmbedded,
		}

		// Graph insertion (GraphBuilder expected to de-dupe)
		req := symboldg.CreateFieldNode{
			Data:        fieldMeta,
			Annotations: annotationsHolder,
		}
		if _, err := v.context.GraphBuilder.AddField(req); err != nil {
			return nil, v.frozenError(err)
		}

		out = append(out, fieldMeta)
	}

	return out, nil
}
