package metadata

import (
	"github.com/gopher-fleece/gleece/core/annotations"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
)

type ControllerMeta struct {
	Struct    StructMeta
	Receivers []ReceiverMeta
}

func (m ControllerMeta) Reduce(
	gleeceConfig *definitions.GleeceConfig,
	metaCache MetaCache,
	syncedProvider IdProvider,
) (definitions.ControllerMetadata, error) {
	// Parse any explicit Security annotations
	security, err := GetSecurityFromContext(m.Struct.Annotations)
	if err != nil {
		return definitions.ControllerMetadata{}, err
	}

	// If there are no explicitly defined securities, check for inherited ones
	if len(security) <= 0 {
		logger.Debug("Controller %s does not have explicit security; Using user-defined defaults", m.Struct.Name)
		security = GetDefaultSecurity(gleeceConfig)
	}

	var reducedReceivers []definitions.RouteMetadata
	for _, rec := range m.Receivers {
		reduced, err := rec.Reduce(metaCache, syncedProvider, security)
		if err != nil {
			logger.Error("Failed to reduce receiver '%s' of controller '%s' - %w", rec.Name, m.Struct.Name, err)
			return definitions.ControllerMetadata{}, err
		}
		reducedReceivers = append(reducedReceivers, reduced)
	}

	meta := definitions.ControllerMetadata{
		Name:        m.Struct.Name,
		PkgPath:     m.Struct.PkgPath,
		Tag:         m.Struct.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationTag),
		Description: m.Struct.Annotations.GetDescription(),
		RestMetadata: definitions.RestMetadata{
			Path: m.Struct.Annotations.GetFirstValueOrEmpty(annotations.GleeceAnnotationRoute),
		},
		Routes:   reducedReceivers,
		Security: security,
	}

	return meta, nil
}
