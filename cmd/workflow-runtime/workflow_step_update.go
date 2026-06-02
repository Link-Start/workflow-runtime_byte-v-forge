package main

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/eventbus"
	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
)

const (
	workflowStepUpdateEventName    = "workflow.step.updated"
	workflowStepUpdateTokenHeader  = "X-Workflow-Runtime-Token"
	workflowStepUpdateBearerPrefix = "Bearer "
)

var errWorkflowStepContextInvalid = errors.New("workflow step update event context is invalid")

func (s *dashboardServer) authorizeWorkflowStepUpdate(r *http.Request) bool {
	if s == nil || strings.TrimSpace(s.stepUpdateToken) == "" {
		return false
	}
	provided := workflowStepUpdateToken(r)
	if provided == "" {
		return false
	}
	expected := []byte(s.stepUpdateToken)
	actual := []byte(provided)
	return len(expected) == len(actual) && subtle.ConstantTimeCompare(expected, actual) == 1
}

func workflowStepUpdateToken(r *http.Request) string {
	if r == nil {
		return ""
	}
	if token := strings.TrimSpace(r.Header.Get(workflowStepUpdateTokenHeader)); token != "" {
		return token
	}
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(authorization, workflowStepUpdateBearerPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(authorization, workflowStepUpdateBearerPrefix))
	}
	return ""
}

func validateWorkflowStepUpdate(req *workflowv1.WorkflowStepUpdateRequest) error {
	if req.GetRunId() == "" && req.GetExecutionId() == "" && req.GetWorkflowId() == "" {
		return errors.New("run_id or execution_id or workflow_id is required")
	}
	if req.GetNodeId() == "" && req.GetNodeName() == "" {
		return errors.New("node_id or node_name is required")
	}
	if err := eventbus.ValidateContext(req.GetContext()); err != nil {
		return errWorkflowStepContextInvalid
	}
	if strings.TrimSpace(req.GetContext().GetEventName()) != workflowStepUpdateEventName {
		return errWorkflowStepContextInvalid
	}
	if strings.TrimSpace(req.GetContext().GetEventVersion()) != eventbus.DefaultEventVersion {
		return errWorkflowStepContextInvalid
	}
	return nil
}
