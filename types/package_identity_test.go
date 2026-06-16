package types

import "testing"

func TestPackageIdentityKeyDeterministic(t *testing.T) {
	identity := PackageIdentity{
		Source:           SourceModrinth,
		Name:             "fabric-api",
		Loader:           PlatformFabric,
		MinecraftVersion: "1.21.8",
		Version:          "0.128.2",
	}

	const want = "modrinth:fabric-api:fabric:1.21.8:0.128.2"
	if identity.Key() != want {
		t.Fatalf("expected key %q, got %q", want, identity.Key())
	}
	if identity.String() != want {
		t.Fatalf("expected string %q, got %q", want, identity.String())
	}
}

func TestPackageRequestValidationCatchesMissingSourceName(t *testing.T) {
	if err := (PackageRequest{Name: "fabric-api"}).Validate(); err == nil {
		t.Fatal("expected missing source to be rejected")
	}

	if err := (PackageRequest{Source: SourceUnknown, Name: "fabric-api"}).Validate(); err == nil {
		t.Fatal("expected unknown source to be rejected")
	}

	if err := (PackageRequest{Source: SourceModrinth}).Validate(); err == nil {
		t.Fatal("expected missing name to be rejected")
	}

	if err := (PackageRequest{Source: SourceModrinth, Name: "fabric-api"}).Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestResolvedPackageValidationRequiresIdentityAndVersion(t *testing.T) {
	validIdentity := PackageIdentity{
		Source:  SourceModrinth,
		Name:    "fabric-api",
		Loader:  PlatformFabric,
		Version: "0.128.2",
	}

	if err := (ResolvedPackage{Identity: validIdentity}).Validate(); err == nil {
		t.Fatal("expected missing resolved version to be rejected")
	}

	if err := (ResolvedPackage{ResolvedVersion: "0.128.2"}).Validate(); err == nil {
		t.Fatal("expected missing identity to be rejected")
	}

	pkg := ResolvedPackage{Identity: validIdentity, ResolvedVersion: "0.128.2"}
	if err := pkg.Validate(); err != nil {
		t.Fatalf("expected valid resolved package, got %v", err)
	}
}

func TestExistingStringablePackageRefBehaviorStillWorks(t *testing.T) {
	var ref StringablePackageRef = VersionedPackageRef{
		Platform: PlatformFabric,
		Name:     "fabric-api",
		Version:  "0.128.2",
	}

	if ref.StringBase() != "fabric/fabric-api" {
		t.Fatalf("unexpected base ref: %q", ref.StringBase())
	}
	if ref.StringFull() != "fabric/fabric-api@0.128.2" {
		t.Fatalf("unexpected full ref: %q", ref.StringFull())
	}
}
