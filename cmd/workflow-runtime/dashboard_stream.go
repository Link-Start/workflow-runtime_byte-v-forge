package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	observabilityv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/observability/v1"
	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const workflowSummaryEventType = "workflow-runtime.summary.updated"

func (s *dashboardServer) handleWorkflowRuntimeStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	prepareSSE(w)
	_ = writeControlSSE(w, observabilityv1.HotStreamControlKind_HOT_STREAM_CONTROL_KIND_CONNECTED, "workflow-runtime stream connected", nil)
	flusher.Flush()

	ctx := r.Context()
	sub := s.projections.subscribe()
	defer s.projections.unsubscribe(sub)
	last := ""
	pages := defaultWorkflowSummaryPageRequest()
	for {
		summary := s.workflowSummaryForContext(ctx, pages)
		fingerprint := workflowSummaryFingerprint(summary)
		if fingerprint != last {
			last = fingerprint
			_ = writeEventSSE(w, workflowSummaryHotStreamEvent(summary))
			flusher.Flush()
		} else {
			_ = writeControlSSE(w, observabilityv1.HotStreamControlKind_HOT_STREAM_CONTROL_KIND_HEARTBEAT, "heartbeat", nil)
			flusher.Flush()
		}

		if !waitWorkflowStream(ctx, sub, workflowStreamInterval(summary)) {
			return
		}
	}
}

func prepareSSE(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
}

func workflowSummaryHotStreamEvent(summary *workflowv1.WorkflowRuntimeSummary) *observabilityv1.HotStreamEvent {
	return &observabilityv1.HotStreamEvent{
		EventId:       fmt.Sprintf("workflow-runtime-summary-%d", time.Now().UnixNano()),
		EventType:     workflowSummaryEventType,
		SourceService: "workflow-runtime",
		ResourceType:  "workflow-runtime",
		ResourceId:    "summary",
		Scope:         "platform",
		OccurredAt:    timestamppb.Now(),
		Attributes: map[string]string{
			"api_status":      summary.GetApiStatus().String(),
			"engine_status":   summary.GetEngineStatus().String(),
			"live_executions": fmt.Sprint(workflowSummaryHasLiveExecution(summary)),
			"execution_count": fmt.Sprint(len(summary.GetExecutions())),
			"workflow_count":  fmt.Sprint(len(summary.GetWorkflows())),
			"run_count":       fmt.Sprint(len(summary.GetRuns())),
		},
	}
}

func workflowSummaryFingerprint(summary *workflowv1.WorkflowRuntimeSummary) string {
	clone := proto.Clone(summary).(*workflowv1.WorkflowRuntimeSummary)
	clone.CheckedAtUnix = 0
	payload, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(clone)
	if err != nil {
		payload = []byte(fmt.Sprintf("%v", clone))
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func workflowStreamInterval(summary *workflowv1.WorkflowRuntimeSummary) time.Duration {
	if workflowSummaryHasLiveExecution(summary) {
		return 2 * time.Second
	}
	return 15 * time.Second
}

func workflowSummaryHasLiveExecution(summary *workflowv1.WorkflowRuntimeSummary) bool {
	for _, run := range summary.GetRuns() {
		switch run.GetStatus() {
		case workflowv1.WorkflowRunStatus_WORKFLOW_RUN_PENDING, workflowv1.WorkflowRunStatus_WORKFLOW_RUN_RUNNING, workflowv1.WorkflowRunStatus_WORKFLOW_RUN_WAITING:
			return true
		}
	}
	for _, execution := range summary.GetExecutions() {
		switch strings.ToLower(strings.TrimSpace(execution.GetStatus())) {
		case "new", "running", "waiting":
			return true
		}
	}
	return false
}

func writeEventSSE(w http.ResponseWriter, event *observabilityv1.HotStreamEvent) error {
	return writeProtoSSE(w, "hotstream", event)
}

func writeControlSSE(w http.ResponseWriter, kind observabilityv1.HotStreamControlKind, message string, attrs map[string]string) error {
	return writeProtoSSE(w, "hotstream.control", &observabilityv1.HotStreamControlEvent{Kind: kind, Message: message, OccurredAt: timestamppb.Now(), Attributes: attrs})
}

func writeProtoSSE(w http.ResponseWriter, eventName string, message proto.Message) error {
	payload, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(message)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, payload)
	return err
}

func waitWorkflowStream(ctx context.Context, changed <-chan struct{}, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-changed:
		return true
	case <-timer.C:
		return true
	}
}
