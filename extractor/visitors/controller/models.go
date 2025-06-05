package controller

import "github.com/gopher-fleece/gleece/definitions"

func (v *ControllerVisitor) addToTypeMap(
	existingTypesMap *map[string]string,
	existingModels *[]definitions.TypeMetadata,
	typeMeta definitions.TypeMetadata,
) error {
	if typeMeta.IsUniverseType {
		return nil
	}

	existsInPackage, exists := (*existingTypesMap)[typeMeta.Name]
	if exists {
		if existsInPackage == typeMeta.PkgPath {
			// Same type referenced from a separate location
			return nil
		}

		return v.getFrozenError(
			"type '%s' exists in more that one package (%s and %s). This is not currently supported",
			typeMeta.Name,
			typeMeta.PkgPath,
			existsInPackage,
		)
	}

	(*existingTypesMap)[typeMeta.Name] = typeMeta.PkgPath
	(*existingModels) = append((*existingModels), typeMeta)
	return nil
}

func (v *ControllerVisitor) insertRouteTypeList(
	existingTypesMap *map[string]string,
	existingModels *[]definitions.TypeMetadata,
	route *definitions.RouteMetadata,
) (bool, error) {

	plainErrorEncountered := false
	for _, param := range route.FuncParams {
		if param.TypeMeta.IsUniverseType && param.TypeMeta.Name == "error" && param.TypeMeta.PkgPath == "" {
			// Mark whether we've encountered any 'error' type
			plainErrorEncountered = true
		}
		err := v.addToTypeMap(existingTypesMap, existingModels, param.TypeMeta)
		if err != nil {
			return plainErrorEncountered, v.frozenError(err)
		}
	}

	for _, param := range route.Responses {
		if param.IsUniverseType && param.Name == "error" && param.PkgPath == "" {
			// Mark whether we've encountered any 'error' type
			plainErrorEncountered = true
		}
		err := v.addToTypeMap(existingTypesMap, existingModels, param.TypeMetadata)
		if err != nil {
			return plainErrorEncountered, v.frozenError(err)
		}
	}

	return plainErrorEncountered, nil
}
