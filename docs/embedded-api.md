# Embedded API

Lucy exposes `github.com/mclucy/lucy/pkg/api` as a public integration boundary for Go consumers that want to embed Lucy directly instead of driving the CLI.

The API is DTO-based and designed for in-memory exchange. Consumers such as Stratum can map their own environment, package, artifact, and metadata models into these request and response types without writing `lucy.yaml`, writing `lucy-lock.yaml`, shelling out to `lucy`, or exchanging manifests through stdin/stdout.

## Package

Use `pkg/api` for embedded integration:

```go
resolver := api.NewNoopResolver()
plan, err := resolver.PlanEnvironment(ctx, api.PlanRequest{
    Environment: api.EnvironmentSpec{
        ID:               "survival",
        MinecraftVersion: "1.21.8",
    },
})
```

The package defines stable request and response DTOs for:

- `Capabilities`
- `EnvironmentSpec`
- `PackageRef`
- `LocalArtifactRef`
- `PlanRequest` and `PlanResult`
- `LockRequest` and `LockResult`
- `StatusRequest` and `StatusResult`

It also defines the public `Resolver` interface:

```go
type Resolver interface {
    Capabilities(ctx context.Context) (Capabilities, error)
    PlanEnvironment(ctx context.Context, req PlanRequest) (PlanResult, error)
    LockEnvironment(ctx context.Context, req LockRequest) (LockResult, error)
    CheckStatus(ctx context.Context, req StatusRequest) (StatusResult, error)
}
```

## Noop Resolver

`NewNoopResolver` returns a deterministic resolver for early integrations and tests. It reports embedded capabilities, returns an empty plan, returns an empty lock, and reports an ok status.

The no-op resolver does not download packages, install Minecraft, invoke providers, run the CLI, or write files. It validates only the lightweight DTO contract.

## Validation

The API includes lightweight validation helpers:

- `EnvironmentSpec` requires `id` and `minecraft_version`.
- `PackageRef` requires `id`, `source`, and `name`.
- `LocalArtifactRef` rejects negative payload sizes and unsafe runtime-relative names.
- `PlanAction` requires `action_type` and rejects unsafe runtime-relative targets.

Validation does not check real Java, Minecraft, loader, Carpet, or MCDR installations.

## Boundaries

This package intentionally does not expose Lucy provider structs or internal package implementation details in public signatures. Consumers should treat the DTOs and `Resolver` interface as the integration contract.

Real provider-backed planning and locking are future work. That future resolver can implement the same `Resolver` interface without forcing embedded consumers to import internal provider packages or use disk-based manifest and lock exchange.

## Non-Goals

- Do not implement real dependency downloads through this API yet.
- Do not force `lucy.yaml` or `lucy-lock.yaml`.
- Do not require CLI execution.
- Do not add a Stratum dependency.
- Do not expose internal provider structs.
- Do not rewrite the provider system.
- Do not install Minecraft, Fabric, Carpet, or MCDR.
