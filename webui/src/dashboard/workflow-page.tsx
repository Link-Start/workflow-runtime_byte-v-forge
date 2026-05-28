import { ExternalLink, RefreshCw, Workflow } from 'lucide-react';
import {
  Badge,
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  WorkspacePanel,
  WorkflowRuntimeStatus,
  api,
  useQuery,
  type WorkflowDefinition,
  type WorkflowExecution,
  type WorkflowRuntimeSummary
} from '@byte-v-forge/common-ui';

export function WorkflowPage() {
  const summaryQuery = useQuery({
    queryKey: ['workflow-runtime', 'summary'],
    queryFn: () => api<WorkflowRuntimeSummary>('/api/workflow-runtime/summary'),
    refetchInterval: 15000
  });
  const summary = summaryQuery.data;

  return (
    <WorkspacePanel>
      <div className="panelHeader">
        <div><Workflow size={17} />Workflow / n8n</div>
        <div className="headerControls">
          <Button variant="outline" size="sm" onClick={() => summaryQuery.refetch()} disabled={summaryQuery.isFetching}>
            <RefreshCw />刷新
          </Button>
          {summary?.editor_url && (
            <Button asChild variant="outline" size="sm">
              <a href={summary.editor_url} target="_blank" rel="noreferrer">
                <ExternalLink />管理员编辑入口
              </a>
            </Button>
          )}
        </div>
      </div>

      <div className="grid gap-3 py-3 md:grid-cols-2">
        <StatusCard title="运行引擎" status={summary?.engine_status} message={summary?.engine_message} />
        <StatusCard title="n8n API" status={summary?.api_status} message={summary?.api_message} />
      </div>

      <div className="emptyBlock mb-3">
        业务侧使用平台原生页面查询/订阅运行状态；n8n editor 仅作为管理员编排入口，登录不进入业务主流程。
      </div>

      <div className="grid min-h-0 flex-1 gap-4 lg:grid-cols-2">
        <WorkflowTable workflows={summary?.workflows ?? []} apiConfigured={summary?.api_configured ?? false} loading={summaryQuery.isLoading} />
        <ExecutionTable executions={summary?.executions ?? []} apiConfigured={summary?.api_configured ?? false} loading={summaryQuery.isLoading} />
      </div>
    </WorkspacePanel>
  );
}

function StatusCard({ title, status, message }: { title: string; status?: WorkflowRuntimeStatus; message?: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-3">
      <div className="mb-2 flex items-center justify-between gap-2">
        <div className="font-semibold">{title}</div>
        <Badge variant={statusBadgeVariant(status)}>{statusLabel(status)}</Badge>
      </div>
      <div className="text-sm text-muted-foreground">{message || '等待状态刷新'}</div>
    </div>
  );
}

function WorkflowTable({ workflows, apiConfigured, loading }: { workflows: WorkflowDefinition[]; apiConfigured: boolean; loading: boolean }) {
  return (
    <section className="min-h-0 overflow-hidden rounded-lg border border-border">
      <div className="border-b border-border px-3 py-2 font-semibold">流程定义</div>
      {workflows.length === 0 ? (
        <div className="emptyBlock m-3">{emptyWorkflowText(loading, apiConfigured)}</div>
      ) : (
        <Table>
          <TableHeader><TableRow><TableHead>名称</TableHead><TableHead>状态</TableHead><TableHead>更新时间</TableHead></TableRow></TableHeader>
          <TableBody>{workflows.map((item) => <WorkflowRow key={item.id} workflow={item} />)}</TableBody>
        </Table>
      )}
    </section>
  );
}

function WorkflowRow({ workflow }: { workflow: WorkflowDefinition }) {
  return (
    <TableRow>
      <TableCell className="font-medium">{workflow.name || workflow.id}</TableCell>
      <TableCell><Badge variant={workflow.active ? 'default' : 'outline'}>{workflow.active ? '启用' : '停用'}</Badge></TableCell>
      <TableCell>{workflow.updated_at || '-'}</TableCell>
    </TableRow>
  );
}

function ExecutionTable({ executions, apiConfigured, loading }: { executions: WorkflowExecution[]; apiConfigured: boolean; loading: boolean }) {
  return (
    <section className="min-h-0 overflow-hidden rounded-lg border border-border">
      <div className="border-b border-border px-3 py-2 font-semibold">最近执行</div>
      {executions.length === 0 ? (
        <div className="emptyBlock m-3">{emptyExecutionText(loading, apiConfigured)}</div>
      ) : (
        <Table>
          <TableHeader><TableRow><TableHead>流程</TableHead><TableHead>状态</TableHead><TableHead>开始时间</TableHead></TableRow></TableHeader>
          <TableBody>{executions.map((item) => <ExecutionRow key={item.id} execution={item} />)}</TableBody>
        </Table>
      )}
    </section>
  );
}

function ExecutionRow({ execution }: { execution: WorkflowExecution }) {
  return (
    <TableRow>
      <TableCell className="font-medium">{execution.workflow_name || execution.workflow_id || execution.id}</TableCell>
      <TableCell><Badge variant="outline">{execution.status || execution.mode || '-'}</Badge></TableCell>
      <TableCell>{execution.started_at || '-'}</TableCell>
    </TableRow>
  );
}

function statusLabel(status?: WorkflowRuntimeStatus) {
  if (status === WorkflowRuntimeStatus.WORKFLOW_RUNTIME_AVAILABLE) return '可用';
  if (status === WorkflowRuntimeStatus.WORKFLOW_RUNTIME_DEGRADED) return '降级';
  if (status === WorkflowRuntimeStatus.WORKFLOW_RUNTIME_UNCONFIGURED) return '未配置';
  if (status === WorkflowRuntimeStatus.WORKFLOW_RUNTIME_UNAVAILABLE) return '不可用';
  return '未知';
}

function statusBadgeVariant(status?: WorkflowRuntimeStatus) {
  if (status === WorkflowRuntimeStatus.WORKFLOW_RUNTIME_AVAILABLE) return 'default' as const;
  if (status === WorkflowRuntimeStatus.WORKFLOW_RUNTIME_UNCONFIGURED) return 'secondary' as const;
  return 'outline' as const;
}

function emptyWorkflowText(loading: boolean, apiConfigured: boolean) {
  if (loading) return '加载中…';
  return apiConfigured ? '暂无流程定义。' : '暂无可展示流程；配置 n8n API key 后自动展示。';
}

function emptyExecutionText(loading: boolean, apiConfigured: boolean) {
  if (loading) return '加载中…';
  return apiConfigured ? '暂无最近执行；平台不做历史回放。' : '暂无执行投影；平台不做历史回放。';
}
