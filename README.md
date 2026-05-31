# workflow-runtime

`workflow-runtime` 是 Byte V Forge 的工作流运行时控制面，负责 n8n 流程定义读取、当前运行投影、状态事件流和 Workflow dashboard 远程模块。

## 当前职责

- 通过 n8n Public API 汇总引擎状态、流程定义和最近执行。
- 将 n8n `nodes/connections/position` 投影为平台流程图，保持和 n8n editor 图结构一致。
- 承载当前 run/step 投影；n8n workflow 在关键节点通过 HTTP Request 节点上报状态。
- 通过 `/api/workflow-runtime/streams/state` 向前端推送 HotStream/SSE 变更事件。
- n8n editor 仅作为管理员编排入口；业务前端使用平台原生页面查询/订阅状态。

## API

- `GET /api/workflow-runtime/summary`：当前流程定义、最近 n8n execution、平台 run 投影。
- `GET /api/workflow-runtime/streams/state`：HotStream/SSE，事件类型 `workflow-runtime.summary.updated`。
- `POST /api/workflow-runtime/runs/steps`：工作流节点状态上报。

n8n HTTP Request 节点上报体示例：

```json
{
  "run_id": "{{$json.run_id}}",
  "workflow_id": "{{$workflow.id}}",
  "workflow_name": "{{$workflow.name}}",
  "execution_id": "{{$execution.id}}",
  "node_name": "当前节点名",
  "status": "WORKFLOW_RUN_RUNNING",
  "occurred_at_unix": "{{Math.floor(Date.now() / 1000)}}"
}
```

节点完成后将 `status` 改为 `WORKFLOW_RUN_SUCCEEDED`；失败分支使用 `WORKFLOW_RUN_FAILED` 并传 `error_message`。

## 运行配置

- `WORKFLOW_RUNTIME_HTTP_ADDR`：HTTP 监听地址，默认 `:8080`。
- `WORKFLOW_RUNTIME_DASHBOARD_STATIC_DIR`：远程前端静态目录，默认 `/app/dashboard/workflow-runtime`。
- `N8N_INTERNAL_URL`：集群内 n8n main 地址。
- `N8N_PUBLIC_URL`：管理员 editor 公网入口。
- `N8N_API_KEY`：n8n public API key。
