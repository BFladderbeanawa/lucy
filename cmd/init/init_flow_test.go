package init

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mclucy/lucy/state"
	"github.com/mclucy/lucy/types"
)

func TestNewInitFlowState_EmptyWorkDir(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewInitFlowState(tmpDir)

	if len(s.ExistingFiles) != 0 {
		t.Errorf("expected no existing files, got %v", s.ExistingFiles)
	}
}

func TestNewInitFlowState_WithExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, string(state.ConfigFile))
	cfg := state.ConfigDefaults()
	if err := state.WriteConfig(tmpDir, &cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	s := NewInitFlowState(tmpDir)

	if !slices.Contains(s.ExistingFiles, string(state.ConfigFile)) {
		t.Errorf("expected ExistingFiles to contain %s, got %v", state.ConfigFile, s.ExistingFiles)
	}
}

func TestBuildResult_PreserveExisting(t *testing.T) {
	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, string(state.ConfigFile))
	cfg := state.ConfigDefaults()
	if err := state.WriteConfig(tmpDir, &cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	s := NewInitFlowState(tmpDir)
	s.GameVersion = "1.21.4"
	s.Platform = "none"
	s.ConflictResolution = PreserveExisting
	s.ManagedRoots = []string{"mods", "plugins"}

	result, err := BuildResult(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ConfigToWrite != nil {
		t.Error("expected ConfigToWrite to be nil when preserving existing")
	}

	if !slices.Contains(result.SkippedFiles, string(state.ConfigFile)) {
		t.Errorf("expected SkippedFiles to contain %s, got %v", state.ConfigFile, result.SkippedFiles)
	}
}

func TestBuildResult_AbortOnConflict(t *testing.T) {
	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, string(state.ConfigFile))
	cfg := state.ConfigDefaults()
	if err := state.WriteConfig(tmpDir, &cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	s := NewInitFlowState(tmpDir)
	s.GameVersion = "1.21.4"
	s.Platform = "none"
	s.ConflictResolution = AbortOnConflict
	s.ManagedRoots = []string{"mods", "plugins"}

	_, err := BuildResult(s)
	if err == nil {
		t.Error("expected error when aborting on conflict with existing files")
	}

	conflictErr, ok := err.(*ErrConflict)
	if !ok {
		t.Fatalf("expected ErrConflict, got %T", err)
	}
	if conflictErr.Mode != AbortOnConflict {
		t.Errorf("expected mode AbortOnConflict, got %v", conflictErr.Mode)
	}
}

func TestBuildResult_OverwriteAll(t *testing.T) {
	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, string(state.ConfigFile))
	cfg := state.ConfigDefaults()
	if err := state.WriteConfig(tmpDir, &cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	s := NewInitFlowState(tmpDir)
	s.GameVersion = "1.21.4"
	s.Platform = "none"
	s.ConflictResolution = OverwriteAll
	s.ManagedRoots = []string{"mods", "plugins"}

	result, err := BuildResult(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ConfigToWrite == nil {
		t.Error("expected ConfigToWrite to be set when OverwriteAll")
	}

	if !slices.Contains(result.WrittenFiles, string(state.ConfigFile)) {
		t.Errorf("expected WrittenFiles to contain %s, got %v", state.ConfigFile, result.WrittenFiles)
	}
}

func TestBuildResult_PersistsCompatiblePlatforms(t *testing.T) {
	tmpDir := t.TempDir()
	s := NewInitFlowState(tmpDir)
	s.GameVersion = "1.21.4"
	s.Platform = "neoforge"
	s.PlatformVersion = "21.1.0"
	s.CompatiblePlatforms = []string{"fabric", "mcdr", "sinytra"}
	s.ManagedRoots = []string{"mods", "plugins"}

	result, err := BuildResult(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ManifestToWrite == nil {
		t.Fatal("expected manifest to be written")
	}
	want := []string{"fabric", "mcdr", "sinytra"}
	if len(result.ManifestToWrite.Environment.CompatiblePlatforms) != len(want) {
		t.Fatalf("expected %d compatible platforms, got %d", len(want), len(result.ManifestToWrite.Environment.CompatiblePlatforms))
	}
	for i, platform := range want {
		if result.ManifestToWrite.Environment.CompatiblePlatforms[i] != platform {
			t.Fatalf("compatible platform %d mismatch: got %q want %q", i, result.ManifestToWrite.Environment.CompatiblePlatforms[i], platform)
		}
	}
}

func TestCanProceed_EmptyGameVersion(t *testing.T) {
	s := &InitFlowState{
		GameVersion:  "",
		Platform:     "none",
		ManagedRoots: []string{"mods", "plugins"},
	}

	if CanProceed(s) {
		t.Error("expected CanProceed to return false when GameVersion is empty")
	}
}

func TestCanProceed_EmptyManagedRoots(t *testing.T) {
	s := &InitFlowState{
		GameVersion:  "1.21.4",
		Platform:     "none",
		ManagedRoots: nil,
	}

	if CanProceed(s) {
		t.Error("expected CanProceed to return false when ManagedRoots is empty")
	}
}

func TestCanProceed_ValidState(t *testing.T) {
	s := &InitFlowState{
		GameVersion:  "1.21.4",
		Platform:     "none",
		ManagedRoots: []string{"mods", "plugins"},
	}

	if !CanProceed(s) {
		t.Error("expected CanProceed to return true for valid state")
	}
}

func TestCanProceed_ValidWithMultipleRoots(t *testing.T) {
	s := &InitFlowState{
		GameVersion:  "1.21.4",
		Platform:     "none",
		ManagedRoots: []string{"mods", "plugins", "config"},
	}

	if !CanProceed(s) {
		t.Error("expected CanProceed to return true for valid state with multiple roots")
	}
}

func TestConflictMode_String(t *testing.T) {
	tests := []struct {
		mode ConflictMode
		want string
	}{
		{PreserveExisting, "preserve"},
		{AbortOnConflict, "abort"},
		{OverwriteAll, "overwrite"},
	}

	for _, tc := range tests {
		if string(tc.mode) != tc.want {
			t.Errorf("expected %q, got %q", tc.want, tc.mode)
		}
	}
}

func TestDiscoverServerDefaults_UsesProbeObservedTakeoverCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	copyFile(
		t,
		filepath.Join("..", "..", "probe", "internal", "detector", "testdata", "fabric", "fabric-server-launch.jar"),
		filepath.Join(tmpDir, "fabric-server-launch.jar"),
	)

	defaults := DiscoverServerDefaults(tmpDir)

	if defaults.Platform != string(types.PlatformFabric) {
		t.Fatalf("expected platform %q, got %q", types.PlatformFabric, defaults.Platform)
	}
	if defaults.PlatformVersion == "" {
		t.Fatal("expected probe-derived platform version to be populated")
	}
	if !slices.Contains(defaults.ManagedRoots, "mods") {
		t.Fatalf("expected mods managed root from probe topology, got %v", defaults.ManagedRoots)
	}
	if !slices.Contains(defaults.DetectedPackages, "fabric/fabric") {
		t.Fatalf("expected runtime identity takeover candidate in detected packages, got %v", defaults.DetectedPackages)
	}
	if defaults.Confidence == ConfidenceNone {
		t.Fatal("expected non-empty discovery confidence")
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

func TestValidatePlatformSelectionRejectsImpossibleCombination(t *testing.T) {
	err := ValidatePlatformSelection("fabric", []string{"sinytra"})
	if err == nil {
		t.Fatal("expected impossible platform combination to fail")
	}
}

func TestBuildSummaryShowsPrimaryAndCompatiblePlatforms(t *testing.T) {
	s := &InitFlowState{
		GameVersion:         "1.21.4",
		Platform:            "neoforge",
		PlatformVersion:     "21.1.0",
		CompatiblePlatforms: []string{"fabric", "mcdr", "sinytra"},
		ManagedRoots:        []string{"mods", "plugins"},
		ConflictResolution:  PreserveExisting,
	}

	summary := buildSummary(s)
	if want := "Primary runtime: neoforge"; !containsLine(summary, want) {
		t.Fatalf("expected summary to contain %q, got:\n%s", want, summary)
	}
	if want := "Compatible with: fabric, mcdr, sinytra"; !containsLine(summary, want) {
		t.Fatalf("expected summary to contain %q, got:\n%s", want, summary)
	}
}

func containsLine(text, want string) bool {
	for line := range strings.SplitSeq(text, "\n") {
		if line == want {
			return true
		}
	}
	return false
}
