package upstream

import "github.com/mclucy/lucy/types"

type Searcher interface {
	Search(query string, options types.SearchOptions) []types.RemotePackageName
}

type Informer interface {
	Info(ref types.PackageRef) types.Metadata
}
