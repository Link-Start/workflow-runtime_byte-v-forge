# workflow-runtime

`workflow-runtime` 是注册系统的 Temporal/worker 运行时基础设施仓。

## 职责

- 提供通用 Temporal/worker 运行时能力。
- 面向开源基础设施场景维护。
- 跨仓共享的 workflow runtime 模型来自公开 `contracts/workflowruntime` proto。
- 各业务仓声明自己的 workflow、activity、worker、内部 proto 和业务数据。
- Temporal history 记录执行过程；业务事实由业务服务自己的数据库、事件和读模型维护。
- 业务开放接口由各业务仓用自己的 gRPC 服务承载。

## 提供能力

- Temporal client 配置和创建。
- Worker spec 校验、注册和按 `context.Context` 生命周期运行。
- Task queue 命名校验。
- 默认 activity timeout 和 retry policy。
- `contracts/workflowruntime` proto 与 Temporal Go SDK policy 的转换。
- 环境变量配置加载。

## 配置

默认环境变量：

- `TEMPORAL_ADDRESS`
- `TEMPORAL_NAMESPACE`
- `TEMPORAL_TASK_QUEUE`
- `TEMPORAL_IDENTITY`

`TEMPORAL_ADDRESS`、`TEMPORAL_NAMESPACE`、`TEMPORAL_TASK_QUEUE` 为必填配置。

## 业务仓使用方式

业务仓应声明自己的 task queue、workflow 名称和 activity 名称，并通过本仓启动 worker：

```go
cfg, err := workflowruntime.LoadConfigFromEnv(nil)
if err != nil {
    return err
}

temporalClient, err := workflowruntime.Dial(cfg)
if err != nil {
    return err
}
defer temporalClient.Close()

return workflowruntime.RunWorker(ctx, temporalClient, workflowruntime.WorkerSpec{
    TaskQueue: cfg.TaskQueue,
    Workflows: []workflowruntime.WorkflowDefinition{
        {Name: "example-registration.workflow.v1", Definition: RegistrationWorkflow},
    },
    Activities: []workflowruntime.ActivityDefinition{
        {Name: "example-registration.prepare-account.v1", Definition: activities.PrepareAccount},
    },
})
```

## 验证

```sh
go test ./...
go test -race ./...
go vet ./...
```

## 生成物

本仓当前没有 IDL 生成流程。未来如果加入 proto 或其他生成源，生成输出按仓库生成规则处理。

## 贡献与安全

- 贡献规则见 `CONTRIBUTING.md`。
- 安全报告规则见 `SECURITY.md`。
- 本仓使用 Apache-2.0 许可证。
