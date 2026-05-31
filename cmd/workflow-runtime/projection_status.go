package main

import (
	"strings"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

func applyRunStatus(run *workflowv1.WorkflowRunProjection, status workflowv1.WorkflowRunStatus, occurred int64, errorMessage string) {
	if status == workflowv1.WorkflowRunStatus_WORKFLOW_RUN_STATUS_UNSPECIFIED {
		status = workflowv1.WorkflowRunStatus_WORKFLOW_RUN_RUNNING
	}
	run.Status = status
	if run.StartedAtUnix <= 0 {
		run.StartedAtUnix = occurred
	}
	if errorMessage != "" {
		run.ErrorMessage = errorMessage
	}
	if isTerminalRunStatus(status) {
		run.CompletedAtUnix = occurred
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
	if node.StartedAtUnix <= 0 && status != "pending" {
		node.StartedAtUnix = occurred
	}
	if isTerminalRunStatus(req.GetStatus()) && node.StartedAtUnix > 0 {
		node.DurationMs = (occurred - node.StartedAtUnix) * 1000
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

func graphStatus(status workflowv1.WorkflowRunStatus) string {
	switch status {
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_PENDING:
		return "pending"
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_RUNNING:
		return "running"
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_WAITING:
		return "waiting"
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SUCCEEDED:
		return "success"
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_FAILED:
		return "error"
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_CANCELED:
		return "canceled"
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_SKIPPED:
		return "skipped"
	default:
		return "running"
	}
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
