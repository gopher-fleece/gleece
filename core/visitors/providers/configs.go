package providers

import (
	"github.com/gopher-fleece/gleece/core/arbitrators"
	"github.com/gopher-fleece/gleece/definitions"
)

type ArbitrationProviderConfig struct {
	arbitrators.PackageFacadeConfig
}

func NewArbitrationProviderConfig(gleeceConfig *definitions.GleeceConfig) ArbitrationProviderConfig {
	globs := []string{"./*.go", "./**/*.go"}
	allowPackageLoadFailures := false

	if gleeceConfig != nil {
		if len(gleeceConfig.CommonConfig.ControllerGlobs) > 0 {
			globs = gleeceConfig.CommonConfig.ControllerGlobs
		}

		allowPackageLoadFailures = gleeceConfig.CommonConfig.AllowPackageLoadFailures
	}

	return ArbitrationProviderConfig{
		PackageFacadeConfig: arbitrators.PackageFacadeConfig{
			Globs:                    globs,
			AllowPackageLoadFailures: allowPackageLoadFailures,
		},
	}
}
