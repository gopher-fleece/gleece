package utils

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gopher-fleece/gleece/cmd"
	"github.com/gopher-fleece/gleece/cmd/arguments"
	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/core/arbitrators/caching"
	"github.com/gopher-fleece/gleece/core/metadata"
	"github.com/gopher-fleece/gleece/core/metadata/typeref"
	"github.com/gopher-fleece/gleece/core/pipeline"
	"github.com/gopher-fleece/gleece/core/visitors"
	"github.com/gopher-fleece/gleece/core/visitors/providers"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
	"github.com/gopher-fleece/gleece/graphs/symboldg"
	. "github.com/onsi/ginkgo/v2"
	"golang.org/x/tools/go/packages"
)

func GetMetadataByRelativeConfig(relativeConfigPath string) (pipeline.GleeceFlattenedMetadata, error) {
	_, meta, err := cmd.GetConfigAndMetadata(
		arguments.CliArguments{
			ConfigPath: constructFullPathOrFail(relativeConfigPath, true),
		},
	)
	return meta, err
}

func GetConfigAndMetadataOrFail(relativeConfigPath string) (
	*definitions.GleeceConfig,
	pipeline.GleeceFlattenedMetadata,
) {
	config, meta, err := cmd.GetConfigAndMetadata(
		arguments.CliArguments{
			ConfigPath: constructFullPathOrFail(relativeConfigPath, true),
		},
	)

	if err != nil {
		Fail(fmt.Sprintf("GetConfigAndMetadata returned an error - %v", err))
	}

	return config, meta
}

func GetDefaultConfigAndMetadataOrFail() (
	*definitions.GleeceConfig,
	pipeline.GleeceFlattenedMetadata,
) {
	return GetConfigAndMetadataOrFail("gleece.test.config.json")
}

func GetMetadataByRelativeConfigOrFail(relativeConfigPath string) pipeline.GleeceFlattenedMetadata {

	meta, err := GetMetadataByRelativeConfig(relativeConfigPath)

	if err != nil {
		Fail(fmt.Sprintf("Could not generate routes - %v", err))
	}
	return meta
}

func GetDefaultMetadataOrFail() pipeline.GleeceFlattenedMetadata {
	_, meta := GetDefaultConfigAndMetadataOrFail()
	return meta
}

func constructFullPathOrFail(relativePath string, failIfNotExists bool) string {
	fullPath := filepath.Join(GetCwdOrFail(), relativePath)

	if failIfNotExists && !FileOrFolderExists(fullPath) {
		Fail(fmt.Sprintf("Path %s does not exist", fullPath))
	}

	return fullPath
}

func FileOrFolderExists(fullPath string) bool {
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadFileByRelativePathOrFail(relativePath string) string {
	filePath := constructFullPathOrFail(relativePath, true)
	data, err := os.ReadFile(filePath)
	if err != nil {
		Fail(fmt.Sprintf("Could not read file from '%s' - %v", filePath, err))
	}

	return string(data)
}

func WriteFileByRelativePathOrFail(relativePath string, data []byte) {
	filePath := constructFullPathOrFail(relativePath, false)
	_, err := os.Stat(filePath)
	if err != nil {
		dirPath := filepath.Dir(filePath)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			Fail(fmt.Sprintf("Could not mkdir %s - %v", dirPath, err))
		}

	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		Fail(fmt.Sprintf("Could not write to '%s' - %v", filePath, err))
	}
}

func GetCwdOrFail() string {
	cwd, err := os.Getwd()
	if err != nil {
		Fail(fmt.Sprintf("Could not determine process working directory - %v", err))
	}
	return cwd
}

func GetAbsPathByRelativeOrFail(relativePath string) string {
	return filepath.Join(GetCwdOrFail(), relativePath)
}

func DeleteDistInCurrentFolderOrFail() {
	DeleteRelativeFolderOrFail("dist")
}

func DeleteRelativeFolderOrFail(folder string) {
	cwd := GetCwdOrFail()
	delPath := GetAbsPathByRelativeOrFail(folder)
	if !IsSubpath(cwd, delPath) {
		Fail(fmt.Sprintf(
			"Requested a deletion of folder '%s' but it is not a sub-folder of CWD '%s'. Halting operation for your safety.",
			delPath,
			cwd,
		))
	}

	err := os.RemoveAll(delPath)
	if err != nil {
		Fail(fmt.Sprintf("Could not delete folder at '%s' - %v", delPath, err))
	}
}

// IsSubpath returns true if targetPath is inside baseDir
func IsSubpath(baseDir, targetPath string) bool {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		Fail(fmt.Sprintf("Cannot resolve base dir: %v", err))
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		Fail(fmt.Sprintf("Cannot resolve target path: %v", err))
	}

	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		Fail(fmt.Sprintf("Cannot compute relative path: %v", err))
	}

	// If the relative path starts with ".." (or is exactly "..") then it escapes baseDir
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}

	return true
}

func LoadPackageOrFail(fullName string, loadMode packages.LoadMode) *packages.Package {
	cfg := &packages.Config{Mode: loadMode}
	matchingPackages, err := packages.Load(cfg, fullName)
	if err != nil || len(matchingPackages) <= 0 {
		FailWithTestCodeError(fmt.Sprintf("Could not load package '%s' for testing", fullName))
	}

	return matchingPackages[0]
}

func LoadGleecePackageOrFail(loadMode packages.LoadMode) *packages.Package {
	return LoadPackageOrFail("github.com/gopher-fleece/gleece", loadMode)
}

func GetFunctionFromPackageOrFail(pkg *packages.Package, name string) *ast.FuncDecl {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if ok && funcDecl.Name.Name == name {
				return funcDecl
			}
		}
	}

	FailWithTestCodeError(fmt.Sprintf("Could not find function '%s' in package '%s'", name, pkg.Name))
	return nil
}

// FailWithTestCodeError Fails the test with a "This is a test issue, not a code issue" message.
// Used to signify something went wrong with test setup or such
func FailWithTestCodeError(message string) {
	Fail(fmt.Sprintf("%s. This indicates a test issue, not a code issue", message))
}

func GetAstFieldByNameOrFail(pkg *packages.Package, structName string, fieldName string) *ast.Field {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != structName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					FailWithTestCodeError(fmt.Sprintf("Type %q is not a struct", structName))
					return nil
				}

				for _, field := range structType.Fields.List {
					for _, name := range field.Names {
						if name.Name == fieldName {
							return field
						}
					}
				}

				FailWithTestCodeError(fmt.Sprintf("Field %q not found in struct %q", fieldName, structName))
				return nil
			}
		}
	}

	FailWithTestCodeError(fmt.Sprintf("Struct %q not found in package", structName))
	return nil
}

func GetPipelineOrFail() pipeline.GleecePipeline {
	configPath := constructFullPathOrFail("gleece.test.config.json", true)
	config, err := cmd.LoadGleeceConfig(configPath)
	if err != nil {
		Fail(fmt.Sprintf("could not load Gleece Config - %v", err))
	}

	pipe, err := pipeline.NewGleecePipeline(config)
	if err != nil {
		Fail(fmt.Sprintf("could not create a pipeline - %v", err))
	}

	return pipe
}

func GetVisitContextByRelativeConfigOrFail(relativeConfigPath string) visitors.VisitContext {
	configPath := constructFullPathOrFail(relativeConfigPath, true)
	config, err := cmd.LoadGleeceConfig(configPath)
	if err != nil {
		Fail(fmt.Sprintf("could not load Gleece Config - %v", err))
	}

	var globs []string
	if len(config.CommonConfig.ControllerGlobs) > 0 {
		globs = config.CommonConfig.ControllerGlobs
	} else {
		globs = []string{"./*.go", "./**/*.go"}
	}

	arbProvider, err := providers.NewArbitrationProvider(globs)
	if err != nil {
		Fail(fmt.Sprintf("could not create an arbitration provider - %v", err))
	}

	metaCache := caching.NewMetadataCache()
	symGraph := symboldg.NewSymbolGraph()

	return visitors.VisitContext{
		GleeceConfig:        config,
		ArbitrationProvider: arbProvider,
		MetadataCache:       metaCache,
		Graph:               &symGraph,
		SyncedProvider:      common.Ptr(providers.NewSyncedProvider()),
	}
}

// helper to create a *gast.FileVersion quickly for tests
func MakeFileVersion(id string, extraHashString string) *gast.FileVersion {
	return &gast.FileVersion{
		Path:    id,
		ModTime: time.Now(),
		Hash:    fmt.Sprintf("hash-%s-%s", id, extraHashString),
	}
}

// helper to create a simple ast.Ident node
func MakeIdent(name string) ast.Node {
	return &ast.Ident{Name: name, NamePos: token.NoPos}
}

func CommentsToCommentBlock(comments []string, callerStackDepth int) gast.CommentBlock {
	// This func also fakes some positions to allow for at-least cursory checks by tests
	fakeStartLine := 45

	commentNodes := make([]gast.CommentNode, len(comments))
	for index, comment := range comments {
		commentNodes[index] = gast.CommentNode{
			Text:  comment,
			Index: index,
			Position: gast.CommentPosition{
				StartLine: fakeStartLine + index,
				EndLine:   fakeStartLine + index,
				StartCol:  0,
				EndCol:    utf8.RuneCountInString(comment),
			},
		}
	}

	// Add file info, if available
	_, file, _, _ := runtime.Caller(callerStackDepth)

	rng := common.ResolvedRange{}
	if len(comments) > 0 {
		lastCommentNode := commentNodes[len(commentNodes)-1]

		rng = common.ResolvedRange{
			StartLine: fakeStartLine,
			EndLine:   lastCommentNode.Position.EndLine,
			StartCol:  0,
			EndCol:    lastCommentNode.Position.EndCol,
		}
	}

	return gast.CommentBlock{
		Comments: commentNodes,
		FileName: file,
		Range:    rng,
	}
}

func GetAnnotationHolderOrFail(comments []string, appliedOn annotations.CommentSource) *annotations.AnnotationHolder {
	holder, err := annotations.NewAnnotationHolder(CommentsToCommentBlock(comments, 2), appliedOn)
	if err != nil {
		Fail(fmt.Sprintf("Failed to create an annotation holder during testing - %v", err))
	}

	return &holder
}

func GetMockParams(number int) []metadata.FuncParam {
	params := make([]metadata.FuncParam, 0, number)
	for i := range number {
		params = append(
			params,
			metadata.FuncParam{
				SymNodeMeta: metadata.SymNodeMeta{
					Name: fmt.Sprintf("param%d", i),
				},
				Ordinal: i,
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name: "int",
					},
				},
			},
		)
	}
	return params
}

func GetMockRetVals(number int) []metadata.FuncReturnValue {
	retVals := make([]metadata.FuncReturnValue, 0, number)
	for i := range number {
		retVals = append(
			retVals,
			metadata.FuncReturnValue{
				SymNodeMeta: metadata.SymNodeMeta{
					Name: "",
				},
				Ordinal: i,
				Type: metadata.TypeUsageMeta{
					SymNodeMeta: metadata.SymNodeMeta{
						Name: "int",
					},
				},
			},
		)
	}
	return retVals
}

// MakeUniverseRoot is a tiny test helper that builds a NamedTypeRef pointing at a universe type.
// Note: typeref.NewNamedTypeRef returns a value so we take its address for Root.
func MakeUniverseRoot(universeName string) *typeref.NamedTypeRef {
	k := graphs.NewUniverseSymbolKey(universeName)
	r := typeref.NewNamedTypeRef(&k, nil)
	return &r
}
