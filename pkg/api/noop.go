package api

import "context"

const noopGeneratedAt = "1970-01-01T00:00:00Z"

type NoopResolver struct{}

func NewNoopResolver() NoopResolver {
	return NoopResolver{}
}

func (NoopResolver) Capabilities(ctx context.Context) (Capabilities, error) {
	if err := ctx.Err(); err != nil {
		return Capabilities{}, err
	}

	return Capabilities{
		SupportedSources: []string{"embedded", "local"},
		SupportedLoaders: []string{"vanilla", "fabric", "forge", "neoforge", "paper"},
		SupportsPlan:     true,
		SupportsLock:     true,
		SupportsStatus:   true,
		Metadata: Metadata{
			"resolver": "noop",
		},
	}, nil
}

func (NoopResolver) PlanEnvironment(ctx context.Context, req PlanRequest) (PlanResult, error) {
	if err := ctx.Err(); err != nil {
		return PlanResult{}, err
	}
	if err := ValidateEnvironmentSpec(req.Environment); err != nil {
		return PlanResult{}, err
	}

	return PlanResult{
		Actions:  []PlanAction{},
		Warnings: []string{},
		Errors:   []string{},
	}, nil
}

func (NoopResolver) LockEnvironment(ctx context.Context, req LockRequest) (LockResult, error) {
	if err := ctx.Err(); err != nil {
		return LockResult{}, err
	}
	if err := ValidateEnvironmentSpec(req.Environment); err != nil {
		return LockResult{}, err
	}
	for _, action := range req.Plan.Actions {
		if err := ValidatePlanAction(action); err != nil {
			return LockResult{}, err
		}
	}

	return LockResult{
		LockID:      "noop",
		LockHash:    "noop",
		GeneratedAt: noopGeneratedAt,
		Packages:    []PackageRef{},
		Artifacts:   []LocalArtifactRef{},
		ProviderMetadata: Metadata{
			"resolver": "noop",
		},
	}, nil
}

func (NoopResolver) CheckStatus(ctx context.Context, req StatusRequest) (StatusResult, error) {
	if err := ctx.Err(); err != nil {
		return StatusResult{}, err
	}
	if err := ValidateEnvironmentSpec(req.Environment); err != nil {
		return StatusResult{}, err
	}

	return StatusResult{
		OK:       true,
		Missing:  []string{},
		Drifted:  []string{},
		Warnings: []string{},
		Errors:   []string{},
	}, nil
}
