package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

type n8nPage[T any] struct {
	Data []T `json:"data"`
}

type n8nWorkflow struct {
	ID        any              `json:"id"`
	Name      string           `json:"name"`
	Active    bool             `json:"active"`
	UpdatedAt string           `json:"updatedAt"`
	Tags      []n8nWorkflowTag `json:"tags"`
}

type n8nWorkflowTag struct {
	Name string `json:"name"`
}

type n8nExecution struct {
	ID           any             `json:"id"`
	WorkflowID   any             `json:"workflowId"`
	WorkflowName string          `json:"workflowName"`
	Status       string          `json:"status"`
	Mode         string          `json:"mode"`
	StartedAt    string          `json:"startedAt"`
	StoppedAt    string          `json:"stoppedAt"`
	WorkflowData n8nWorkflowData `json:"workflowData"`
}

type n8nWorkflowData struct {
	Name string `json:"name"`
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

func (c *workflowRuntimeClient) summary(ctx context.Context) *workflowv1.WorkflowRuntimeSummary {
	out := &workflowv1.WorkflowRuntimeSummary{
		EngineStatus:  workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_UNAVAILABLE,
		EngineMessage: "n8n internal URL is not configured",
		ApiStatus:     workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_UNCONFIGURED,
		ApiMessage:    "n8n API key is not configured",
		ApiConfigured: false,
		EditorUrl:     c.editorURL,
		CheckedAtUnix: time.Now().Unix(),
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
	executions, err := c.listExecutions(ctx)
	if err != nil {
		out.ApiStatus = workflowv1.WorkflowRuntimeStatus_WORKFLOW_RUNTIME_DEGRADED
		out.ApiMessage = err.Error()
		return out
	}
	out.Workflows = workflows
	out.Executions = executions
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

func (c *workflowRuntimeClient) listWorkflows(ctx context.Context) ([]*workflowv1.WorkflowDefinition, error) {
	var page n8nPage[n8nWorkflow]
	if err := c.getAPI(ctx, "/api/v1/workflows?limit=50", &page); err != nil {
		return nil, err
	}
	workflows := make([]*workflowv1.WorkflowDefinition, 0, len(page.Data))
	for _, item := range page.Data {
		workflows = append(workflows, &workflowv1.WorkflowDefinition{
			Id:        valueString(item.ID),
			Name:      item.Name,
			Active:    item.Active,
			UpdatedAt: item.UpdatedAt,
			Tags:      workflowTagNames(item.Tags),
		})
	}
	return workflows, nil
}

func (c *workflowRuntimeClient) listExecutions(ctx context.Context) ([]*workflowv1.WorkflowExecution, error) {
	var page n8nPage[n8nExecution]
	if err := c.getAPI(ctx, "/api/v1/executions?limit=25&includeData=false", &page); err != nil {
		return nil, err
	}
	executions := make([]*workflowv1.WorkflowExecution, 0, len(page.Data))
	for _, item := range page.Data {
		name := item.WorkflowName
		if name == "" {
			name = item.WorkflowData.Name
		}
		executions = append(executions, &workflowv1.WorkflowExecution{
			Id:           valueString(item.ID),
			WorkflowId:   valueString(item.WorkflowID),
			WorkflowName: name,
			Status:       item.Status,
			Mode:         item.Mode,
			StartedAt:    item.StartedAt,
			StoppedAt:    item.StoppedAt,
		})
	}
	return executions, nil
}

func (c *workflowRuntimeClient) getAPI(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(path), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-N8N-API-KEY", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("n8n API unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("n8n API returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode n8n API response: %w", err)
	}
	return nil
}

func (c *workflowRuntimeClient) endpoint(path string) string {
	parsed, err := url.Parse(path)
	if err == nil && parsed.IsAbs() {
		return parsed.String()
	}
	return c.internalURL + "/" + strings.TrimLeft(path, "/")
}

func workflowTagNames(tags []n8nWorkflowTag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag.Name != "" {
			names = append(names, tag.Name)
		}
	}
	return names
}

func valueString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}
