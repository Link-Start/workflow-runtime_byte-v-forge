package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
	"google.golang.org/protobuf/proto"
)

type workflowProjectionStore struct {
	mu          sync.RWMutex
	runs        map[string]*workflowv1.WorkflowRunProjection
	subscribers map[chan struct{}]struct{}
}

func newWorkflowProjectionStore() *workflowProjectionStore {
	return &workflowProjectionStore{runs: map[string]*workflowv1.WorkflowRunProjection{}, subscribers: map[chan struct{}]struct{}{}}
}

func (s *workflowProjectionStore) list() []*workflowv1.WorkflowRunProjection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*workflowv1.WorkflowRunProjection, 0, len(s.runs))
	for _, run := range s.runs {
		out = append(out, proto.Clone(run).(*workflowv1.WorkflowRunProjection))
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetUpdatedAtUnix() > out[j].GetUpdatedAtUnix() })
	return out
}

func paginateWorkflowRuns(runs []*workflowv1.WorkflowRunProjection, req workflowPageRequest) ([]*workflowv1.WorkflowRunProjection, string) {
	start := pageOffset(req.token)
	if start >= len(runs) {
		return nil, ""
	}
	end := start + int(req.size)
	if end > len(runs) {
		end = len(runs)
	}
	next := ""
	if end < len(runs) {
		next = strconv.Itoa(end)
	}
	return runs[start:end], next
}

func (s *workflowProjectionStore) apply(req *workflowv1.WorkflowStepUpdateRequest) *workflowv1.WorkflowRunProjection {
	now := time.Now().Unix()
	occurred := req.GetOccurredAtUnix()
	if occurred <= 0 {
		occurred = now
	}
	runID := stableRunID(req)
	nodeID := stableNodeID(req)

	s.mu.Lock()
	run := s.runs[runID]
	if run == nil {
		run = &workflowv1.WorkflowRunProjection{RunId: runID, StartedAtUnix: occurred}
		s.runs[runID] = run
	}
	mergeRunIdentity(run, req)
	applyRunStatus(run, req.GetStatus(), occurred, req.GetErrorMessage())
	run.CurrentNodeId = nodeID
	run.CurrentNodeName = strings.TrimSpace(req.GetNodeName())
	run.UpdatedAtUnix = occurred
	applyNodeStatus(run, req, nodeID, occurred)
	out := proto.Clone(run).(*workflowv1.WorkflowRunProjection)
	s.mu.Unlock()

	s.notify()
	return out
}

func (s *workflowProjectionStore) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.subscribers[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *workflowProjectionStore) unsubscribe(ch chan struct{}) {
	s.mu.Lock()
	delete(s.subscribers, ch)
	close(ch)
	s.mu.Unlock()
}

func (s *workflowProjectionStore) notify() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func stableRunID(req *workflowv1.WorkflowStepUpdateRequest) string {
	for _, value := range []string{req.GetRunId(), req.GetExecutionId(), req.GetWorkflowId()} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "workflow-run"
}

func stableNodeID(req *workflowv1.WorkflowStepUpdateRequest) string {
	if trimmed := strings.TrimSpace(req.GetNodeId()); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(req.GetNodeName())
}

func mergeRunIdentity(run *workflowv1.WorkflowRunProjection, req *workflowv1.WorkflowStepUpdateRequest) {
	if req.GetWorkflowId() != "" {
		run.WorkflowId = req.GetWorkflowId()
	}
	if req.GetWorkflowName() != "" {
		run.WorkflowName = req.GetWorkflowName()
	}
	if req.GetExecutionId() != "" {
		run.ExecutionId = req.GetExecutionId()
	}
}
