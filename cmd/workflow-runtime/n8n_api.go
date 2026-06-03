package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

func (c *workflowRuntimeClient) listWorkflows(ctx context.Context) ([]*workflowv1.WorkflowDefinition, error) {
	var page n8nPage[n8nWorkflow]
	if err := c.getAPI(ctx, "/api/v1/workflows?limit=50", &page); err != nil {
		return nil, err
	}
	workflows := make([]*workflowv1.WorkflowDefinition, 0, len(page.Data))
	for _, item := range page.Data {
		workflows = append(workflows, &workflowv1.WorkflowDefinition{
			Id:         valueString(item.ID),
			Name:       item.Name,
			Active:     item.Active,
			UpdatedAt:  runtimeTimestamp(item.UpdatedAt),
			Tags:       workflowTagNames(item.Tags),
			GraphNodes: workflowGraphNodes(item.Nodes, "", nil, n8nResultData{}),
			GraphEdges: workflowGraphEdges(item.Nodes, item.Conns, nil),
		})
	}
	return workflows, nil
}

func (c *workflowRuntimeClient) listExecutions(ctx context.Context, req workflowPageRequest) ([]*workflowv1.WorkflowExecution, string, error) {
	var page n8nPage[n8nExecution]
	if err := c.getAPI(ctx, c.executionsPath(req), &page); err != nil {
		return nil, "", err
	}
	executions := make([]*workflowv1.WorkflowExecution, 0, len(page.Data))
	for _, item := range page.Data {
		name := item.WorkflowName
		if name == "" {
			name = item.WorkflowData.Name
		}
		result := item.Data.ResultData
		executions = append(executions, &workflowv1.WorkflowExecution{
			Id:           valueString(item.ID),
			WorkflowId:   valueString(item.WorkflowID),
			WorkflowName: name,
			Status:       runStatusFromString(item.Status),
			Mode:         item.Mode,
			StartedAt:    runtimeTimestamp(item.StartedAt),
			StoppedAt:    runtimeTimestamp(item.StoppedAt),
			GraphNodes:   workflowGraphNodes(item.WorkflowData.Nodes, item.Status, result.RunData, result),
			GraphEdges:   workflowGraphEdges(item.WorkflowData.Nodes, item.WorkflowData.Conns, result.RunData),
		})
	}
	return executions, page.NextCursor, nil
}

func (c *workflowRuntimeClient) executionsPath(req workflowPageRequest) string {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(int(req.size)))
	q.Set("includeData", "true")
	if req.token != "" {
		q.Set("cursor", req.token)
	}
	return "/api/v1/executions?" + q.Encode()
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
