package main

import "strings"

type nodeRunData struct {
	Status        string
	StartedAtUnix int64
	DurationMs    int64
	ErrorMessage  string
	Iterations    int32
}

func nodeRunProjection(name string, runData n8nRunData, result n8nResultData, executionStatus string) nodeRunData {
	tasks := runData[name]
	if len(tasks) == 0 {
		if result.LastNodeExecuted == name && (executionStatus == "running" || executionStatus == "waiting") {
			return nodeRunData{Status: executionStatus}
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
		Status:        status,
		StartedAtUnix: millisToUnix(task.StartTime),
		DurationMs:    task.ExecutionTime,
		ErrorMessage:  err,
		Iterations:    int32(len(tasks)),
	}
}

func unexecutedNodeStatus(executionStatus string, hasRunData bool) string {
	if !hasRunData {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(executionStatus)) {
	case "success", "error", "failed", "crashed", "canceled", "cancelled", "aborted":
		return "skipped"
	case "running", "waiting", "new", "pending", "created", "queued":
		return "pending"
	default:
		return ""
	}
}

func edgeExecutionStatus(sourceName string, targetName string, runData n8nRunData) string {
	if len(runData) == 0 {
		return ""
	}
	targetTasks := runData[targetName]
	if len(targetTasks) == 0 {
		return ""
	}
	if errorMessage(targetTasks[len(targetTasks)-1].Error) != "" {
		return "error"
	}
	if len(runData[sourceName]) > 0 {
		return "success"
	}
	return ""
}
