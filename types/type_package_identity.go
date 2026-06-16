package types

import (
	"errors"
	"fmt"
)

// PackageIdentity identifies one concrete package version after provider
// resolution. It is distinct from a user request and from a downloadable file.
type PackageIdentity struct {
	Source           SourceId
	Name             BarePackageName
	Loader           PlatformId
	MinecraftVersion BareVersion
	Version          BareVersion
}

func (p PackageIdentity) Key() string {
	return p.Source.String() + ":" + p.Name.String() + ":" +
		p.Loader.String() + ":" + p.MinecraftVersion.String() + ":" +
		p.Version.String()
}

func (p PackageIdentity) String() string {
	return p.Key()
}

func (p PackageIdentity) Validate() error {
	if p.Source == SourceAuto || p.Source == SourceUnknown {
		return errors.New("package identity source is required")
	}
	if p.Name == "" {
		return errors.New("package identity name is required")
	}
	if p.Version == "" || p.Version.IsInvalid() {
		return errors.New("package identity version is required")
	}
	return nil
}

// PackageRequest describes a requested package before provider resolution.
type PackageRequest struct {
	Source            SourceId
	Name              BarePackageName
	VersionConstraint BareVersion
	Loader            PlatformId
	MinecraftVersion  BareVersion
	Required          bool
	Metadata          map[string]string
}

func (p PackageRequest) Validate() error {
	if p.Source == SourceAuto || p.Source == SourceUnknown {
		return errors.New("package request source is required")
	}
	if p.Name == "" {
		return errors.New("package request name is required")
	}
	return nil
}

// ResolvedPackage is provider output metadata. Installation and download remain
// separate concerns even when a download URL or source reference is available.
type ResolvedPackage struct {
	Identity        PackageIdentity
	ResolvedVersion BareVersion
	Provider        SourceId
	DownloadURL     string
	SourceRef       string
	Checksum        string
	Dependencies    []Dependency
	Metadata        map[string]string
}

func (p ResolvedPackage) Validate() error {
	if err := p.Identity.Validate(); err != nil {
		return fmt.Errorf("resolved package identity: %w", err)
	}
	if p.ResolvedVersion == "" || p.ResolvedVersion.IsInvalid() {
		return errors.New("resolved package version is required")
	}
	return nil
}
