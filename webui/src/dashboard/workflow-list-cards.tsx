import { Badge, CursorPager, RecordCard, RecordField, RecordIdentity, RecordList, formatUnix } from '@byte-v-forge/common-ui';
import type { WorkflowDefinition, WorkflowExecution, WorkflowRuntimePageInfo, WorkflowRunProjection } from '@byte-v-forge/common-ui';
import { GitBranch, History, RadioTower } from 'lucide-react';
import { badgeVariant, runStatusLabel } from './workflow-status';

export function RunCards({ runs, pageInfo, loadingMore, selectedID, onSelect, onNext }: {
  runs: WorkflowRunProjection[];
  pageInfo?: WorkflowRuntimePageInfo;
  loadingMore: boolean;
  selectedID?: string;
  onSelect: (run: WorkflowRunProjection) => void;
  onNext: () => void;
}) {
  return (
    <div className="workflowPagedList">
      <RecordList emptyText="暂无实时投影；运行中的流程会在这里显示。" className="workflowRecordList">
        {runs.map((run) => <RunCard key={run.run_id} run={run} selected={run.run_id === selectedID} onSelect={onSelect} />)}
      </RecordList>
      <CursorPager itemCount={runs.length} pageSize={pageInfo?.page_size} hasNext={Boolean(pageInfo?.next_page_token)} loading={loadingMore} onNext={onNext} />
    </div>
  );
}

function RunCard({ run, selected, onSelect }: { run: WorkflowRunProjection; selected: boolean; onSelect: (run: WorkflowRunProjection) => void }) {
  return (
    <RecordCard selected={selected} onClick={() => onSelect(run)} className="workflowRecordCard">
      <RecordIdentity icon={<RadioTower size={15} />} title={run.workflow_name || run.workflow_id || run.run_id} subtitle={run.run_id} />
      <div className="workflowRecordFields">
        <RecordField label="状态"><Badge variant={badgeVariant(run.status)}>{runStatusLabel(run.status)}</Badge></RecordField>
        <RecordField label="节点" value={run.current_node_name || run.current_node_id || '-'} />
        <RecordField label="更新" value={formatUnix(run.updated_at_unix)} />
      </div>
    </RecordCard>
  );
}

export function WorkflowCards({ workflows, selectedID, loading, apiConfigured, onSelect }: {
  workflows: WorkflowDefinition[];
  selectedID?: string;
  loading: boolean;
  apiConfigured: boolean;
  onSelect: (workflow: WorkflowDefinition) => void;
}) {
  return (
    <RecordList emptyText={emptyWorkflowText(loading, apiConfigured)} className="workflowRecordList">
      {workflows.map((workflow) => <WorkflowCard key={workflow.id} workflow={workflow} selected={workflow.id === selectedID} onSelect={onSelect} />)}
    </RecordList>
  );
}

function WorkflowCard({ workflow, selected, onSelect }: { workflow: WorkflowDefinition; selected: boolean; onSelect: (workflow: WorkflowDefinition) => void }) {
  return (
    <RecordCard selected={selected} onClick={() => onSelect(workflow)} className="workflowRecordCard">
      <RecordIdentity icon={<GitBranch size={15} />} title={workflow.name || workflow.id} subtitle={workflow.id} />
      <div className="workflowRecordFields workflowDefinitionFields">
        <RecordField label="状态"><Badge variant={workflow.active ? 'default' : 'outline'}>{workflow.active ? '启用' : '停用'}</Badge></RecordField>
        <RecordField label="节点" value={workflow.graph_nodes.length} />
      </div>
    </RecordCard>
  );
}

export function ExecutionCards({ executions, pageInfo, loadingMore, selectedID, loading, apiConfigured, onSelect, onNext }: {
  executions: WorkflowExecution[];
  pageInfo?: WorkflowRuntimePageInfo;
  loadingMore: boolean;
  selectedID?: string;
  loading: boolean;
  apiConfigured: boolean;
  onSelect: (execution: WorkflowExecution) => void;
  onNext: () => void;
}) {
  return (
    <div className="workflowPagedList">
      <RecordList emptyText={emptyExecutionText(loading, apiConfigured)} className="workflowRecordList">
        {executions.map((execution) => <ExecutionCard key={execution.id} execution={execution} selected={execution.id === selectedID} onSelect={onSelect} />)}
      </RecordList>
      <CursorPager itemCount={executions.length} pageSize={pageInfo?.page_size} hasNext={Boolean(pageInfo?.next_page_token)} loading={loadingMore} onNext={onNext} />
    </div>
  );
}

function ExecutionCard({ execution, selected, onSelect }: { execution: WorkflowExecution; selected: boolean; onSelect: (execution: WorkflowExecution) => void }) {
  return (
    <RecordCard selected={selected} onClick={() => onSelect(execution)} className="workflowRecordCard">
      <RecordIdentity icon={<History size={15} />} title={execution.workflow_name || execution.workflow_id || execution.id} subtitle={`#${execution.id}`} />
      <div className="workflowRecordFields">
        <RecordField label="状态"><Badge variant={badgeVariant(execution.status)}>{runStatusLabel(execution.status)}</Badge></RecordField>
        <RecordField label="模式" value={execution.mode || '-'} />
        <RecordField label="开始" value={execution.started_at || '-'} />
      </div>
    </RecordCard>
  );
}

function emptyWorkflowText(loading: boolean, apiConfigured: boolean) {
  if (loading) return '加载中…';
  return apiConfigured ? '暂无流程定义。' : '配置 n8n API key 后自动展示流程定义。';
}

function emptyExecutionText(loading: boolean, apiConfigured: boolean) {
  if (loading) return '加载中…';
  return apiConfigured ? '暂无最近执行；平台不做历史回放。' : '暂无执行投影；平台不做历史回放。';
}
