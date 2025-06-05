package caching

import (
	"go/ast"

	"github.com/gopher-fleece/gleece/definitions"
)

type SymbolMetadataCache struct {
	funcs           map[*ast.FuncDecl]*definitions.RouteMetadata
	types           map[*ast.TypeSpec]*definitions.TypeMetadata
	receiverParams  map[*ast.Ident]*definitions.FuncParam
	receiverRetVals map[*ast.Ident]*definitions.FuncReturnValue
}

func (c *SymbolMetadataCache) GetRouteMeta(funcDecl *ast.FuncDecl) *definitions.RouteMetadata {
	return c.funcs[funcDecl]
}

func (c *SymbolMetadataCache) GetTypeMeta(typeSpec *ast.TypeSpec) *definitions.TypeMetadata {
	return c.types[typeSpec]
}

func (c *SymbolMetadataCache) GetReceiverParam(ident *ast.Ident) *definitions.FuncParam {
	return c.receiverParams[ident]
}

func (c *SymbolMetadataCache) GetReceiverRetVal(ident *ast.Ident) *definitions.FuncReturnValue {
	return c.receiverRetVals[ident]
}

func (c *SymbolMetadataCache) PutRouteMeta(funcDecl *ast.FuncDecl, newValue *definitions.RouteMetadata) {
	existingValue := c.funcs[funcDecl]
	if existingValue == nil {
		c.funcs[funcDecl] = newValue
		return
	}

	if existingValue.FVersion.Equals(newValue.FVersion) {
		return
	}

	c.evictParamsIfNeeded(existingValue, newValue)
	c.evictRetValsIfNeeded(existingValue, newValue)

	// Finally update the route metadata cache with new value
	c.funcs[funcDecl] = newValue
}

func (c *SymbolMetadataCache) PutTypeMeta(typeSpec *ast.TypeSpec) {
	// Todo
}

func (c *SymbolMetadataCache) PutReceiverParam(ident *ast.Ident) {
	// Todo
}

func (c *SymbolMetadataCache) PutReceiverRetVal(ident *ast.Ident) {
	// Todo
}

func (c *SymbolMetadataCache) evictParamsIfNeeded(oldMeta *definitions.RouteMetadata, newMeta *definitions.RouteMetadata) {
	// Evict dependent parameters whose type/version changed or were removed
	for _, oldParam := range oldMeta.FuncParams {
		evict := true
		for _, newParam := range newMeta.FuncParams {
			if paramEquals(oldParam, newParam) {

				evict = false
				break
			}
		}
		if evict {
			delete(c.receiverParams, oldParam.Ident)
		}
	}
}

func (c *SymbolMetadataCache) evictRetValsIfNeeded(oldMeta *definitions.RouteMetadata, newMeta *definitions.RouteMetadata) {
	// Evict dependent parameters whose type/version changed or were removed
	for _, oldResponse := range oldMeta.Responses {
		evict := true
		for _, newResponse := range newMeta.Responses {
			if !oldResponse.TypeMetadata.Equals(newResponse.TypeMetadata) {
				evict = false
				break
			}
		}
		if evict {
			delete(c.receiverRetVals, oldResponse.Ident)
		}
	}
}

func paramEquals(a definitions.FuncParam, b definitions.FuncParam) bool {
	return a.Name == b.Name && a.NameInSchema == b.NameInSchema && a.TypeMeta.Equals(b.TypeMeta)
}
