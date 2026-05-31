import { Tabs, TabsContent, TabsList, TabsTrigger } from '@byte-v-forge/common-ui';
import type { WorkflowDefinition, WorkflowExecution, WorkflowRuntimePageInfo, WorkflowRunProjection } from '@byte-v-forge/common-ui';
import { ExecutionCards, RunCards, WorkflowCards } from './workflow-list-cards';

export type WorkflowSelectionKind = 'runs' | 'workflows' | 'executions';

export function WorkflowSidebar({
  tab, workflows, runs, executions, runsPageInfo, executionsPageInfo, selectedID, loading, loadingMoreRuns, loadingMoreExecutions,
  apiConfigured, onTabChange, onRunSelect, onWorkflowSelect, onExecutionSelect, onNextRuns, onNextExecutions
}: {
  tab: WorkflowSelectionKind;
  workflows: WorkflowDefinition[];
  runs: WorkflowRunProjection[];
  executions: WorkflowExecution[];
  runsPageInfo?: WorkflowRuntimePageInfo;
  executionsPageInfo?: WorkflowRuntimePageInfo;
  selectedID?: string;
  loading: boolean;
  loadingMoreRuns: boolean;
  loadingMoreExecutions: boolean;
  apiConfigured: boolean;
  onTabChange: (tab: WorkflowSelectionKind) => void;
  onRunSelect: (run: WorkflowRunProjection) => void;
  onWorkflowSelect: (workflow: WorkflowDefinition) => void;
  onExecutionSelect: (execution: WorkflowExecution) => void;
  onNextRuns: () => void;
  onNextExecutions: () => void;
}) {
  return (
    <aside className="workflowSidebar">
      <Tabs value={tab} onValueChange={(next) => onTabChange(next as WorkflowSelectionKind)} className="workflowSidebarTabs">
        <TabsList className="workflowSidebarTabsList">
          <TabsTrigger value="runs">实时 {runs.length}</TabsTrigger>
          <TabsTrigger value="workflows">定义 {workflows.length}</TabsTrigger>
          <TabsTrigger value="executions">执行 {executions.length}</TabsTrigger>
        </TabsList>
        <TabsContent value="runs" className="workflowSidebarContent">
          <RunCards runs={runs} pageInfo={runsPageInfo} loadingMore={loadingMoreRuns} selectedID={tab === 'runs' ? selectedID : ''} onSelect={onRunSelect} onNext={onNextRuns} />
        </TabsContent>
        <TabsContent value="workflows" className="workflowSidebarContent">
          <WorkflowCards workflows={workflows} selectedID={tab === 'workflows' ? selectedID : ''} loading={loading} apiConfigured={apiConfigured} onSelect={onWorkflowSelect} />
        </TabsContent>
        <TabsContent value="executions" className="workflowSidebarContent">
          <ExecutionCards executions={executions} pageInfo={executionsPageInfo} loadingMore={loadingMoreExecutions} selectedID={tab === 'executions' ? selectedID : ''} loading={loading} apiConfigured={apiConfigured} onSelect={onExecutionSelect} onNext={onNextExecutions} />
        </TabsContent>
      </Tabs>
    </aside>
  );
}
