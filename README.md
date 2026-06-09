# workflow-runtime

`workflow-runtime` 是工作流运行时控制面，负责 n8n 编排状态接入、平台 run/step 投影、状态事件流和 Workflow dashboard 远程模块。

## 核心能力

- 通过 n8n Public API 汇总引擎状态、workflow 定义和最近 execution。
- 将 n8n 节点、连线和位置投影为平台流程图，保持与管理员 editor 一致。
- 接收 n8n HTTP Request 节点上报的 run/step 状态，维护当前运行投影。
- 通过 SSE/HotStream 向前端推送工作流状态变化。
- 提供 Workflow dashboard 远程模块；n8n editor 仅作为管理员编排入口。

## 使用方式

业务前端查询平台原生状态页，不直接 iframe 或跳转到 n8n editor。业务服务通过 API、事件或 workflow 节点上报协作；GPT、Mailbox、SMS、Proxy 等业务状态机留在各自服务内。

## 入口

- 服务入口：`cmd/workflow-runtime`
- Dashboard 模块：`webui/`
- 状态 API：`/api/workflow-runtime/*`
- 步骤上报：`POST /api/workflow-runtime/runs/steps`
- 状态流：`GET /api/workflow-runtime/streams/state`

## 常用检查

```sh
(cd webui && npm run lint)
git diff --check
```
