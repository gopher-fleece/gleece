package caching

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gopher-fleece/gleece/v2/core/metadata"
	"github.com/gopher-fleece/gleece/v2/gast"
	"github.com/gopher-fleece/gleece/v2/graphs"
)

type MetadataCache struct {
	controllers map[graphs.SymbolKey]*metadata.ControllerMeta
	receivers   map[graphs.SymbolKey]*metadata.ReceiverMeta
	structs     map[graphs.SymbolKey]*metadata.StructMeta
	enums       map[graphs.SymbolKey]*metadata.EnumMeta
	aliases     map[graphs.SymbolKey]*metadata.AliasMeta
	visited     map[graphs.SymbolKey]struct{}
	inProgress  map[graphs.SymbolKey]struct{}

	fileVersions map[*ast.File]*gast.FileVersion
}

func NewMetadataCache() *MetadataCache {
	return &MetadataCache{
		controllers:  make(map[graphs.SymbolKey]*metadata.ControllerMeta),
		receivers:    make(map[graphs.SymbolKey]*metadata.ReceiverMeta),
		structs:      make(map[graphs.SymbolKey]*metadata.StructMeta),
		enums:        make(map[graphs.SymbolKey]*metadata.EnumMeta),
		aliases:      make(map[graphs.SymbolKey]*metadata.AliasMeta),
		visited:      make(map[graphs.SymbolKey]struct{}),
		inProgress:   make(map[graphs.SymbolKey]struct{}),
		fileVersions: map[*ast.File]*gast.FileVersion{},
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

func (c *MetadataCache) GetAlias(key graphs.SymbolKey) *metadata.AliasMeta {
	return c.aliases[key]
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

func (c *MetadataCache) HasEnum(key graphs.SymbolKey) bool {
	return c.enums[key] != nil
}

func (c *MetadataCache) HasAlias(key graphs.SymbolKey) bool {
	return c.aliases[key] != nil
}

func (c *MetadataCache) HasVisited(key graphs.SymbolKey) bool {
	_, ok := c.visited[key]
	return ok
}

func (c *MetadataCache) GetFileVersion(file *ast.File, fileSet *token.FileSet) (*gast.FileVersion, error) {
	fVersion := c.fileVersions[file]
	if fVersion != nil {
		return fVersion, nil
	}

	version, err := gast.NewFileVersionFromAstFile(file, fileSet)
	if err != nil {
		return nil, err
	}

	c.fileVersions[file] = &version
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

func (c *MetadataCache) AddAlias(meta *metadata.AliasMeta) error {
	key := graphs.NewSymbolKey(meta.Node, meta.FVersion)
	return addEntity(key, c.aliases, c.visited, meta)
}

// StartMaterializing claims the key for materialization.
// Returns true if the caller should proceed (not visited, not already in-progress).
func (c *MetadataCache) StartMaterializing(key graphs.SymbolKey) bool {
	if _, seen := c.visited[key]; seen {
		return false // already done
	}
	if _, doing := c.inProgress[key]; doing {
		return false // someone else is already working on it
	}
	c.inProgress[key] = struct{}{}
	return true
}

// FinishMaterializing clears in-progress. If success==true, also mark visited.
func (c *MetadataCache) FinishMaterializing(key graphs.SymbolKey, success bool) {
	delete(c.inProgress, key)
	if success {
		c.visited[key] = struct{}{}
	}
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
