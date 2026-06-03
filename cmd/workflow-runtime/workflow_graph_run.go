package main

import (
	"strings"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

type nodeRunData struct {
	Status        workflowv1.WorkflowGraphElementStatus
	StartedAtUnix int64
	DurationMs    int64
	ErrorMessage  string
	Iterations    int32
}

func nodeRunProjection(name string, runData n8nRunData, result n8nResultData, executionStatus string) nodeRunData {
	tasks := runData[name]
	if len(tasks) == 0 {
		if result.LastNodeExecuted == name && (executionStatus == "running" || executionStatus == "waiting") {
			return nodeRunData{Status: graphStatusFromString(executionStatus)}
		}
		return nodeRunData{}
	}
	task := tasks[len(tasks)-1]
	status := task.ExecutionStatus
	if status == "" {
		status = "success"
	}
	err := errorMessage(task.Error)
	if err == "" && result.LastNodeExecuted == name {
		err = errorMessage(result.Error)
	}
	if err != "" {
		status = "error"
	}
	return nodeRunData{
		Status:        graphStatusFromString(status),
		StartedAtUnix: millisToUnix(task.StartTime),
		DurationMs:    task.ExecutionTime,
		ErrorMessage:  err,
		Iterations:    int32(len(tasks)),
	}
}

func unexecutedNodeStatus(executionStatus string, hasRunData bool) workflowv1.WorkflowGraphElementStatus {
	if !hasRunData {
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_STATUS_UNSPECIFIED
	}
	switch strings.ToLower(strings.TrimSpace(executionStatus)) {
	case "success", "error", "failed", "crashed", "canceled", "cancelled", "aborted":
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_SKIPPED
	case "running", "waiting", "new", "pending", "created", "queued":
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_PENDING
	default:
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_STATUS_UNSPECIFIED
	}
}

func edgeExecutionStatus(sourceName string, targetName string, runData n8nRunData) workflowv1.WorkflowGraphElementStatus {
	if len(runData) == 0 {
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_STATUS_UNSPECIFIED
	}
	targetTasks := runData[targetName]
	if len(targetTasks) == 0 {
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_STATUS_UNSPECIFIED
	}
	if errorMessage(targetTasks[len(targetTasks)-1].Error) != "" {
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_FAILED
	}
	if len(runData[sourceName]) > 0 {
		return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_SUCCEEDED
	}
	return workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_STATUS_UNSPECIFIED
}
