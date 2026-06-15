package api

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestEnvironmentSpecJSONStable(t *testing.T) {
	spec := EnvironmentSpec{
		ID:               "survival",
		MinecraftVersion: "1.21.8",
		JavaVersion:      "21",
		LoaderType:       "fabric",
		LoaderVersion:    "0.16.14",
		ServerCore:       "paper",
		CarpetRequired:   true,
		MCDRRequired:     true,
		RuntimeProfileID: "prod",
		Packages: []PackageRef{{
			ID:                "fabric-api",
			Source:            "modrinth",
			Name:              "fabric-api",
			VersionConstraint: ">=0.100.0",
			MinecraftVersion:  "1.21.8",
			Loader:            "fabric",
			Required:          true,
			Metadata:          Metadata{"channel": "release"},
		}},
		LocalArtifacts: []LocalArtifactRef{{
			ArtifactID:       "local-config",
			PayloadAlgorithm: "sha256",
			PayloadHash:      "abc123",
			PayloadSize:      42,
			ArtifactType:     "config",
			RuntimeName:      "config/server.properties",
			Metadata:         Metadata{"owner": "stratum"},
		}},
		Metadata: Metadata{"tenant": "example"},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	const want = `{"id":"survival","minecraft_version":"1.21.8","java_version":"21","loader_type":"fabric","loader_version":"0.16.14","server_core":"paper","carpet_required":true,"mcdr_required":true,"runtime_profile_id":"prod","packages":[{"id":"fabric-api","source":"modrinth","name":"fabric-api","version_constraint":"\u003e=0.100.0","minecraft_version":"1.21.8","loader":"fabric","required":true,"metadata":{"channel":"release"}}],"local_artifacts":[{"artifact_id":"local-config","payload_algorithm":"sha256","payload_hash":"abc123","payload_size":42,"artifact_type":"config","runtime_name":"config/server.properties","metadata":{"owner":"stratum"}}],"metadata":{"tenant":"example"}}`
	if string(data) != want {
		t.Fatalf("unexpected JSON:\nwant %s\n got %s", want, data)
	}

	var got EnvironmentSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(got, spec) {
		t.Fatalf("round trip mismatch:\nwant %#v\n got %#v", spec, got)
	}
}

func TestPackageRefJSONStable(t *testing.T) {
	ref := PackageRef{
		ID:                "spark",
		Source:            "modrinth",
		Name:              "spark",
		VersionConstraint: "latest",
		MinecraftVersion:  "1.21.8",
		Loader:            "paper",
		Required:          true,
		Metadata:          Metadata{"side": "server"},
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	const want = `{"id":"spark","source":"modrinth","name":"spark","version_constraint":"latest","minecraft_version":"1.21.8","loader":"paper","required":true,"metadata":{"side":"server"}}`
	if string(data) != want {
		t.Fatalf("unexpected JSON:\nwant %s\n got %s", want, data)
	}

	var got PackageRef
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(got, ref) {
		t.Fatalf("round trip mismatch:\nwant %#v\n got %#v", ref, got)
	}
}

func TestLocalArtifactRefRejectsTraversalRuntimeName(t *testing.T) {
	badNames := []string{"../server.jar", "config/../../server.jar", filepath.Join("..", "server.jar")}
	for _, name := range badNames {
		err := ValidateLocalArtifactRef(LocalArtifactRef{RuntimeName: name})
		if err == nil {
			t.Fatalf("expected %q to be rejected", name)
		}
	}
}

func TestNoopResolverCapabilitiesDeterministic(t *testing.T) {
	resolver := NewNoopResolver()
	first, err := resolver.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("capabilities failed: %v", err)
	}
	second, err := resolver.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("capabilities failed: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("capabilities are not deterministic:\nfirst %#v\nsecond %#v", first, second)
	}
	if !first.SupportsPlan || !first.SupportsLock || !first.SupportsStatus {
		t.Fatalf("expected plan, lock, and status support: %#v", first)
	}
}

func TestNoopResolverPlanLockStatusDoNotWriteFiles(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	resolver := NewNoopResolver()
	env := EnvironmentSpec{ID: "env", MinecraftVersion: "1.21.8"}
	plan, err := resolver.PlanEnvironment(context.Background(), PlanRequest{Environment: env})
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	lock, err := resolver.LockEnvironment(context.Background(), LockRequest{Environment: env, Plan: plan})
	if err != nil {
		t.Fatalf("lock failed: %v", err)
	}
	status, err := resolver.CheckStatus(context.Background(), StatusRequest{Environment: env, Lock: lock})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !status.OK {
		t.Fatalf("expected ok status: %#v", status)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("noop resolver created files: %s", strings.Join(names, ", "))
	}
}

func TestPublicAPIDoesNotImportProviderPackages(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		text := string(data)
		if strings.Contains(text, "/internal/") || strings.Contains(text, "upstream/providers") {
			t.Fatalf("%s imports or references internal provider packages", entry.Name())
		}
	}
}
