package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

type workflowRuntimeClient struct {
	internalURL string
	editorURL   string
	apiKey      string
	httpClient  *http.Client
}
type workflowRuntimeConfig struct {
	InternalURL string
	EditorURL   string
	APIKey      string
}

func newWorkflowRuntimeClient(cfg workflowRuntimeConfig) *workflowRuntimeClient {
	return &workflowRuntimeClient{
		internalURL: strings.TrimRight(cfg.InternalURL, "/"),
		editorURL:   strings.TrimRight(cfg.EditorURL, "/"),
		apiKey:      cfg.APIKey,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *workflowRuntimeClient) summary(ctx context.Context, pages workflowSummaryPageRequest) *workflowv1.WorkflowRuntimeSummary {
	out := &workflowv1.WorkflowRuntimeSummary{
		EngineStatus:       workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_UNAVAILABLE,
		EngineMessage:      "n8n internal URL is not configured",
		ApiStatus:          workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_UNCONFIGURED,
		ApiMessage:         "n8n API key is not configured",
		ApiConfigured:      false,
		EditorUrl:          c.editorURL,
		CheckedAtUnix:      time.Now().Unix(),
		ExecutionsPageInfo: workflowPageInfo(pages.executions, 0, ""),
	}
	if c.internalURL == "" {
		return out
	}

	if err := c.checkHealth(ctx); err != nil {
		out.EngineMessage = err.Error()
	} else {
		out.EngineStatus = workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_AVAILABLE
		out.EngineMessage = "n8n engine is reachable"
	}

	if c.apiKey == "" {
		return out
	}
	out.ApiConfigured = true
	out.ApiStatus = workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_AVAILABLE
	out.ApiMessage = "n8n API is reachable"

	workflows, err := c.listWorkflows(ctx)
	if err != nil {
		out.ApiStatus = workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_DEGRADED
		out.ApiMessage = err.Error()
		return out
	}
	executions, nextCursor, err := c.listExecutions(ctx, pages.executions)
	if err != nil {
		out.ApiStatus = workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_DEGRADED
		out.ApiMessage = err.Error()
		return out
	}
	out.Workflows = workflows
	out.Executions = executions
	out.ExecutionsPageInfo = workflowPageInfo(pages.executions, len(executions), nextCursor)
	return out
}

func (c *workflowRuntimeClient) checkHealth(ctx context.Context) error {
	if err := c.getNoAuth(ctx, "/healthz/readiness"); err == nil {
		return nil
	}
	return c.getNoAuth(ctx, "/healthz")
}

func (c *workflowRuntimeClient) getNoAuth(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(path), nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("n8n engine unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("n8n health returned HTTP %d", resp.StatusCode)
	}
	return nil
}
