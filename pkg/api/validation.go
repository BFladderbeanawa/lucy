package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func ValidateEnvironmentSpec(spec EnvironmentSpec) error {
	if strings.TrimSpace(spec.ID) == "" {
		return errors.New("environment id is required")
	}
	if strings.TrimSpace(spec.MinecraftVersion) == "" {
		return errors.New("environment minecraft_version is required")
	}

	for i, pkg := range spec.Packages {
		if err := ValidatePackageRef(pkg); err != nil {
			return fmt.Errorf("packages[%d]: %w", i, err)
		}
	}
	for i, artifact := range spec.LocalArtifacts {
		if err := ValidateLocalArtifactRef(artifact); err != nil {
			return fmt.Errorf("local_artifacts[%d]: %w", i, err)
		}
	}

	return nil
}

func ValidatePackageRef(ref PackageRef) error {
	if strings.TrimSpace(ref.ID) == "" {
		return errors.New("package id is required")
	}
	if strings.TrimSpace(ref.Source) == "" {
		return errors.New("package source is required")
	}
	if strings.TrimSpace(ref.Name) == "" {
		return errors.New("package name is required")
	}

	return nil
}

func ValidateLocalArtifactRef(ref LocalArtifactRef) error {
	if strings.TrimSpace(ref.PayloadHash) != "" && strings.ContainsAny(ref.PayloadHash, " \t\r\n") {
		return errors.New("artifact payload_hash must not contain whitespace")
	}
	if ref.PayloadSize < 0 {
		return errors.New("artifact payload_size must not be negative")
	}
	if strings.TrimSpace(ref.RuntimeName) != "" && !isSafeRuntimeRelative(ref.RuntimeName) {
		return fmt.Errorf("artifact runtime_name %q must be runtime-relative", ref.RuntimeName)
	}

	return nil
}

func ValidatePlanAction(action PlanAction) error {
	if strings.TrimSpace(action.ActionType) == "" {
		return errors.New("plan action_type is required")
	}
	if strings.TrimSpace(action.Target) != "" && !isSafeRuntimeRelative(action.Target) {
		return fmt.Errorf("plan target %q must be runtime-relative", action.Target)
	}
	if action.Size < 0 {
		return errors.New("plan size must not be negative")
	}

	return nil
}

func (spec EnvironmentSpec) Validate() error {
	return ValidateEnvironmentSpec(spec)
}

func (ref PackageRef) Validate() error {
	return ValidatePackageRef(ref)
}

func (ref LocalArtifactRef) Validate() error {
	return ValidateLocalArtifactRef(ref)
}

func (action PlanAction) Validate() error {
	return ValidatePlanAction(action)
}

func isSafeRuntimeRelative(name string) bool {
	if filepath.IsAbs(name) {
		return false
	}
	clean := filepath.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return false
	}
	for _, part := range strings.FieldsFunc(clean, func(r rune) bool { return r == '/' || r == '\\' }) {
		if part == ".." {
			return false
		}
	}

	return true
}
