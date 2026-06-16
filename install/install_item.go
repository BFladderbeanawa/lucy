package install

import (
	"strings"

	"github.com/mclucy/lucy/syntax"
	"github.com/mclucy/lucy/types"
)

// PackageOptions carries per-package install behavior options.
type PackageOptions struct {
	Optional bool
}

// InstallItem binds a ScopedPackageRef with a version constraint and per-package
// install options. This is the install core's input unit.
type InstallItem struct {
	Ref     types.ScopedPackageRef
	Version types.BareVersion
	Options PackageOptions
}

// ParseInstallItem parses a user-facing package string into an InstallItem.
func ParseInstallItem(s string, bareSource string, optional bool) (InstallItem, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	var ref types.PackageRef
	ref, err := syntax.ParsePackageRef(s)
	if err != nil {
		return InstallItem{}, err
	}

	var version types.BareVersion
	if len(strings.Split(s, "@")) > 1 {
		version = types.BareVersion(strings.Split(s, "@")[1])
	} else {
		version = types.VersionAny
	}

	scope := types.ParseSource(bareSource)

	return InstallItem{
		Ref: types.ScopedPackageRef{
			PackageRef: ref,
			Scope:      scope,
		},
		Version: version,
		Options: PackageOptions{
			Optional: optional,
		},
	}, nil
}
