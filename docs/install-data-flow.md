# Lucy Install Data Flow

## Files inspected

| File | Purpose |
|------|---------|
| `main.go` | Program entrypoint; calls `cmd.Execute()` |
| `cmd/root.go` | Cobra root command; `Execute()` entry |
| `cmd/cmd_install.go` | `lucy install` command; builds plan from manifest/lock |
| `cmd/cmd_add.go` | `lucy add` command; parses CLI args, calls `install.Install`/`install.InstallMany`, updates state |
| `cmd/parser.go` | CLI package completion helpers |
| `install/install.go` | `Install()` single package, `installPlatform()` identity installers |
| `install/install_many.go` | `InstallMany()` batch install; main orchestration loop |
| `install/install_request.go` | `PackageRequest` type, `ParsePackageRequest()` from string |
| `install/install_options.go` | `InstallOptions` struct |
| `install/install_recursive_types.go` | `RecursiveTransaction`, `CandidateNode`, `ReconcileDiff`, `ApplyPlan`, phases |
| `install/install_recursive_resolve.go` | `BuildCandidateGraph()` — recursive dependency expansion |
| `install/install_recursive_resolve_adapter.go` | `providerCandidateResolver` — calls routing.FetchMany/DependenciesMany |
| `install/install_recursive_resolution_plan.go` | `recursiveResolutionPlan` — per-iteration plan/refine |
| `install/install_recursive_download.go` | `downloadBatchPackages()` — concurrent download to staging |
| `install/install_recursive_verify.go` | `VerifyDownloadedArtifacts()` — local JAR analysis |
| `install/install_recursive_reconcile.go` | `ReconcileTransaction()` — diff candidate vs verified |
| `install/install_recursive_constraints.go` | `MergeConstraintGraph()` — pure constraint solver |
| `install/install_recursive_installed.go` | `SnapshotInstalledConstraints()` — snapshot current state |
| `install/install_recursive_apply.go` | `ApplyValidatedClosure()` — move files to target, remove extras |
| `install/install_recursive_errors.go` | `ConstraintConflictError` |
| `install/install_output.go` | Logging/progress helpers |
| `install/install_helpers.go` | `ensureServerPlatformMatch()` |
| `types/type_core.go` | `PackageRef`, `VersionedPackageRef`, `StringablePackageRef` |
| `types/type_package.go` | `Package` (Id, Dependencies, Local, Remote, Supports, Information) |
| `types/type_id.go` | `PlatformId`, identity package logic |
| `types/type_dependency.go` | `Dependency`, `BareVersion`, `VersionExpr`, constraint operators |
| `types/type_source.go` | `SourceId` enum (Auto, CurseForge, Modrinth, ...) |
| `types/type_meta.go` | `Metadata` for info display |
| `types/type_package_identity.go` | `PackageIdentity`, `PackageRequest` (legacy), `ResolvedPackage` |
| `types/type_server_runtime.go` | Server runtime types |
| `types/type_server_topology.go` | RuntimeTopology, Capabilities |
| `syntax/syntax.go` | `Parse()`, `ParsePackageRef()` — string→types parsing |
| `state/manifest.go` | `Manifest`, `ManifestPackage`, roles, Upsert/Update helpers |
| `state/lock.go` | `Lock`, `LockedPackage`, validation |
| `state/service.go` | `ProjectStateService` — load/save lucy.yaml + lucy-lock.yaml |
| `state/io.go` | Read/write manifest/lock files |
| `upstream/upstream.go` | `Provider` interface (Fetch, Dependencies, Support, Id) |
| `upstream/upstream_types.go` | `FetchResult`, raw conversion contracts |
| `upstream/upstream_descriptive.go` | `Searcher`, `Informer` interfaces |
| `upstream/upstream_parsable.go` | `VersionSelectorResolver`, `ArtifactMapper` interfaces |
| `upstream/routing/routing.go` | `ResolveProviders()`, `ResolveProvidersFromTopology()`, provider maps |
| `upstream/routing/routing_topology.go` | `providerSourcesFromTopology()` — capability→source mapping |
| `upstream/routing/routing_policy.go` | Source priority ordering, platform→sources rules |
| `upstream/routing/routing_execution.go` | `FetchMany()`, `DependenciesMany()`, `SearchMany()` — parallel execution |
| `cache/cache.go` | Cache store |
| `util/download.go` | `CachedDownload()` — HTTP download with cache+integrity |
| `probe/` | Server info detection (runtime, mods, topology) |
| `artifact/` | Local JAR analysis (dependencies, versions) |
| `pkg/api/` | Embedded integration contract (Plan/Lock/Status) |

## Current install flow

### Entry points

There are two CLI entry points that trigger installation:

1. **`lucy add <pkg>...`** — `cmd/cmd_add.go:73` `actionAdd()`
   - Parses args via `install.ParsePackageRequest()` → `install.PackageRequest`
   - Calls `install.Install()` for single requests, `install.InstallMany()` for multiple
   - Writes manifest + lock

2. **`lucy install`** — `cmd/cmd_install.go:33` `actionInstall()`
   - No args; reads state from `lucy.yaml` + `lucy-lock.yaml`
   - Loads `state.ProjectStateService`
   - Builds plan via `buildInstallSyncPlan()` — either exact lock replay or manifest-required resolution
   - Calls `install.InstallMany()` with the plan
   - Writes updated lock

### Pipeline (InstallMany)

`install.InstallMany()` at `install/install_many.go:13` drives the full pipeline:

1. **Partition** — Split requests into identity packages (platforms) and regular packages
2. **Identity install** — Install platform loaders via `installPlatform()` (forge, fabric, etc.)
3. **Provider routing** — `routing.ResolveProvidersFromTopology()` maps server topology → `[]upstream.Provider`
4. **Resolution plan** — `newRecursiveResolutionPlan()` seeds roots + installed constraints
5. **Reconcile loop** (up to 3 iterations):
   a. **Candidate expansion** — `BuildCandidateGraph()` expands advisory dependency closure
   b. **Download** — `downloadBatchPackages()` fetches artifacts to staging dir
   c. **Verify** — `VerifyDownloadedArtifacts()` runs local JAR detectors → `VerifiedGraph`
   d. **Reconcile** — `ReconcileTransaction()` diffs advisory vs verified; if not stable, refine plan and loop
6. **Apply plan** — `BuildRecursiveApplyPlan()` → `ApplyValidatedClosure()` moves files to target

### Identity package flow

`install.Install()` at `install/install.go:27` handles non-recursive identity packages:
- If not an identity package, delegates to `InstallMany()`
- Identity packages go through `installPlatform()` → platform-specific installer (vanilla, forge, fabric, neoforge, MCDR)

## Type map

```
User Input (string)
  │
  ▼
syntax.Parse() / syntax.ParsePackageRef()
  │
  ├── types.PackageRef { Platform, Name }          // raw input, no version
  ├── types.VersionedPackageRef { Platform, Name, Version }  // full ref
  │
  ▼
install.ParsePackageRequest()
  │
  ├── install.PackageRequest { Ref, Version, Optional, Source }
  │
  ▼
install.InstallMany()
  │
  ├── types.VersionedPackageRef[]  // after requestsToIds() + prepareBatchIDs()
  │
  ▼
routing.ResolveProvidersFromTopology()
  │
  ├── []upstream.Provider     // runtime executor instances
  │
  ▼
BuildCandidateGraph()
  │
  ├── candidateGraphPlanner
  │   ├── uses providerCandidateResolver
  │   │   ├── routing.FetchMany() → upstream.Fetch() → provider.Fetch()
  │   │   │   └── upstream.FetchResult { ResolvedID, Remote: PackageRemote }
  │   │   └── routing.DependenciesMany() → upstream.Dependencies() → provider.Dependencies()
  │   │       └── types.PackageDependencies
  │   └── MergeConstraintGraph() → ConstraintGraph
  │
  ├── RecursiveTransaction
  │   ├── Roots []VersionedPackageRef
  │   ├── InstalledConstraints []InstalledConstraint
  │   ├── Providers []upstream.Provider
  │   ├── CandidateGraph map[string]CandidateNode
  │   ├── DownloadedArtifacts map[string]string
  │   ├── VerifiedGraph map[string]CandidateNode
  │   ├── ReconcileDiff ReconcileDiff
  │   ├── Apply *ApplyPlan
  │   └── StagingDir string
  │
  ├── CandidateNode { Package: types.Package, ProvenancePath, Advisory }
  │
  ▼
downloadBatchPackages()
  │
  ├── util.CachedDownload(url, stagingDir, opts)
  │   ├── cache.Network().Get() → hit? copy from cache
  │   └── downloadAndCache() → http.Get → hash + cache.IngestEntry
  │       └── integrity verification (sha1/sha512)
  │
  ▼
VerifyDownloadedArtifacts()
  │
  ├── artifact.Analyze(path) → []ArtifactInfo
  │   └── local JAR detector (fabric/forge/neoforge)
  │
  ├── VerifiedGraph map[string]CandidateNode  // authoritative deps
  │
  ▼
ReconcileTransaction()
  │
  ├── ReconcileDiff { Missing, Extra, Tightened }
  │
  ▼ (stable)
BuildRecursiveApplyPlan() → ApplyPlan { Install, Remove }
  │
  ▼
ApplyValidatedClosure()
  │
  ├── os.Rename() staging → target directory
  ├── os.Remove() for extra packages
  │
  ▼
Lock/manifest update
  ├── buildUpdatedManifest() → state.UpsertManifestRequiredIntent()
  ├── buildUpdatedLock() → state.Lock { LockedPackage[] }
  └── state.WriteManifest() + state.WriteLock()

```

### State types

```
state.Manifest
  ├── FormatVersion string
  ├── Environment ManifestEnvironment
  ├── Packages []ManifestPackage { ID, Version, Source, Role, Side, Optional, Pinned }
  └── Bundles []ManifestBundle

state.Lock
  ├── Version, GeneratedAt, ManifestFingerprint
  ├── GameVersion, Platform, PlatformVersion
  ├── Packages []LockedPackage { ID, Version, Source, URL, Filename, Hash, HashAlgorithm, InstallPath, Side, Optional, Embedded, Provenance, Requester }
  └── Bundles []LockedBundle
```

### Upstream interfaces

```
upstream.Provider
  ├── Fetch(id VersionedPackageRef) → RawPackageRemote, error
  ├── Dependencies(id VersionedPackageRef) → RawPackageDependencies, error
  ├── Support(name BarePackageName) → RawProjectSupport, error
  └── Id() → SourceId

upstream.Searcher
  └── Search(q Query) → SearchResponse, error

upstream.Informer
  └── Info(ref PackageRef) → Metadata, error

upstream.VersionSelectorResolver
  └── ResolveVersionSelector(ref VersionedPackageRef) → VersionedPackageRef, error
```

## Data flow diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  CLI entry                                                      │
│                                                                  │
│  lucy add "fabric/ae2@19.0.0"  ──┐                              │
│  lucy install (no args)         ──┤                              │
│                                  │                              │
│  cmd/cmd_add.go:73 actionAdd()   │                              │
│  cmd/cmd_install.go:33 actionInstall()                           │
└────────────────────────┬────────┘                              │
                         │                                         │
                         ▼                                         │
┌────────────────────────────────────────────────────┐            │
│  Input parsing                                      │            │
│                                                     │            │
│  install.ParsePackageRequest() → PackageRequest     │◄───────────┘
│  syntax.Parse() → VersionedPackageRef               │
└─────────────────────┬──────────────────────────────┘
                      │
                      ▼
┌────────────────────────────────────────────────────┐
│  Partition                                          │
│                                                     │
│  requestsToIds(requests) → VersionedPackageRef[]    │
│  prepareBatchIDs() → deduplicate + normalize        │
│  partitionBatchIDs() → identityIds + regularIds     │
└─────────┬───────────────────────────┬──────────────┘
          │                           │
          ▼                           ▼
┌──────────────────┐    ┌────────────────────────────────────────┐
│  Identity install │    │  Regular (plugin/mod) install           │
│                   │    │                                          │
│  installPlatform()│    │  routing.ResolveProvidersFromTopology()  │
│  (forge/fabric/   │    │  → []upstream.Provider                   │
│   vanilla/neoforge)   │                                          │
└──────────────────┘    └────────────────┬─────────────────────────┘
                                         │
                                         ▼
                              ┌────────────────────────────────────┐
                              │  Resolution plan                    │
                              │                                     │
                              │  newRecursiveResolutionPlan()       │
                              │  SnapshotInstalledConstraints()     │
                              │                                     │
                              │  Reconcile loop (max 3 iterations): │
                              │                                     │
                              │  ┌─────────────────────────────┐    │
                              │  │  BuildCandidateGraph()      │    │
                              │  │  → providerCandidateResolver│    │
                              │  │    → routing.FetchMany()    │    │
                              │  │    → routing.DependenciesMany│   │
                              │  │    → MergeConstraintGraph() │    │
                              │  └──────────┬──────────────────┘    │
                              │             ▼                       │
                              │  ┌─────────────────────────────┐    │
                              │  │  downloadBatchPackages()     │    │
                              │  │  → CachedDownload()         │    │
                              │  │    → cache check + HTTP     │    │
                              │  │    → hash verification      │    │
                              │  └──────────┬──────────────────┘    │
                              │             ▼                       │
                              │  ┌─────────────────────────────┐    │
                              │  │  VerifyDownloadedArtifacts() │    │
                              │  │  → artifact.Analyze()        │    │
                              │  │  → VerifiedGraph             │    │
                              │  └──────────┬──────────────────┘    │
                              │             ▼                       │
                              │  ┌─────────────────────────────┐    │
                              │  │  ReconcileTransaction()      │    │
                              │  │  → diff candidate vs verify │    │
                              │  │  → stable? break            │    │
                              │  │  → refine plan, loop        │    │
                              │  └──────────┬──────────────────┘    │
                              │             ▼ (stable)              │
                              └────────────────┬───────────────────┘
                                               │
                                               ▼
                              ┌────────────────────────────────────┐
                              │  Apply                              │
                              │                                     │
                              │  BuildRecursiveApplyPlan()          │
                              │  → ApplyPlan { Install, Remove }    │
                              │                                     │
                              │  ApplyValidatedClosure()            │
                              │  → os.Rename() staging → mods/      │
                              │  → os.Remove() extra packages      │
                              └────────────────┬───────────────────┘
                                               │
                                               ▼
                              ┌────────────────────────────────────┐
                              │  Post-install state update          │
                              │                                     │
                              │  buildUpdatedManifest()             │
                              │  buildUpdatedLock()                 │
                              │  state.WriteManifest()             │
                              │  state.WriteLock()                 │
                              └────────────────────────────────────┘
```

## Problem map

### 1. Input vs normalized package ref

`PackageRef` (raw input: `Platform`, `Name`) and `VersionedPackageRef` (adds `Version`) are both used interchangeably across the boundary. The `PackageRequest` at `install/install_request.go:19` wraps a `PackageRef` and a separate `Version`, while `types.VersionedPackageRef` bundles them. There is a TODO comment chain about removing this wrapping.

**File:** `install/install_request.go:19`, `install/install_many.go:191` `requestsToIds()`

### 2. Package request vs resolved package

`install.PackageRequest` holds user intent (ref + version + source + optional flag). `types.Package` is the resolved envelope (Id + Remote + Local + Dependencies). There is no intermediate "resolved package request" type — the gap is bridged by `providerCandidateResolver.ResolvePackage()` which returns a `types.Package` directly from a `VersionedPackageRef`.

Additionally, `types.PackageIdentity` and `types.PackageRequest` in `type_package_identity.go` define an entirely separate model that is **not used** by the install pipeline. These are dead or speculative types.

**File:** `install/install_request.go:19`, `types/type_package.go:10`, `types/type_package_identity.go:10`

### 3. Provider selection vs resolution

Provider selection happens in `routing.ResolveProvidersFromTopology()` based on server topology capabilities. Resolution (Fetch + Dependencies) happens inside `providerCandidateResolver` which calls `routing.FetchMany()` and `routing.DependenciesMany()`. The two are clearly separated but the `providerCandidateResolver` at `install/install_recursive_resolve_adapter.go:11` also implements the "try multiple version fallbacks" logic that could be considered resolution policy.

**File:** `upstream/routing/routing.go:285`, `install/install_recursive_resolve_adapter.go:15`

### 4. Resolution vs download

Resolution (via provider Fetch) returns metadata (`PackageRemote` with URL, hash, filename). Download happens in `downloadBatchPackages()` which takes the URL and hash from `Package.Remote` and fetches the artifact. These are conceptually separate but `BuildCandidateGraph()` must complete before download starts, and they are mediated by the `RecursiveTransaction` phases.

**File:** `install/install_recursive_resolve.go:32`, `install/install_recursive_download.go:78`

### 5. Download vs install

Download writes to a temp staging directory. Install (`ApplyValidatedClosure()`) moves files from staging to the target mods/plugins directory. These are cleanly separated by the `PhaseDownloaded → PhaseVerified → PhaseCommitted` phase machine.

**File:** `install/install_recursive_types.go:13` (phases), `install/install_recursive_apply.go:94`

### 6. Provider metadata vs package identity

`PackageRemote.Source` is a `SourceId` provenance marker, not a provider selection key. The `PackageIdentity` type in `types/type_package_identity.go` includes `Source`, `Name`, `Loader`, `MinecraftVersion`, and `Version` — a superset of what `VersionedPackageRef` carries. However, `PackageIdentity` is not used in the actual install pipeline.

**File:** `types/type_package_identity.go:10`, `types/type_package.go:38`

### 7. Filesystem layout vs package semantics

Install destination is determined at `install/install_recursive_apply.go:14` `recursiveInstallDestination()` by checking platform type (modding → `ModPath[0]`, MCDR → plugin directory, else root). This mixes filesystem knowledge with package semantics. The mod path detection comes from `probe.ServerInfo()` which reads the actual server directory.

**File:** `install/install_recursive_apply.go:14`

## Refactor seams

### S1: Remove dead `PackageIdentity` / `PackageRequest` in types

**Files:** `types/type_package_identity.go`

**Why:** `types.PackageIdentity`, `types.PackageRequest`, and `types.ResolvedPackage` are unused by the actual install pipeline. They exist as a speculative model that predates the current `install.PackageRequest` + `types.Package` approach. Removing them eliminates confusion and dead code.

**Behavior unchanged:** CLI output, install behavior, lock/manifest format.

**Tests:** Remove the dead types; no test changes needed since nothing references them.

### S2: Inline `requestsToIds` conversion

**Files:** `install/install_many.go:191`, call sites in `cmd/cmd_add.go`, `cmd/cmd_install.go`

**Why:** The `PackageRequest → VersionedPackageRef` conversion is a thin wrapper that loses the `Source` field. After the conversion, `batchSource` is recovered from `requests[0].Source` as a workaround. Making the pipeline accept `PackageRequest` directly would simplify the boundary.

**Behavior unchanged:** Same conversion, just moved closer to the boundary.

**Tests:** Unit test for `PrepareBatchIDs` can be updated to accept `PackageRequest`.

### S3: Extract a `ResolveBatch` step from `BuildCandidateGraph`

**Files:** `install/install_recursive_resolve.go`, `install/install_recursive_resolve_adapter.go`

**Why:** `BuildCandidateGraph()` does both provider Fetch + dependency expansion + constraint merging. The provider fallback logic (compatible → latest → any version attempts) is inside `providerCandidateResolver.ResolvePackage()`. Extracting a pure "fetch metadata" phase before expansion would clarify the boundary between resolution and dependency walking.

**Behavior unchanged:** Same provider calls, same fallback order.

**Tests:** Add unit tests for the extracted fetch step with mock providers.

### S4: Normalize `PackageRemote.Source` to always match fetch source

**Files:** `install/install_recursive_resolve_adapter.go:60` `providersForSource()`

**Why:** `providersForSource()` already filters by `PackageRemote.Source`, but the `Source` field is populated by the provider's `ToPackageRemote()` conversion rather than being enforced by routing. Normalizing this at the routing boundary would make provenance more reliable.

**Behavior unchanged:** Same source label on fetched packages.

**Tests:** Add a test that verifies `PackageRemote.Source` equals the provider's `Id()`.

### S5: Extract install destination into a dedicated type

**Files:** `install/install_recursive_apply.go:14`, `probe/` workspace types

**Why:** `recursiveInstallDestination()` embeds platform-specific directory logic inside the apply phase. Moving this to a `InstallTarget` type (computed once from `probe.Workspace`) would make the apply phase purely mechanical.

**Behavior unchanged:** Same destination directories for same platforms.

**Tests:** Unit test `InstallTarget` resolution for each platform.

## pkg/api recommendation

**Recommendation: Delete `pkg/api`.**

The package at `pkg/api/` defines a `Resolver` interface with `PlanEnvironment`, `LockEnvironment`, and `CheckStatus` — a speculative embedded integration contract that is:

1. **Not used** — Zero callers in the codebase. The `NoopResolver` is the only implementation.
2. **Out of sync** — The types (`PackageRef`, `MinecraftVersion`, `LoaderType`) follow a different model than the actual install pipeline (`install.PackageRequest`, `types.VersionedPackageRef`, `types.PlatformId`). There is no adapter or conversion between the two.
3. **Premature** — The contract prescribes a plan/lock/status workflow that does not match how the current pipeline works (recursive reconcile loop, advisory vs verified graphs, etc.).
4. **Misinforms consumers** — An external consumer implementing this contract would build against an abstraction that has never been tested against real Lucy behavior.

The commit history shows `5722471 api: add embedded integration contract` was added before the recursive install pipeline was fully developed. The actual pipeline has evolved significantly since. Keeping this package as "experimental" creates maintenance burden without value. Delete it now; reintroduce a well-grounded API only after the install pipeline stabilizes and external consumers exist.
