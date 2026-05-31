package main

import (
	"fmt"
	"sort"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

func workflowGraphEdges(nodes []n8nNode, connections n8nConnections, runData n8nRunData) []*workflowv1.WorkflowGraphEdge {
	nameToID := make(map[string]string, len(nodes))
	for _, node := range nodes {
		if shortNodeType(node.Type) == "stickyNote" {
			continue
		}
		nameToID[node.Name] = nodeID(node)
	}
	var out []*workflowv1.WorkflowGraphEdge
	for sourceName, groups := range connections {
		sourceID := nameToID[sourceName]
		if sourceID == "" {
			continue
		}
		for _, ports := range groups {
			for portIndex, port := range ports {
				for edgeIndex, conn := range port {
					targetID := nameToID[conn.Node]
					if targetID == "" {
						continue
					}
					out = append(out, &workflowv1.WorkflowGraphEdge{
						Id:     fmt.Sprintf("%s:%s:%d:%d", sourceID, targetID, portIndex, edgeIndex),
						Source: sourceID,
						Target: targetID,
						Status: edgeExecutionStatus(sourceName, conn.Node, runData),
					})
				}
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}
