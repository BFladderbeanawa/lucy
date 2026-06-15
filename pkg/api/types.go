package api

import "context"

type Metadata map[string]string

type Capabilities struct {
	SupportedSources []string `json:"supported_sources"`
	SupportedLoaders []string `json:"supported_loaders"`
	SupportsPlan     bool     `json:"supports_plan"`
	SupportsLock     bool     `json:"supports_lock"`
	SupportsStatus   bool     `json:"supports_status"`
	Metadata         Metadata `json:"metadata,omitempty"`
}

type EnvironmentSpec struct {
	ID               string             `json:"id"`
	MinecraftVersion string             `json:"minecraft_version"`
	JavaVersion      string             `json:"java_version,omitempty"`
	LoaderType       string             `json:"loader_type,omitempty"`
	LoaderVersion    string             `json:"loader_version,omitempty"`
	ServerCore       string             `json:"server_core,omitempty"`
	CarpetRequired   bool               `json:"carpet_required,omitempty"`
	MCDRRequired     bool               `json:"mcdr_required,omitempty"`
	RuntimeProfileID string             `json:"runtime_profile_id,omitempty"`
	Packages         []PackageRef       `json:"packages,omitempty"`
	LocalArtifacts   []LocalArtifactRef `json:"local_artifacts,omitempty"`
	Metadata         Metadata           `json:"metadata,omitempty"`
}

type PackageRef struct {
	ID                string   `json:"id"`
	Source            string   `json:"source"`
	Name              string   `json:"name"`
	VersionConstraint string   `json:"version_constraint,omitempty"`
	MinecraftVersion  string   `json:"minecraft_version,omitempty"`
	Loader            string   `json:"loader,omitempty"`
	Required          bool     `json:"required,omitempty"`
	Metadata          Metadata `json:"metadata,omitempty"`
}

type LocalArtifactRef struct {
	ArtifactID       string   `json:"artifact_id"`
	PayloadAlgorithm string   `json:"payload_algorithm,omitempty"`
	PayloadHash      string   `json:"payload_hash,omitempty"`
	PayloadSize      int64    `json:"payload_size,omitempty"`
	ArtifactType     string   `json:"artifact_type,omitempty"`
	RuntimeName      string   `json:"runtime_name,omitempty"`
	Metadata         Metadata `json:"metadata,omitempty"`
}

type PlanRequest struct {
	Environment EnvironmentSpec `json:"environment"`
	Metadata    Metadata        `json:"metadata,omitempty"`
}

type PlanResult struct {
	Actions            []PlanAction `json:"actions"`
	Warnings           []string     `json:"warnings"`
	Errors             []string     `json:"errors"`
	RequiresLockUpdate bool         `json:"requires_lock_update"`
	Metadata           Metadata     `json:"metadata,omitempty"`
}

type PlanAction struct {
	ActionType string   `json:"action_type"`
	PackageID  string   `json:"package_id,omitempty"`
	ArtifactID string   `json:"artifact_id,omitempty"`
	Source     string   `json:"source,omitempty"`
	Target     string   `json:"target,omitempty"`
	Hash       string   `json:"hash,omitempty"`
	Size       int64    `json:"size,omitempty"`
	Metadata   Metadata `json:"metadata,omitempty"`
}

type LockRequest struct {
	Environment EnvironmentSpec `json:"environment"`
	Plan        PlanResult      `json:"plan"`
	Metadata    Metadata        `json:"metadata,omitempty"`
}

type LockResult struct {
	LockID           string             `json:"lock_id"`
	LockHash         string             `json:"lock_hash"`
	GeneratedAt      string             `json:"generated_at"`
	Packages         []PackageRef       `json:"packages"`
	Artifacts        []LocalArtifactRef `json:"artifacts"`
	ProviderMetadata Metadata           `json:"provider_metadata,omitempty"`
}

type StatusRequest struct {
	Environment EnvironmentSpec `json:"environment"`
	Lock        LockResult      `json:"lock"`
	Metadata    Metadata        `json:"metadata,omitempty"`
}

type StatusResult struct {
	OK       bool     `json:"ok"`
	Missing  []string `json:"missing"`
	Drifted  []string `json:"drifted"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type Resolver interface {
	Capabilities(ctx context.Context) (Capabilities, error)
	PlanEnvironment(ctx context.Context, req PlanRequest) (PlanResult, error)
	LockEnvironment(ctx context.Context, req LockRequest) (LockResult, error)
	CheckStatus(ctx context.Context, req StatusRequest) (StatusResult, error)
}
