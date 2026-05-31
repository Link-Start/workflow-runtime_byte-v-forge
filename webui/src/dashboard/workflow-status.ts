type StatusValue = string | number | undefined | null;

export function runtimeStatusLabel(status?: StatusValue) {
  const value = String(status || '');
  if (value === 'WORKFLOW_RUNTIME_AVAILABLE') return '可用';
  if (value === 'WORKFLOW_RUNTIME_DEGRADED') return '降级';
  if (value === 'WORKFLOW_RUNTIME_UNCONFIGURED') return '未配置';
  if (value === 'WORKFLOW_RUNTIME_UNAVAILABLE') return '不可用';
  return '未知';
}

export function runStatusLabel(status?: StatusValue) {
  const value = String(status || '');
  if (!value || value === 'WORKFLOW_RUN_STATUS_UNSPECIFIED') return '未知';
  const labels: Record<string, string> = {
    WORKFLOW_RUN_PENDING: '排队',
    WORKFLOW_RUN_RUNNING: '运行中',
    WORKFLOW_RUN_WAITING: '等待',
    WORKFLOW_RUN_SUCCEEDED: '成功',
    WORKFLOW_RUN_FAILED: '失败',
    WORKFLOW_RUN_CANCELED: '取消',
    WORKFLOW_RUN_SKIPPED: '跳过',
    success: '成功',
    error: '失败',
    running: '运行中',
    waiting: '等待',
    canceled: '取消',
    skipped: '未执行',
    new: '排队',
    active: '启用',
    inactive: '停用'
  };
  return labels[value] || value.replace(/^WORKFLOW_RUN_/, '').toLowerCase();
}

export function statusTone(status?: StatusValue) {
  const value = String(status || '').toLowerCase();
  if (value.includes('available') || value.includes('succeeded') || value === 'success' || value === 'active') return 'good';
  if (value.includes('running') || value.includes('waiting') || value.includes('pending') || value === 'new') return 'active';
  if (value.includes('failed') || value.includes('unavailable') || value.includes('error') || value.includes('crashed')) return 'bad';
  return 'neutral';
}

export function badgeVariant(status?: StatusValue) {
  const tone = statusTone(status);
  if (tone === 'good') return 'default' as const;
  if (tone === 'bad') return 'destructive' as const;
  if (tone === 'active') return 'secondary' as const;
  return 'outline' as const;
}
