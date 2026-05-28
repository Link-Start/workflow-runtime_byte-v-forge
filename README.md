# workflow-runtime

`workflow-runtime` 是 Byte V Forge 的工作流运行时控制面，负责 n8n 运行状态投影、Workflow dashboard 远程模块和 `/api/workflow-runtime/*` API。

## 当前职责

- 对接 n8n 内部 API，汇总引擎、API key、流程定义和最近执行状态。
- 暴露 dashboard API：`/api/workflow-runtime/summary`。
- 暴露 dashboard Module Federation remote：`/mf/workflow-runtime/remoteEntry.js`。
- n8n editor 仅作为管理员编排入口；业务前端使用平台原生页面查询/订阅状态。

## 运行配置

- `WORKFLOW_RUNTIME_HTTP_ADDR`：HTTP 监听地址，默认 `:8080`。
- `WORKFLOW_RUNTIME_DASHBOARD_STATIC_DIR`：远程前端静态目录，默认 `/app/dashboard/workflow-runtime`。
- `N8N_INTERNAL_URL`：集群内 n8n main 地址。
- `N8N_PUBLIC_URL`：管理员 editor 公网入口。
- `N8N_API_KEY`：n8n public API key。
