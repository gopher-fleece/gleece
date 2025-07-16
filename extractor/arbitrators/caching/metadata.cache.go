package caching

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gopher-fleece/gleece/extractor/metadata"
	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/graphs"
)

type MetadataCache struct {
	controllers map[graphs.SymbolKey]*metadata.ControllerMeta
	receivers   map[graphs.SymbolKey]*metadata.ReceiverMeta
	structs     map[graphs.SymbolKey]*metadata.StructMeta
	enums       map[graphs.SymbolKey]*metadata.EnumMeta
	visited     map[graphs.SymbolKey]struct{}

	FileVersions map[*ast.File]*gast.FileVersion
}

func NewMetadataCache() *MetadataCache {
	return &MetadataCache{
		controllers:  make(map[graphs.SymbolKey]*metadata.ControllerMeta),
		receivers:    make(map[graphs.SymbolKey]*metadata.ReceiverMeta),
		structs:      make(map[graphs.SymbolKey]*metadata.StructMeta),
		enums:        make(map[graphs.SymbolKey]*metadata.EnumMeta),
		visited:      make(map[graphs.SymbolKey]struct{}),
		FileVersions: map[*ast.File]*gast.FileVersion{},
	}
}

func (c *MetadataCache) GetStruct(key graphs.SymbolKey) *metadata.StructMeta {
	return c.structs[key]
}

func (c *MetadataCache) GetReceiver(key graphs.SymbolKey) *metadata.ReceiverMeta {
	return c.receivers[key]
}

func (c *MetadataCache) GetEnum(key graphs.SymbolKey) *metadata.EnumMeta {
	return c.enums[key]
}

func (c *MetadataCache) HasController(meta *metadata.ControllerMeta) bool {
	key := graphs.NewSymbolKey(meta.Struct.Node, meta.Struct.FVersion)
	return c.controllers[key] != nil
}

func (c *MetadataCache) HasReceiver(key graphs.SymbolKey) bool {
	return c.receivers[key] != nil
}

func (c *MetadataCache) HasStruct(key graphs.SymbolKey) bool {
	return c.structs[key] != nil
}

func (c *MetadataCache) HasEnum(meta *metadata.EnumMeta) bool {
	key := graphs.NewSymbolKey(meta.Node, meta.FVersion)
	return c.enums[key] != nil
}

func (c *MetadataCache) HasVisited(key graphs.SymbolKey) bool {
	_, ok := c.visited[key]
	return ok
}

func (c *MetadataCache) GetFileVersion(file *ast.File, fileSet *token.FileSet) (*gast.FileVersion, error) {
	fVersion := c.FileVersions[file]
	if fVersion != nil {
		return fVersion, nil
	}

	version, err := gast.NewFileVersionFromAstFile(file, fileSet)
	if err != nil {
		return nil, err
	}

	c.FileVersions[file] = &version
	return &version, nil

}

func (c *MetadataCache) AddController(meta *metadata.ControllerMeta) error {
	key := graphs.NewSymbolKey(meta.Struct.Node, meta.Struct.FVersion)
	return addEntity(key, c.controllers, c.visited, meta)
}

func (c *MetadataCache) AddReceiver(meta *metadata.ReceiverMeta) error {
	key := graphs.NewSymbolKey(meta.Node, meta.FVersion)
	return addEntity(key, c.receivers, c.visited, meta)
}

func (c *MetadataCache) AddStruct(meta *metadata.StructMeta) error {
	key := graphs.NewSymbolKey(meta.Node, meta.FVersion)
	return addEntity(key, c.structs, c.visited, meta)
}

func (c *MetadataCache) AddEnum(meta *metadata.EnumMeta) error {
	key := graphs.NewSymbolKey(meta.Node, meta.FVersion)
	return addEntity(key, c.enums, c.visited, meta)
}

func addEntity[TMeta any](
	key graphs.SymbolKey,
	cache map[graphs.SymbolKey]*TMeta,
	visitedCache map[graphs.SymbolKey]struct{},
	meta *TMeta,
) error {
	if _, exists := cache[key]; exists {
		return fmt.Errorf("key %v already exists in cache", key)
	}

	cache[key] = meta
	// Mark key as 'visited'
	visitedCache[key] = struct{}{}
	return nil
}
