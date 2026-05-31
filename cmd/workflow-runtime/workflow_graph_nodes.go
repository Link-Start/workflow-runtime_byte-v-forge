package main

import (
	"sort"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

func workflowTagNames(tags []n8nWorkflowTag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag.Name != "" {
			names = append(names, tag.Name)
		}
	}
	return names
}

func workflowGraphNodes(nodes []n8nNode, status string, runData n8nRunData, result n8nResultData) []*workflowv1.WorkflowGraphNode {
	out := make([]*workflowv1.WorkflowGraphNode, 0, len(nodes))
	for _, node := range nodes {
		if shortNodeType(node.Type) == "stickyNote" {
			continue
		}
		id := nodeID(node)
		if id == "" {
			continue
		}
		run := nodeRunProjection(node.Name, runData, result, status)
		graphNode := &workflowv1.WorkflowGraphNode{
			Id:            id,
			Name:          node.Name,
			Kind:          shortNodeType(node.Type),
			Status:        run.Status,
			TypeVersion:   valueString(node.TypeVersion),
			Disabled:      node.Disabled,
			StartedAtUnix: run.StartedAtUnix,
			DurationMs:    run.DurationMs,
			ErrorMessage:  run.ErrorMessage,
			Iterations:    run.Iterations,
		}
		if graphNode.Status == "" && node.Disabled {
			graphNode.Status = "skipped"
		}
		if graphNode.Status == "" {
			graphNode.Status = unexecutedNodeStatus(status, len(runData) > 0)
		}
		if len(node.Position) >= 2 {
			graphNode.X = node.Position[0]
			graphNode.Y = node.Position[1]
		}
		out = append(out, graphNode)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}
