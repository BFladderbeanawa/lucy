package upstream

import "github.com/mclucy/lucy/types"

type SupportedPlatformsReporter interface {
	SupportedPlatforms() []types.Platform
}

type DependencyResolver interface {
	ResolveDependency() []types.Dependency
}

type ArtifactMapper interface {
	NameByHash(artifact Hashable) types.RemotePackageName
	VersionedRefByHash(artifact Hashable) types.PackageId
}

type Hashable interface{}

type ArtifactResolver interface {
	ResolveArtifact() ResolvedArtifact
}

type ResolvedArtifact struct {
	Ref           types.PackageRef
	Version       types.BareVersion
	Source        types.Source
	FileURL       string
	Filename      string
	Hash          string
	HashAlgorithm string
}

type VersionSelectorResolver interface{}
