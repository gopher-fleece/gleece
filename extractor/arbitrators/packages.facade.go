package arbitrators

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
)

type PackagesFacade struct {
	fileSet       *token.FileSet
	files         map[string]*ast.File         // filename → *ast.File
	fileToPackage map[string]*packages.Package // filename → owning *packages.Package

	packagesCache  map[string]*packages.Package // pkgPath → *packages.Package
	packageToFiles map[string][]*ast.File       // pkgPath → []*ast.File

}

func NewPackagesFacade(globs []string) (PackagesFacade, error) {
	facade := PackagesFacade{
		fileSet: token.NewFileSet(),

		files:         make(map[string]*ast.File),
		fileToPackage: make(map[string]*packages.Package),

		packagesCache:  make(map[string]*packages.Package),
		packageToFiles: map[string][]*ast.File{},
	}

	err := facade.initWithGlobs(globs)
	return facade, err
}

func (facade *PackagesFacade) FSet() *token.FileSet {
	return facade.fileSet
}

func (facade *PackagesFacade) GetAllSourceFiles() []*ast.File {
	result := make([]*ast.File, 0, len(facade.files))
	for _, file := range facade.files {
		result = append(result, file)
	}
	return result
}

func (facade *PackagesFacade) initWithGlobs(globs []string) error {

	pkgPathsToLoad := MapSet.NewSet[string]()

	// For each glob expression (provided via gleece.config), parse all matching files
	for _, globExpr := range globs {
		globbedSources, err := doublestar.FilepathGlob(globExpr)
		if err != nil {
			return err
		}

		for _, sourceFile := range globbedSources {
			absSourcePath, err := filepath.Abs(sourceFile)
			if err != nil {
				return err
			}

			file, err := parser.ParseFile(facade.fileSet, absSourcePath, nil, parser.ParseComments)
			if err != nil {
				logger.Error("Error parsing file %s - %v", absSourcePath, err)
				return err
			}

			pkgPath, err := gast.GetFileFullPath(file, facade.fileSet)
			if err != nil {
				return err
			}
			if pkgPath == "" {
				return fmt.Errorf("could not determine PkgPath for file %v", file.Name.Name)
			}

			pkgPathsToLoad.Add(pkgPath)
		}
	}

	err := facade.LoadPackages(pkgPathsToLoad.ToSlice())
	if err != nil {
		logger.Error("Could not load one or more packages (%v) - %v", pkgPathsToLoad.ToSlice(), err)
		return err
	}

	return nil
}

func (facade *PackagesFacade) registerParsedFile(absSourceFilePath string, file *ast.File, pkg *packages.Package) {
	if facade.files[absSourceFilePath] != nil {
		// Idempotency guard- packageToFiles and fileSet need that
		return
	}

	facade.files[absSourceFilePath] = file
	facade.fileToPackage[absSourceFilePath] = pkg
	facade.packageToFiles[pkg.PkgPath] = append(facade.packageToFiles[pkg.PkgPath], file)

	// Sync token.File from pkg.Fset into our shared fileSet
	pkgFile := pkg.Fset.File(file.Pos())
	if pkgFile != nil {

		facade.fileSet.AddFile(
			absSourceFilePath,   // name
			-1,                  // base (-1 means let it choose)
			int(pkgFile.Size()), // size
		)
	}
}

func (facade *PackagesFacade) GetAllPackages() []*packages.Package {
	allCached := []*packages.Package{}
	for _, value := range facade.packagesCache {
		allCached = append(allCached, value)
	}
	return allCached
}

func (facade *PackagesFacade) GetPackage(packageExpression string) (*packages.Package, error) {
	matches, err := facade.GetPackages([]string{packageExpression})
	if err != nil {
		return nil, err
	}

	if len(matches) > 0 {
		return matches[0], nil
	}

	return nil, nil
}

func (facade *PackagesFacade) LoadPackages(packageExpressions []string) error {

	if len(packageExpressions) <= 0 {
		return nil
	}

	_, missingExpressions := facade.getCachedPkgsAndUnCachedExpressions(packageExpressions)

	var err error
	if len(missingExpressions) > 0 {
		_, err = facade.loadAndCacheExpressions(missingExpressions)
	}

	return err
}

func (facade *PackagesFacade) loadAndCacheExpressions(packageExpressions []string) ([]*packages.Package, error) {
	// We're using LoadAllSyntax here which probably tanks performance.
	// Should improve, at a later point
	cfg := &packages.Config{Mode: packages.LoadAllSyntax, Fset: facade.fileSet}
	matchingPackages, err := packages.Load(cfg, packageExpressions...)
	if err != nil {
		return nil, err
	}

	// Note that packages.Load does *not* guarantee order
	for _, pkg := range matchingPackages {
		facade.cachePackage(pkg)
	}

	return matchingPackages, nil
}

func (facade *PackagesFacade) cachePackage(pkg *packages.Package) {
	facade.packagesCache[pkg.PkgPath] = pkg
	for _, file := range pkg.Syntax {
		if file == nil {
			continue
		}

		if pos := pkg.Fset.Position(file.Package); pos.IsValid() {
			if absPath, err := filepath.Abs(pos.Filename); err == nil {
				facade.registerParsedFile(absPath, file, pkg)
			} else {
				logger.Warn("could not determine absolute path of file %v in package %s", pos.Filename, pkg.Name)
			}
		} else {
			logger.Warn("could not determine position or position is invalid for one of package %s AST files", pkg.Name)
		}
	}
}

func (facade *PackagesFacade) GetPackages(packageExpressions []string) ([]*packages.Package, error) {
	matches, missingExpressions := facade.getCachedPkgsAndUnCachedExpressions(packageExpressions)

	if len(missingExpressions) <= 0 {
		return matches, nil
	}

	pkgs, err := facade.loadAndCacheExpressions(missingExpressions)
	if err != nil {
		return matches, err
	}

	// Return both cached and freshly fetched matches, if available.
	return append(matches, pkgs...), err
}

func (facade *PackagesFacade) getCachedPkgsAndUnCachedExpressions(
	allPkgExpressions []string,
) ([]*packages.Package, []string) {
	cachedPkgs := []*packages.Package{}
	expressionsToLookup := []string{}

	for _, expr := range allPkgExpressions {
		if pkg, exists := facade.packagesCache[expr]; exists {
			cachedPkgs = append(cachedPkgs, pkg)
		} else {
			expressionsToLookup = append(expressionsToLookup, expr)
		}
	}

	return cachedPkgs, expressionsToLookup
}

func (facade *PackagesFacade) EvictCache() {
	for k := range facade.packagesCache {
		delete(facade.packagesCache, k)
	}
}

func (facade *PackagesFacade) GetPackageNameByNamedEntity(namedEntity *types.Named) (string, error) {
	pkg := namedEntity.Obj().Pkg()
	if pkg == nil {
		return "", nil // Built-in types, unnamed types, etc.
	}

	// Find the package where the struct is defined
	pkgPath := pkg.Path()
	targetPkg, err := facade.GetPackage(pkgPath)
	if err != nil {
		return "", err // Package not found in loaded AST
	}

	return targetPkg.PkgPath, nil
}

func (facade *PackagesFacade) GetPackageByTypeName(typeName *types.TypeName) (*packages.Package, error) {
	pkg := typeName.Pkg()
	if pkg == nil {
		return nil, nil // Built-in types, unnamed types, etc.
	}

	return facade.GetPackage(pkg.Path())
}

func (facade *PackagesFacade) GetAstFile(absPath string) *ast.File {
	return facade.files[absPath]
}

func (facade *PackagesFacade) GetPackageForFileName(absPath string) *packages.Package {
	return facade.fileToPackage[absPath]
}

func (facade *PackagesFacade) GetPackageForFile(file *ast.File) (*packages.Package, error) {
	name := gast.GetAstFileName(facade.fileSet, file)
	if name == "" {
		return nil, fmt.Errorf("could not determine name for given AST file - possibly missing from FileSet")
	}

	return facade.fileToPackage[name], nil
}

func (facade *PackagesFacade) GetFilesForPackage(pkgPath string) []*ast.File {
	return facade.packageToFiles[pkgPath]
}
