package main

import (
	"net/http"
	"strconv"
	"strings"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

const (
	defaultWorkflowPageSize = 12
	maxWorkflowPageSize     = 50
)

type workflowSummaryPageRequest struct {
	runs              workflowPageRequest
	executions        workflowPageRequest
	includeRuns       bool
	includeWorkflows  bool
	includeExecutions bool
}

type workflowPageRequest struct {
	size  int32
	token string
}

func defaultWorkflowSummaryPageRequest() workflowSummaryPageRequest {
	page := workflowPageRequest{size: defaultWorkflowPageSize}
	return workflowSummaryPageRequest{runs: page, executions: page, includeRuns: true}
}

func workflowSummaryPageRequestFromHTTP(r *http.Request) workflowSummaryPageRequest {
	q := r.URL.Query()
	req := workflowSummaryPageRequest{
		runs:       workflowPageRequestFromQuery(q.Get("runs_page_size"), q.Get("runs_page_token")),
		executions: workflowPageRequestFromQuery(q.Get("executions_page_size"), q.Get("executions_page_token")),
	}
	req.includeRuns, req.includeWorkflows, req.includeExecutions = workflowSummaryIncludes(q["include"])
	return req
}

func workflowSummaryIncludes(values []string) (bool, bool, bool) {
	if len(values) == 0 {
		return true, false, false
	}
	var runs, workflows, executions bool
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			switch strings.ToLower(strings.TrimSpace(part)) {
			case "runs", "run", "live", "realtime":
				runs = true
			case "workflows", "workflow", "definitions", "definition":
				workflows = true
			case "executions", "execution":
				executions = true
			}
		}
	}
	return runs, workflows, executions
}

func workflowPageRequestFromQuery(sizeValue string, tokenValue string) workflowPageRequest {
	size := int32(defaultWorkflowPageSize)
	if parsed, err := strconv.Atoi(strings.TrimSpace(sizeValue)); err == nil && parsed > 0 {
		size = int32(parsed)
	}
	if size > maxWorkflowPageSize {
		size = maxWorkflowPageSize
	}
	return workflowPageRequest{size: size, token: strings.TrimSpace(tokenValue)}
}

func workflowPageInfo(req workflowPageRequest, itemCount int, nextToken string) *workflowv1.WorkflowRuntimePageInfo {
	return &workflowv1.WorkflowRuntimePageInfo{
		PageSize:      req.size,
		ItemCount:     int32(itemCount),
		NextPageToken: nextToken,
	}
}

func pageOffset(token string) int {
	offset, err := strconv.Atoi(strings.TrimSpace(token))
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}
