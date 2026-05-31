import { useEffect, useMemo, useState } from 'react';
import { Maximize2, Minimize2 } from 'lucide-react';
import { Badge, Button } from '@byte-v-forge/common-ui';
import { WorkflowGraph, type WorkflowGraphNode } from '@byte-v-forge/common-ui/workflow-graph';
import type { WorkflowDefinition, WorkflowExecution, WorkflowRunProjection } from '@byte-v-forge/common-ui';
import { workflowDefinitionGraph, workflowExecutionGraph, workflowRunGraph } from './workflow-graph-adapter';
import { badgeVariant, runStatusLabel } from './workflow-status';

type GraphScope = {
  workflow?: WorkflowDefinition | null;
  execution?: WorkflowExecution | null;
  run?: WorkflowRunProjection | null;
  workflows: WorkflowDefinition[];
};

export function WorkflowGraphPanel({ workflow, execution, run, workflows }: GraphScope) {
  const [selectedNode, setSelectedNode] = useState<WorkflowGraphNode | null>(null);
  const [fullscreen, setFullscreen] = useState(false);
  const graph = useMemo(() => run ? workflowRunGraph(run, workflows) : execution ? workflowExecutionGraph(execution, workflows) : workflowDefinitionGraph(workflow), [workflow, execution, run, workflows]);
  const title = run ? run.workflow_name || run.workflow_id || run.run_id : execution ? execution.workflow_name || execution.workflow_id || execution.id : workflow?.name || '流程图';
  const meta = run ? `run #${run.run_id}` : execution ? `execution #${execution.id}` : workflow?.id || '选择流程后展示节点图';
  const status = run?.status || execution?.status || (workflow?.active ? 'active' : 'inactive');
  const selected = selectedNode && graph.nodes.some((node) => node.id === selectedNode.id) ? selectedNode : null;
  const fitView = graph.nodes.length > 0 && graph.nodes.length <= 18;
  const showStatus = Boolean(run || execution);
  useEffect(() => setSelectedNode(null), [title, meta]);
  useEffect(() => {
    if (!fullscreen) return undefined;
    const close = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setFullscreen(false);
    };
    window.addEventListener('keydown', close);
    return () => window.removeEventListener('keydown', close);
  }, [fullscreen]);
  return (
    <section className={`workflowGraphPanel ${fullscreen ? 'fullscreen' : ''}`}>
      <div className="workflowGraphHeader">
        <div><strong>{title}</strong><span>{meta}</span></div>
        <div className="workflowGraphHeaderMeta">
          {showStatus && <Badge variant={badgeVariant(status)}>{runStatusLabel(status)}</Badge>}
          <span>{graph.nodes.length} nodes</span><span>{graph.edges.length} edges</span>
          <Button variant="ghost" size="icon-sm" aria-label={fullscreen ? '退出全屏' : '全屏'} onClick={() => setFullscreen((value) => !value)}>
            {fullscreen ? <Minimize2 /> : <Maximize2 />}
          </Button>
        </div>
      </div>
      <div className={`workflowGraphBody ${selected ? 'hasInspector' : ''}`}>
        <WorkflowGraph nodes={graph.nodes} edges={graph.edges} selectedNodeId={selected?.id} fitView={fitView} initialZoom={0.72} showNodeStatus={showStatus} showMiniMap={graph.nodes.length > 10} emptyText="当前流程没有可展示节点" className="workflowRuntimeGraph" onNodeSelect={setSelectedNode} onPaneClick={() => setSelectedNode(null)} />
        {selected && <NodeInspector node={selected} showStatus={showStatus} />}
      </div>
    </section>
  );
}

function NodeInspector({ node, showStatus }: { node: WorkflowGraphNode; showStatus: boolean }) {
  return (
    <aside className="workflowNodeInspector">
      <div className="workflowInspectorTitle"><span>{node.label}</span>{showStatus && <Badge variant={badgeVariant(node.status)}>{runStatusLabel(node.status)}</Badge>}</div>
      <dl>
        <dt>类型</dt><dd>{node.kind || node.subtitle || '-'}</dd>
        <dt>开始</dt><dd>{node.startedAt || '-'}</dd>
        <dt>耗时</dt><dd>{node.duration || '-'}</dd>
        <dt>信息</dt><dd>{node.error || node.message || '-'}</dd>
      </dl>
    </aside>
  );
}
