# AGENTS.md

本文件适用于 `workflow-runtime` 子仓。

## 边界

- 承载 Workflow/n8n 平台运行时控制面、状态投影 API、dashboard 远程模块和 n8n adapter。
- n8n editor 只作为管理员编排入口；业务前端不得 iframe/跳转替代平台原生状态页。
- 不承载 GPT、邮箱、SMS、代理等业务状态机；跨服务协作通过 proto/API/事件完成。
- 公共契约来自 `common-lib/proto/byte/v/forge/contracts/...`，不得复制公共模型。

## 实现

- 后端使用 Go，配置来自环境变量，secret 只通过 Kubernetes Secret 注入。
- 前端只发布 module federation dashboard module；shell 装载由 `deploy/frontend-modules.json` 声明。
- 不做历史回放；前端展示当前投影/查询结果，运行态后续通过事件/投影扩展。
