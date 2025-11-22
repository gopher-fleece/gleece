package metadata

type AliasKind int

const (
	AliasKindTypedef AliasKind = iota
	AliasKindAssigned
)

type AliasMeta struct {
	SymNodeMeta
	AliasType AliasKind
	Type      TypeUsageMeta
}
