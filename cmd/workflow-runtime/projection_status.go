package main

import (
	"strings"
	"time"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func applyRunStatus(run *workflowv1.WorkflowRunProjection, status workflowv1.WorkflowRunStatus, occurred int64, errorMessage string) {
	if status == workflowv1.WorkflowRunStatus_WORKFLOW_RUN_STATUS_UNSPECIFIED {
		status = workflowv1.WorkflowRunStatus_WORKFLOW_RUN_RUNNING
	}
	run.Status = status
	if run.GetStartedAt() == nil {
		run.StartedAt = timestampFromUnix(occurred)
	}
	if errorMessage != "" {
		run.ErrorMessage = errorMessage
	}
	if isTerminalRunStatus(status) {
		run.CompletedAt = timestampFromUnix(occurred)
	}
}

func applyNodeStatus(run *workflowv1.WorkflowRunProjection, req *workflowv1.WorkflowStepUpdateRequest, nodeID string, occurred int64) {
	if nodeID == "" {
		return
	}
	status := graphStatus(req.GetStatus())
	node := findGraphNode(run.GraphNodes, nodeID)
	if node == nil {
		node = &workflowv1.WorkflowGraphNode{Id: nodeID, Name: req.GetNodeName()}
		run.GraphNodes = append(run.GraphNodes, node)
	}
	if req.GetNodeName() != "" {
		node.Name = req.GetNodeName()
	}
	node.Status = status
	if node.GetStartedAt() == nil && status != workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_PENDING {
		node.StartedAt = timestampFromUnix(occurred)
	}
	if isTerminalRunStatus(req.GetStatus()) && timestampUnix(node.GetStartedAt()) > 0 {
		node.DurationMs = (occurred - timestampUnix(node.GetStartedAt())) * 1000
	}
	if req.GetErrorMessage() != "" {
		node.ErrorMessage = req.GetErrorMessage()
	}
}

func findGraphNode(nodes []*workflowv1.WorkflowGraphNode, id string) *workflowv1.WorkflowGraphNode {
	for _, node := range nodes {
		if node.GetId() == id || node.GetName() == id {
			return node
		}
	}
	return nil
}

func graphStatus(status workflowv1.WorkflowRunStatus) workflowv1.WorkflowGraphElementStatus {
	switch status {
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_PENDING:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_PENDING
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_RUNNING:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_RUNNING
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_WAITING:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_RUNNING
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SUCCEEDED:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_SUCCEEDED
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_FAILED:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_FAILED
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_CANCELED:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_FAILED
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SKIPPED:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_SKIPPED
	default:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_RUNNING
	}
}

func graphStatusFromString(value string) workflowv1.WorkflowGraphElementStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending", "created", "new", "queued":
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_PENDING
	case "running", "waiting", "started", "in_progress":
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_RUNNING
	case "success", "succeeded", "completed", "done":
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_SUCCEEDED
	case "failed", "error", "crashed", "canceled", "cancelled", "aborted":
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_FAILED
	case "skipped", "skip":
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_SKIPPED
	default:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_STATUS_UNSPECIFIED
	}
}

func timestampFromUnix(value int64) *timestamppb.Timestamp {
	if value <= 0 {
		return nil
	}
	return timestamppb.New(time.Unix(value, 0))
}

func isTerminalRunStatus(status workflowv1.WorkflowRunStatus) bool {
	switch status {
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SUCCEEDED,
		workflowv1.WorkflowRunStatus_WORKFLOW_RUN_FAILED,
		workflowv1.WorkflowRunStatus_WORKFLOW_RUN_CANCELED,
		workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SKIPPED:
		return true
	default:
		return false
	}
}

func runStatusFromString(value string) workflowv1.WorkflowRunStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pending", "created", "new", "queued":
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_PENDING
	case "running", "started", "in_progress":
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_RUNNING
	case "waiting", "wait":
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_WAITING
	case "success", "succeeded", "completed", "done":
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SUCCEEDED
	case "failed", "error", "crashed":
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_FAILED
	case "canceled", "cancelled", "aborted":
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_CANCELED
	case "skipped", "skip":
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SKIPPED
	default:
		return workflowv1.WorkflowRunStatus_WORKFLOW_RUN_STATUS_UNSPECIFIED
	}
}
