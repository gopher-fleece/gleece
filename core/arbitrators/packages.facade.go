package arbitrators

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	MapSet "github.com/deckarep/golang-set/v2"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	"golang.org/x/tools/go/packages"
)

type PackagesFacade struct {
	config PackageFacadeConfig

	fileSet       *token.FileSet
	files         map[string]*ast.File         // filename → *ast.File
	fileToPackage map[string]*packages.Package // filename → owning *packages.Package

	packagesCache  map[string]*packages.Package // pkgPath → *packages.Package
	packageToFiles map[string][]*ast.File       // pkgPath → []*ast.File

}

func NewPackagesFacade(config PackageFacadeConfig) (PackagesFacade, error) {
	facade := PackagesFacade{
		config:  config,
		fileSet: token.NewFileSet(),

		files:         make(map[string]*ast.File),
		fileToPackage: make(map[string]*packages.Package),

		packagesCache:  make(map[string]*packages.Package),
		packageToFiles: map[string][]*ast.File{},
	}

	err := facade.initWithGlobs()
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

func (facade *PackagesFacade) initWithGlobs() error {

	pkgPathsToLoad := MapSet.NewSet[string]()

	// Keep track of real, absolute paths returned by the globs.
	// Globs return actual files and package loading of specific files sometimes
	// yields weird ephemeral packages with names like "command-line-arguments" which
	// mess with type recognition later on.
	//
	// To mitigate this, we load the entire package directory (probably not ideal),
	// and keep track of which files were actually matched by the globs so when we cache Package<==>AST Files, we can ignore any un-requested file
	matchedAbsPaths := map[string]struct{}{}

	// For each glob expression (provided via gleece.config), parse all matching files
	for _, globExpr := range facade.config.Globs {
		globbedSources, err := doublestar.FilepathGlob(globExpr)
		if err != nil {
			return err
		}

		for _, sourceFile := range globbedSources {
			absSourcePath, err := filepath.Abs(sourceFile)
			matchedAbsPaths[absSourcePath] = struct{}{}

			if err != nil {
				return err
			}

			file, err := parser.ParseFile(facade.fileSet, absSourcePath, nil, parser.ParseComments)
			if err != nil {
				logger.Error("Error parsing file %s - %v", absSourcePath, err)
				return fmt.Errorf("failed to parse file '%s' - parser raised error: %w", absSourcePath, err)
			}

			pkgPath, err := gast.GetFileFullPath(file, facade.fileSet)
			if err != nil {
				return err
			}
			if pkgPath == "" {
				return fmt.Errorf("could not determine PkgPath for file %v", file.Name.Name)
			}

			pkgPathsToLoad.Add(filepath.Dir(pkgPath))
		}
	}

	err := facade.loadPackagesFiltered(pkgPathsToLoad.ToSlice(), matchedAbsPaths)
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

func (facade *PackagesFacade) loadPackagesFiltered(
	packageExpressions []string,
	relevantFiles map[string]struct{},
) error {
	if len(packageExpressions) <= 0 {
		return nil
	}

	_, missingExpressions := facade.getCachedPkgsAndUnCachedExpressions(packageExpressions)

	var err error
	if len(missingExpressions) > 0 {
		_, err = facade.loadAndCacheExpressions(missingExpressions, relevantFiles)
	}

	return err
}

func (facade *PackagesFacade) loadAndCacheExpressions(
	packageExpressions []string,
	relevantFiles map[string]struct{},
) ([]*packages.Package, error) {
	// We're using LoadAllSyntax here which probably tanks performance.
	// Should improve, at a later point
	cfg := &packages.Config{Mode: packages.LoadAllSyntax, Fset: facade.fileSet}
	matchingPackages, err := packages.Load(cfg, packageExpressions...)
	if err != nil {
		return nil, err
	}

	var pkgErrs []packages.Error
	var failedPackageNames []string

	failedPkgCount := 0
	for _, p := range matchingPackages {
		if len(p.Errors) > 0 {
			failedPkgCount++
			pkgErrs = append(pkgErrs, p.Errors...)
			failedPackageNames = append(failedPackageNames, p.Name)
		}
	}

	if len(pkgErrs) > 0 {
		failMessage := fmt.Sprintf(
			"Failed to fully load the following packages: %v. Reported errors: %v",
			failedPackageNames,
			pkgErrs,
		)

		if facade.config.AllowPackageLoadFailures {
			logger.Warn(failMessage)
		} else {
			logger.Fatal(failMessage)
			return nil, fmt.Errorf(
				"encountered %d errors over %d package/s (%v) during load - %v",
				len(pkgErrs),
				failedPkgCount,
				failedPackageNames,
				pkgErrs,
			)
		}
	}

	// Note that packages.Load does *not* guarantee order
	for _, pkg := range matchingPackages {
		facade.cachePackage(pkg, relevantFiles)
	}

	return matchingPackages, nil
}

func (facade *PackagesFacade) cachePackage(pkg *packages.Package, relevantFiles map[string]struct{}) {
	facade.packagesCache[pkg.PkgPath] = pkg
	for _, file := range pkg.Syntax {
		if file == nil {
			continue
		}

		pos := pkg.Fset.Position(file.Package)
		if !pos.IsValid() {
			logger.Warn("invalid file position in package %s", pkg.Name)
			continue
		}

		absPath, err := filepath.Abs(pos.Filename)
		if err != nil {
			logger.Warn("could not get absolute path for %s", pos.Filename)
			continue
		}

		// If we're given a relevantFiles filter, load only files that match said filter
		if relevantFiles != nil {
			if _, ok := relevantFiles[absPath]; !ok {
				continue
			}
		}

		// Load and register the file and package
		facade.registerParsedFile(absPath, file, pkg)
	}
}

func (facade *PackagesFacade) GetPackages(packageExpressions []string) ([]*packages.Package, error) {
	matches, missingExpressions := facade.getCachedPkgsAndUnCachedExpressions(packageExpressions)

	if len(missingExpressions) <= 0 {
		return matches, nil
	}

	pkgs, err := facade.loadAndCacheExpressions(missingExpressions, nil)
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

func (facade *PackagesFacade) GetPackageForFile(file *ast.File) (*packages.Package, error) {
	name := gast.GetAstFileName(facade.fileSet, file)
	if name == "" {
		return nil, fmt.Errorf("could not determine name for given AST file - possibly missing from FileSet")
	}

	return facade.fileToPackage[name], nil
}
