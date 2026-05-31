import { useEffect, useMemo, useState } from 'react';
import { ExternalLink, RefreshCw, Workflow } from 'lucide-react';
import { Button, WorkspacePanel, api, createHotStreamURL, useCursorPages, useHotStreamInvalidation, type WorkflowRuntimeSummary } from '@byte-v-forge/common-ui';
import { WorkflowGraphPanel } from './workflow-graph-panel';
import { WorkflowSidebar, type WorkflowSelectionKind } from './workflow-sidebar';

const WORKFLOW_PAGE_SIZE = 12;

export function WorkflowPage() {
  const [tab, setTab] = useState<WorkflowSelectionKind>('runs');
  const [selectedID, setSelectedID] = useState('');
  const summaryRootKey = ['workflow-runtime'] as const;
  useEffect(() => {
    const target = selectionFromSearch(location.search);
    if (!target) return;
    setTab(target.tab);
    setSelectedID(target.id);
  }, []);
  useHotStreamInvalidation({
    url: createHotStreamURL('/api/workflow-runtime', { eventTypes: ['workflow-runtime.summary.updated'], resourceTypes: ['workflow-runtime'] }),
    rules: [{ queryKey: summaryRootKey, eventTypes: ['workflow-runtime.summary.updated'], resourceTypes: ['workflow-runtime'] }]
  });
  const runQuery = useCursorPages<WorkflowRuntimeSummary>({
    queryKey: ['workflow-runtime', 'runs'],
    queryFn: (cursor) => api<WorkflowRuntimeSummary>(summaryURL(cursor, '')),
    nextCursor: (page) => page.runs_page_info?.next_page_token || ''
  });
  const executionQuery = useCursorPages<WorkflowRuntimeSummary>({
    queryKey: ['workflow-runtime', 'executions'],
    queryFn: (cursor) => api<WorkflowRuntimeSummary>(summaryURL('', cursor)),
    nextCursor: (page) => page.executions_page_info?.next_page_token || ''
  });
  const summary = useMemo(() => mergeSummary(runQuery.pages, executionQuery.pages), [runQuery.pages, executionQuery.pages]);
  const selection = useMemo(() => selectGraphScope(summary, tab, selectedID), [summary, tab, selectedID]);

  return (
    <WorkspacePanel panelClassName="workflowRuntimePanel">
      <div className="panelHeader">
        <div><Workflow size={17} />Workflow / n8n</div>
        <div className="headerControls">
          <Button variant="outline" size="icon-sm" title="刷新" aria-label="刷新" onClick={() => { runQuery.refetch(); executionQuery.refetch(); }} disabled={runQuery.isFetching || executionQuery.isFetching}><RefreshCw /></Button>
          {summary?.editor_url && <Button asChild variant="outline" size="sm"><a href={summary.editor_url} target="_blank" rel="noreferrer"><ExternalLink />n8n</a></Button>}
        </div>
      </div>
      <div className="workflowRuntimeLayout">
        <WorkflowSidebar
          tab={tab}
          workflows={summary?.workflows ?? []}
          runs={summary?.runs ?? []}
          executions={summary?.executions ?? []}
          runsPageInfo={summary?.runs_page_info}
          executionsPageInfo={summary?.executions_page_info}
          selectedID={selectedID}
          loading={runQuery.isLoading || executionQuery.isLoading}
          loadingMoreRuns={runQuery.pagination.loading}
          loadingMoreExecutions={executionQuery.pagination.loading}
          apiConfigured={summary?.api_configured ?? false}
          onTabChange={(next) => { setTab(next); setSelectedID(''); }}
          onRunSelect={(run) => { setTab('runs'); setSelectedID(run.run_id); }}
          onWorkflowSelect={(workflow) => { setTab('workflows'); setSelectedID(workflow.id); }}
          onExecutionSelect={(execution) => { setTab('executions'); setSelectedID(execution.id); }}
          onNextRuns={runQuery.loadMore}
          onNextExecutions={executionQuery.loadMore}
        />
        <WorkflowGraphPanel {...selection} workflows={summary?.workflows ?? []} />
      </div>
    </WorkspacePanel>
  );
}

function summaryURL(runsToken: string, executionsToken: string) {
  const params = new URLSearchParams();
  appendPage(params, 'runs', runsToken);
  appendPage(params, 'executions', executionsToken);
  return `/api/workflow-runtime/summary?${params.toString()}`;
}

function appendPage(params: URLSearchParams, name: string, token: string) {
  params.set(`${name}_page_size`, String(WORKFLOW_PAGE_SIZE));
  if (token) params.set(`${name}_page_token`, token);
}

function mergeSummary(runPages?: WorkflowRuntimeSummary[], executionPages?: WorkflowRuntimeSummary[]) {
  const first = runPages?.[0] || executionPages?.[0];
  if (!first) return undefined;
  return {
    ...first,
    workflows: first.workflows || [],
    runs: (runPages || [first]).flatMap((page) => page.runs || []),
    executions: (executionPages || [first]).flatMap((page) => page.executions || []),
    runs_page_info: lastRunPageInfo(runPages, first.runs_page_info),
    executions_page_info: lastExecutionPageInfo(executionPages, first.executions_page_info)
  };
}

function lastRunPageInfo(pages: WorkflowRuntimeSummary[] | undefined, fallback: WorkflowRuntimeSummary['runs_page_info']) {
  return pages?.[pages.length - 1]?.runs_page_info || fallback;
}

function lastExecutionPageInfo(pages: WorkflowRuntimeSummary[] | undefined, fallback: WorkflowRuntimeSummary['executions_page_info']) {
  return pages?.[pages.length - 1]?.executions_page_info || fallback;
}

function selectionFromSearch(search: string): { tab: WorkflowSelectionKind; id: string } | null {
  const params = new URLSearchParams(search);
  const executionID = params.get('execution_id');
  if (executionID) return { tab: 'executions', id: executionID };
  const runID = params.get('run_id');
  if (runID) return { tab: 'runs', id: runID };
  const workflowID = params.get('workflow_id');
  if (workflowID) return { tab: 'workflows', id: workflowID };
  return null;
}

function selectGraphScope(summary: WorkflowRuntimeSummary | undefined, tab: WorkflowSelectionKind, selectedID: string) {
  const runs = summary?.runs || [];
  const executions = summary?.executions || [];
  const workflows = summary?.workflows || [];
  const run = tab === 'runs' ? runs.find((item) => item.run_id === selectedID) || runs[0] : undefined;
  const execution = tab === 'executions' ? executions.find((item) => item.id === selectedID) || executions[0] : undefined;
  const workflowID = run?.workflow_id || execution?.workflow_id || selectedID;
  const workflow = workflows.find((item) => item.id === workflowID) || workflows[0] || null;
  if (run) return { run, workflow };
  if (execution) return { execution, workflow };
  return { workflow };
}
