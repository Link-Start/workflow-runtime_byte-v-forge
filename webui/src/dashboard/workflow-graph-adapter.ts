import { formatUnix } from '@byte-v-forge/common-ui';
import type { WorkflowDefinition, WorkflowExecution, WorkflowRunProjection } from '@byte-v-forge/common-ui';
import type { WorkflowGraphEdge, WorkflowGraphNode } from '@byte-v-forge/common-ui/workflow-graph';
import type { WorkflowGraphNode as ProtoGraphNode } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/workflow/v1/workflow';

export function workflowDefinitionGraph(workflow?: WorkflowDefinition | null): { nodes: WorkflowGraphNode[]; edges: WorkflowGraphEdge[] } {
  if (!workflow) return { nodes: [], edges: [] };
  return { nodes: workflow.graph_nodes.map((node, index) => graphNode(node, index)), edges: workflow.graph_edges.map(graphEdge) };
}

export function workflowExecutionGraph(execution: WorkflowExecution, workflows: WorkflowDefinition[]) {
  if (execution.graph_nodes.length) return { nodes: execution.graph_nodes.map((node, index) => graphNode(node, index)), edges: execution.graph_edges.map(graphEdge) };
  const workflow = workflows.find((item) => item.id === execution.workflow_id);
  const graph = workflowDefinitionGraph(workflow);
  return graph;
}

export function workflowRunGraph(run: WorkflowRunProjection, workflows: WorkflowDefinition[]) {
  const base = workflowDefinitionGraph(workflows.find((item) => item.id === run.workflow_id));
  const overlay = new Map(run.graph_nodes.map((node, index) => [node.id || node.name, graphNode(node, index)]));
  if (!base.nodes.length) return { nodes: [...overlay.values()], edges: run.graph_edges.map(graphEdge) };
  return {
    nodes: base.nodes.map((node) => ({ ...node, ...(overlay.get(node.id) || overlay.get(node.label) || {}), position: node.position })),
    edges: base.edges.length ? base.edges.map((edge) => ({ ...edge, status: edge.status || edgeStatus(edge, overlay) })) : run.graph_edges.map(graphEdge)
  };
}

function graphNode(node: ProtoGraphNode, index: number): WorkflowGraphNode {
  const position = nodePosition(node);
  return {
    id: node.id,
    label: node.name || node.id,
    subtitle: nodeSubtitle(node),
    kind: node.kind,
    status: node.status || '',
    startedAt: node.started_at_unix ? formatUnix(node.started_at_unix) : undefined,
    duration: durationText(node.duration_ms),
    message: node.iterations > 1 ? `${node.iterations} iterations` : undefined,
    error: node.error_message || undefined,
    order: index,
    position
  };
}

function graphEdge(edge: { id: string; source: string; target: string; label: string; status: string }): WorkflowGraphEdge {
  return { id: edge.id, source: edge.source, target: edge.target, label: edge.label, status: edge.status };
}

function edgeStatus(edge: WorkflowGraphEdge, overlay: Map<string, WorkflowGraphNode>) {
  const target = overlay.get(edge.target);
  return target?.status || '';
}

function nodeSubtitle(node: ProtoGraphNode) {
  return [node.kind, node.type_version ? `v${node.type_version}` : '', node.disabled ? 'disabled' : ''].filter(Boolean).join(' · ');
}

function durationText(ms: number) {
  if (!ms) return undefined;
  if (ms < 1000) return `${ms}ms`;
  return `${Math.round(ms / 100) / 10}s`;
}

function nodePosition(node: { x?: number; y?: number }) {
  if (!Number.isFinite(node.x) && !Number.isFinite(node.y)) return undefined;
  return {
    x: Number.isFinite(node.x) ? Number(node.x) : 0,
    y: Number.isFinite(node.y) ? Number(node.y) : 0
  };
}
