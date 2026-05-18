# 贡献指南

## 边界

本仓只承载通用 workflow/worker 运行时基础设施。

以下内容不进入本仓：

- 业务 workflow；
- 业务 activity；
- 业务状态机；
- 业务私有 proto；
- 私有契约仓或私有 SDK 仓依赖。

## 开发流程

1. 保持 API 通用，不绑定具体业务。
2. 通过构造函数或参数注入外部依赖。
3. Workflow 代码必须保持确定性，side effect 放到 activity。
4. 修改后运行验证命令。

## 验证

```sh
go test ./...
go test -race ./...
go vet ./...
```

## 契约

跨仓共享模型来自公开 `contracts/workflowruntime`，Go 类型来自 `contracts-go`。
