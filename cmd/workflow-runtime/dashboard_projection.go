package main

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
	"github.com/byte-v-forge/common-lib/protojsonhttp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *dashboardServer) handleWorkflowStepUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorizeWorkflowStepUpdate(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req workflowv1.WorkflowStepUpdateRequest
	if err := protojsonhttp.ReadRequest(r, &req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := validateWorkflowStepUpdate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := s.projections.apply(r.Context(), &req)
	if err != nil {
		http.Error(w, "workflow projection store unavailable", http.StatusServiceUnavailable)
		return
	}
	writeProtoJSON(w, http.StatusOK, &workflowv1.WorkflowStepUpdateResponse{Run: result.run, Duplicate: result.duplicate})
}

func (s *dashboardServer) workflowSummary(r *http.Request) *workflowv1.WorkflowRuntimeSummary {
	return s.workflowSummaryForContext(r.Context(), workflowSummaryPageRequestFromHTTP(r))
}

func (s *dashboardServer) workflowSummaryForContext(ctx context.Context, pages workflowSummaryPageRequest) *workflowv1.WorkflowRuntimeSummary {
	summary := s.workflowRuntime.summary(ctx, pages)
	projected, err := s.projections.list(ctx)
	if err != nil {
		summary.ApiStatus = workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_DEGRADED
		summary.ApiMessage = "workflow projection store unavailable"
	}
	runs := mergeLiveExecutionRuns(projected, summary.GetExecutions())
	pagedRuns, nextRunToken := paginateWorkflowRuns(runs, pages.runs)
	summary.Runs = pagedRuns
	summary.RunsPageInfo = workflowPageInfo(pages.runs, len(pagedRuns), nextRunToken)
	return summary
}

func mergeLiveExecutionRuns(projected []*workflowv1.WorkflowRunProjection, executions []*workflowv1.WorkflowExecution) []*workflowv1.WorkflowRunProjection {
	out := append([]*workflowv1.WorkflowRunProjection{}, projected...)
	seen := make(map[string]struct{}, len(out)*2)
	for _, run := range out {
		remember(seen, run.GetRunId())
		remember(seen, run.GetExecutionId())
	}
	for _, execution := range executions {
		status := execution.GetStatus()
		if !isRealtimeRunStatus(status) || execution.GetId() == "" {
			continue
		}
		if _, ok := seen[execution.GetId()]; ok {
			continue
		}
		run := liveExecutionRun(execution, status)
		out = append(out, run)
		remember(seen, run.GetRunId())
		remember(seen, run.GetExecutionId())
	}
	sort.SliceStable(out, func(i, j int) bool {
		return timestampUnix(out[i].GetUpdatedAt()) > timestampUnix(out[j].GetUpdatedAt())
	})
	return out
}

func liveExecutionRun(execution *workflowv1.WorkflowExecution, status workflowv1.WorkflowRunStatus) *workflowv1.WorkflowRunProjection {
	started := execution.GetStartedAt()
	updated := execution.GetStoppedAt()
	if updated == nil {
		updated = timestamppb.New(time.Now())
	}
	nodeID, nodeName := currentExecutionNode(execution.GetGraphNodes())
	return &workflowv1.WorkflowRunProjection{
		RunId:           execution.GetId(),
		WorkflowId:      execution.GetWorkflowId(),
		WorkflowName:    execution.GetWorkflowName(),
		ExecutionId:     execution.GetId(),
		Status:          status,
		CurrentNodeId:   nodeID,
		CurrentNodeName: nodeName,
		StartedAt:       started,
		UpdatedAt:       updated,
		GraphNodes:      execution.GetGraphNodes(),
		GraphEdges:      execution.GetGraphEdges(),
	}
}

func currentExecutionNode(nodes []*workflowv1.WorkflowGraphNode) (string, string) {
	for _, node := range nodes {
		if node.GetStatus() == workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_RUNNING {
			return node.GetId(), node.GetName()
		}
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		if nodes[i].GetStatus() != workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_STATUS_UNSPECIFIED &&
			nodes[i].GetStatus() != workflowv1.WorkflowGraphElementStatus_WORKFLOW_GRAPH_ELEMENT_PENDING {
			return nodes[i].GetId(), nodes[i].GetName()
		}
	}
	return "", ""
}

func isRealtimeRunStatus(status workflowv1.WorkflowRunStatus) bool {
	switch status {
	case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_PENDING,
		workflowv1.WorkflowRunStatus_WORKFLOW_RUN_RUNNING,
		workflowv1.WorkflowRunStatus_WORKFLOW_RUN_WAITING:
		return true
	default:
		return false
	}
}

func runtimeTimestamp(value string) *timestamppb.Timestamp {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, trimmed); err == nil {
		return timestamppb.New(parsed)
	}
	return nil
}

func timestampUnix(value *timestamppb.Timestamp) int64 {
	if value == nil {
		return 0
	}
	if err := value.CheckValid(); err != nil {
		return 0
	}
	return value.AsTime().Unix()
}

func remember(index map[string]struct{}, value string) {
	if strings.TrimSpace(value) != "" {
		index[value] = struct{}{}
	}
}
