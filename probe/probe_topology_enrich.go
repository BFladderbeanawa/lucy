package probe

import (
	"strings"

	"github.com/mclucy/lucy/types"
)

func EnrichTopologyFromPackages(exec *types.ExecutableInfo, packages []types.Package) {
	if exec == nil || exec.Topology == nil || !exec.Topology.Resolved() {
		return
	}

	evidence := detectedRuntimeEvidence(packages)
	for _, nodeID := range evidence {
		entry, ok := FindEntry(nodeID)
		if !ok {
			continue
		}

		annotation := BuildTopologyFromEntry(entry)
		if annotation == nil {
			continue
		}

		mergeTopology(exec.Topology, annotation)
	}
}

func detectedRuntimeEvidence(packages []types.Package) []types.RuntimeNodeID {
	names := make(map[string]struct{}, len(packages))
	for _, pkg := range packages {
		normalized := strings.ToLower(strings.TrimSpace(pkg.Id.Name.String()))
		if normalized == "" {
			continue
		}
		names[normalized] = struct{}{}
	}

	detected := make([]types.RuntimeNodeID, 0, 6)
	if hasAnyName(names, "connector", "sinytra-connector", "fabric-connector") {
		detected = append(detected, RuntimeNodeConnector)
	}
	if hasAnyName(names, "kilt") {
		detected = append(detected, RuntimeNodeKilt)
	}
	if hasAnyName(names, "velocity", "velocity-proxy") {
		detected = append(detected, RuntimeNodeVelocity)
	}
	if hasAnyName(names, "bungeecord", "bungee") {
		detected = append(detected, RuntimeNodeBungeecord)
	}
	if hasAnyName(names, "waterfall") {
		detected = append(detected, RuntimeNodeWaterfall)
	}
	if hasAnyName(names, "geyser", "geyser-spigot", "geyser-fabric") {
		detected = append(detected, RuntimeNodeGeyser)
	}

	return detected
}

func hasAnyName(names map[string]struct{}, candidates ...string) bool {
	for _, candidate := range candidates {
		if _, ok := names[candidate]; ok {
			return true
		}
	}

	return false
}

func mergeTopology(dst *types.RuntimeTopology, src *types.RuntimeTopology) {
	if dst == nil || src == nil {
		return
	}

	seenNodes := make(map[types.RuntimeNodeID]struct{}, len(dst.Nodes)+len(src.Nodes))
	for _, node := range dst.Nodes {
		seenNodes[node.ID] = struct{}{}
	}

	for _, node := range src.Nodes {
		if _, exists := seenNodes[node.ID]; exists {
			continue
		}
		dst.Nodes = append(dst.Nodes, node)
		seenNodes[node.ID] = struct{}{}
	}

	seenEdges := make(map[types.RuntimeEdge]struct{}, len(dst.Edges)+len(src.Edges))
	for _, edge := range dst.Edges {
		seenEdges[edge] = struct{}{}
	}

	for _, edge := range src.Edges {
		if _, exists := seenEdges[edge]; exists {
			continue
		}
		dst.Edges = append(dst.Edges, edge)
		seenEdges[edge] = struct{}{}
	}
}
