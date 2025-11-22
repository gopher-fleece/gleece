package arbitrators

type PackageFacadeConfig struct {
	Globs                    []string `json:"globs"`
	AllowPackageLoadFailures bool     `json:"failOnAnyPackageLoadError"`
}
