package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	workflowv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/workflow/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
)

const workflowProjectionSchema = `
CREATE TABLE IF NOT EXISTS workflow_runtime_run_projections (
  run_id text PRIMARY KEY,
  updated_at_unix bigint NOT NULL,
  projection bytea NOT NULL
);

CREATE INDEX IF NOT EXISTS workflow_runtime_run_projections_updated_idx
  ON workflow_runtime_run_projections (updated_at_unix DESC, run_id ASC);

CREATE TABLE IF NOT EXISTS workflow_runtime_step_updates (
  idempotency_key text PRIMARY KEY,
  run_id text NOT NULL,
  occurred_at_unix bigint NOT NULL,
  created_at_unix bigint NOT NULL
);

CREATE INDEX IF NOT EXISTS workflow_runtime_step_updates_run_idx
  ON workflow_runtime_step_updates (run_id, occurred_at_unix DESC);
`

var (
	errWorkflowDatabaseConfig      = errors.New("workflow runtime database config is invalid")
	errWorkflowDatabaseUnavailable = errors.New("workflow runtime database is unavailable")
)

type workflowProjectionStoreConfig struct {
	DatabaseURL string
}

type workflowProjectionStore struct {
	pool        *pgxpool.Pool
	mu          sync.RWMutex
	subscribers map[chan struct{}]struct{}
}

type workflowStepApplyResult struct {
	run       *workflowv1.WorkflowRunProjection
	duplicate bool
}

func newWorkflowProjectionStore(ctx context.Context, cfg workflowProjectionStoreConfig) (*workflowProjectionStore, error) {
	databaseURL := strings.TrimSpace(cfg.DatabaseURL)
	if databaseURL == "" {
		return nil, errWorkflowDatabaseConfig
	}
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errWorkflowDatabaseConfig
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, errWorkflowDatabaseUnavailable
	}
	store := &workflowProjectionStore{pool: pool, subscribers: map[chan struct{}]struct{}{}}
	if err := store.ensureSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *workflowProjectionStore) close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *workflowProjectionStore) ensureSchema(ctx context.Context) error {
	if s == nil || s.pool == nil {
		return errWorkflowDatabaseUnavailable
	}
	if err := s.pool.Ping(ctx); err != nil {
		return errWorkflowDatabaseUnavailable
	}
	if _, err := s.pool.Exec(ctx, workflowProjectionSchema); err != nil {
		return errWorkflowDatabaseUnavailable
	}
	return nil
}

func (s *workflowProjectionStore) list(ctx context.Context) ([]*workflowv1.WorkflowRunProjection, error) {
	if s == nil || s.pool == nil {
		return nil, errWorkflowDatabaseUnavailable
	}
	rows, err := s.pool.Query(ctx, `SELECT projection FROM workflow_runtime_run_projections ORDER BY updated_at_unix DESC, run_id ASC`)
	if err != nil {
		return nil, errWorkflowDatabaseUnavailable
	}
	defer rows.Close()

	out := []*workflowv1.WorkflowRunProjection{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, errWorkflowDatabaseUnavailable
		}
		run := &workflowv1.WorkflowRunProjection{}
		if err := proto.Unmarshal(payload, run); err != nil {
			return nil, fmt.Errorf("decode workflow run projection: %w", err)
		}
		out = append(out, run)
	}
	if err := rows.Err(); err != nil {
		return nil, errWorkflowDatabaseUnavailable
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].GetUpdatedAtUnix() > out[j].GetUpdatedAtUnix() })
	return out, nil
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

func (s *workflowProjectionStore) apply(ctx context.Context, req *workflowv1.WorkflowStepUpdateRequest) (workflowStepApplyResult, error) {
	if s == nil || s.pool == nil {
		return workflowStepApplyResult{}, errWorkflowDatabaseUnavailable
	}
	now := time.Now().Unix()
	occurred := workflowOccurredAtUnix(req, now)
	runID := stableRunID(req)
	idempotencyKey := strings.TrimSpace(req.GetContext().GetIdempotencyKey())
	if idempotencyKey == "" {
		return workflowStepApplyResult{}, errWorkflowStepContextInvalid
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return workflowStepApplyResult{}, errWorkflowDatabaseUnavailable
	}
	defer rollbackWorkflowProjectionTx(ctx, tx)

	inserted, storedRunID, err := insertWorkflowStepUpdate(ctx, tx, idempotencyKey, runID, occurred, now)
	if err != nil {
		return workflowStepApplyResult{}, err
	}
	if !inserted {
		run, err := loadWorkflowProjectionForUpdate(ctx, tx, storedRunID)
		if err != nil {
			return workflowStepApplyResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return workflowStepApplyResult{}, errWorkflowDatabaseUnavailable
		}
		return workflowStepApplyResult{run: run, duplicate: true}, nil
	}

	run, err := loadWorkflowProjectionForUpdate(ctx, tx, runID)
	if err != nil {
		return workflowStepApplyResult{}, err
	}
	if run == nil {
		run = &workflowv1.WorkflowRunProjection{RunId: runID, StartedAtUnix: occurred}
	}
	applyWorkflowStepUpdate(run, req, stableNodeID(req), occurred)
	if err := saveWorkflowProjection(ctx, tx, run); err != nil {
		return workflowStepApplyResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return workflowStepApplyResult{}, errWorkflowDatabaseUnavailable
	}

	out := proto.Clone(run).(*workflowv1.WorkflowRunProjection)
	s.notify()
	return workflowStepApplyResult{run: out}, nil
}

func insertWorkflowStepUpdate(ctx context.Context, tx pgx.Tx, idempotencyKey string, runID string, occurred int64, created int64) (bool, string, error) {
	var storedRunID string
	err := tx.QueryRow(ctx, `
INSERT INTO workflow_runtime_step_updates (idempotency_key, run_id, occurred_at_unix, created_at_unix)
VALUES ($1, $2, $3, $4)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING run_id`, idempotencyKey, runID, occurred, created).Scan(&storedRunID)
	if err == nil {
		return true, storedRunID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, "", errWorkflowDatabaseUnavailable
	}
	if err := tx.QueryRow(ctx, `SELECT run_id FROM workflow_runtime_step_updates WHERE idempotency_key = $1`, idempotencyKey).Scan(&storedRunID); err != nil {
		return false, "", errWorkflowDatabaseUnavailable
	}
	return false, storedRunID, nil
}

func loadWorkflowProjectionForUpdate(ctx context.Context, tx pgx.Tx, runID string) (*workflowv1.WorkflowRunProjection, error) {
	var payload []byte
	err := tx.QueryRow(ctx, `SELECT projection FROM workflow_runtime_run_projections WHERE run_id = $1 FOR UPDATE`, runID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errWorkflowDatabaseUnavailable
	}
	run := &workflowv1.WorkflowRunProjection{}
	if err := proto.Unmarshal(payload, run); err != nil {
		return nil, fmt.Errorf("decode workflow run projection: %w", err)
	}
	return run, nil
}

func saveWorkflowProjection(ctx context.Context, tx pgx.Tx, run *workflowv1.WorkflowRunProjection) error {
	payload, err := proto.Marshal(run)
	if err != nil {
		return fmt.Errorf("encode workflow run projection: %w", err)
	}
	_, err = tx.Exec(ctx, `
INSERT INTO workflow_runtime_run_projections (run_id, updated_at_unix, projection)
VALUES ($1, $2, $3)
ON CONFLICT (run_id) DO UPDATE SET
  updated_at_unix = EXCLUDED.updated_at_unix,
  projection = EXCLUDED.projection`, run.GetRunId(), run.GetUpdatedAtUnix(), payload)
	if err != nil {
		return errWorkflowDatabaseUnavailable
	}
	return nil
}

func rollbackWorkflowProjectionTx(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func applyWorkflowStepUpdate(run *workflowv1.WorkflowRunProjection, req *workflowv1.WorkflowStepUpdateRequest, nodeID string, occurred int64) {
	mergeRunIdentity(run, req)
	applyRunStatus(run, req.GetStatus(), occurred, req.GetErrorMessage())
	run.CurrentNodeId = nodeID
	run.CurrentNodeName = strings.TrimSpace(req.GetNodeName())
	run.UpdatedAtUnix = occurred
	applyNodeStatus(run, req, nodeID, occurred)
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

func workflowOccurredAtUnix(req *workflowv1.WorkflowStepUpdateRequest, fallback int64) int64 {
	if req.GetOccurredAtUnix() > 0 {
		return req.GetOccurredAtUnix()
	}
	if ts := req.GetContext().GetOccurredAt(); ts != nil && ts.IsValid() {
		return ts.AsTime().Unix()
	}
	return fallback
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
