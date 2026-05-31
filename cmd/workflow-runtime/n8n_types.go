package main

type n8nPage[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"nextCursor"`
}

type n8nWorkflow struct {
	ID        any              `json:"id"`
	Name      string           `json:"name"`
	Active    bool             `json:"active"`
	UpdatedAt string           `json:"updatedAt"`
	Tags      []n8nWorkflowTag `json:"tags"`
	Nodes     []n8nNode        `json:"nodes"`
	Conns     n8nConnections   `json:"connections"`
}

type n8nWorkflowTag struct {
	Name string `json:"name"`
}

type n8nNode struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	TypeVersion any       `json:"typeVersion"`
	Disabled    bool      `json:"disabled"`
	Position    []float64 `json:"position"`
}

type n8nConnections map[string]map[string][][]n8nConnection

type n8nConnection struct {
	Node  string `json:"node"`
	Type  string `json:"type"`
	Index int32  `json:"index"`
}

type n8nExecution struct {
	ID           any              `json:"id"`
	WorkflowID   any              `json:"workflowId"`
	WorkflowName string           `json:"workflowName"`
	Status       string           `json:"status"`
	Mode         string           `json:"mode"`
	StartedAt    string           `json:"startedAt"`
	StoppedAt    string           `json:"stoppedAt"`
	WorkflowData n8nWorkflowData  `json:"workflowData"`
	Data         n8nExecutionData `json:"data"`
}

type n8nWorkflowData struct {
	Name  string         `json:"name"`
	Nodes []n8nNode      `json:"nodes"`
	Conns n8nConnections `json:"connections"`
}

type n8nExecutionData struct {
	ResultData n8nResultData `json:"resultData"`
}

type n8nResultData struct {
	RunData          n8nRunData `json:"runData"`
	LastNodeExecuted string     `json:"lastNodeExecuted"`
	Error            n8nError   `json:"error"`
}

type n8nRunData map[string][]n8nTaskData

type n8nTaskData struct {
	StartTime       int64    `json:"startTime"`
	ExecutionTime   int64    `json:"executionTime"`
	ExecutionStatus string   `json:"executionStatus"`
	Error           n8nError `json:"error"`
}

type n8nError struct {
	Name        string   `json:"name"`
	Message     string   `json:"message"`
	Description string   `json:"description"`
	Messages    []string `json:"messages"`
}
