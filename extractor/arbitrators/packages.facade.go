package arbitrators

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

type PackagesFacade struct {
	packagesCache map[string]*packages.Package
}

func NewPackagesFacade() PackagesFacade {
	return PackagesFacade{
		packagesCache: make(map[string]*packages.Package),
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

	expressionsToLookup := packageExpressions[:]

	for idx, expr := range packageExpressions {
		_, exists := facade.packagesCache[expr]
		if exists {
			// Already cached, drop the expression (i.e., don't look for it)
			expressionsToLookup = append(expressionsToLookup[:idx], expressionsToLookup[idx+1:]...)
		}
	}

	// We're using LoadAllSyntax here which probably tanks performance.
	// Should improve, at a later point
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	matchingPackages, err := packages.Load(cfg, expressionsToLookup...)
	if err != nil {
		return err
	}

	// Note that packages.Load does *not* guarantee order
	for _, pkg := range matchingPackages {
		facade.packagesCache[pkg.PkgPath] = pkg
	}

	return err
}

func (facade *PackagesFacade) GetPackages(packageExpressions []string) ([]*packages.Package, error) {
	var matches []*packages.Package
	expressionsToLookup := packageExpressions[:]

	for idx, expr := range packageExpressions {
		existing := facade.packagesCache[expr]
		if existing != nil {
			matches = append(matches, existing)
			expressionsToLookup = append(expressionsToLookup[:idx], expressionsToLookup[idx+1:]...)
		}
	}

	if len(expressionsToLookup) <= 0 {
		return matches, nil
	}

	// We're using LoadAllSyntax here which probably tanks performance.
	// Should improve, at a later point
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	matchingPackages, err := packages.Load(cfg, expressionsToLookup...)
	if err != nil {
		return append(matches, matchingPackages...), err
	}

	for idx, pkg := range matchingPackages {
		facade.packagesCache[packageExpressions[idx]] = pkg
	}

	// Return both cached and freshly fetched matches, if available.
	return append(matches, matchingPackages...), err
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
