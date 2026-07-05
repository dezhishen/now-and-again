# taskkind: chain

## 类型

**任务链** — 将多个不同类型的任务串联为有序流水线。每一步可使用不同的任务类型（simple/inspection/...），前一步完成后自动激活下一步。

## 数据模型

```
任务链 (Task, kind=chain, OwnerKind=chain)
  ├── 步骤1 (Task, kind=simple,  OwnerKind=chain)
  │     └── 子级可能由步骤 handler 自动生成（如 inspection 分支任务）
  ├── 步骤2 (Task, kind=inspection, OwnerKind=chain)
  │     └── 巡检子任务 (Task, kind=simple, OwnerKind=inspection)
  └── 步骤3 (Task, kind=simple, OwnerKind=chain)

chain_steps 表记录步骤定义（TaskID=root, SortOrder, Name, Kind, ChildTaskID）
```

| 表 | 说明 |
|---|------|
| `chain_steps` | 步骤定义，关联到链根任务。`Kind`=步骤真实类型，`ChildTaskID`=实际子任务 |

## 调度机制

链式推进基于 `OwnerKind` 委托模式：

```
Root(chain) todo 完成
  → CompleteTodo 调度 OwnerKind="chain" → chain.OnTodo
    → 查找 sort_order=0 → CreateTodo(步骤1)

步骤1(inspection) todo 完成
  → CompleteTodo 调度 OwnerKind="chain" → chain.OnTodo
    → 1. 委托: LookupHandler("inspection").OnTodo()  → 巡检逻辑
    → 2. 查找 sort_order+1 → CreateTodo(步骤2)

步骤2(simple) todo 完成
  → CompleteTodo 调度 OwnerKind="chain" → chain.OnTodo
    → 1. 委托: LookupHandler("simple").OnTodo()  → (空操作)
    → 2. 查找 sort_order+1 → 无更多步骤 → 链结束
```

## 关键设计

### OwnerKind

- 链根任务：`SaveExtra` 将 `root.OwnerKind` 更新为 `"chain"`（DB 默认 `"simple"`）
- 每步子任务：`Kind = step.Kind`（真实类型），`OwnerKind = "chain"`
- 步骤内的嵌套子任务（如 inspection 分支）：由对应 handler 设置 `OwnerKind = "inspection"`

### OnTodo 委托

`chain.OnTodo` 是复合 handler 的 **入口** ，负责两件事：
1. 委托真实 handler：`storage.LookupHandler(todo.Task.Kind).OnTodo(...)`
2. 链推进：`CreateTodo(nextStep.ChildTaskID)`

非复合 handler（simple/inspection）的 `OnTodo` **不需要** 做任何委托——调度已在 `CompleteTodo` 中通过 `OwnerKind` 优先完成。

### SaveExtra 透传

步骤的 `extra` 字段支持 JSON 透传：chain 创建子任务时调用 `CreateNoRootTask(child, s.Extra)`，这会自动触发子任务 kind 的 `SaveExtra`，因此 inspection 步骤的 `check_items` 等数据能正确传递。

## 适用场景

| 场景 | 示例 |
|------|------|
| 审批流 | 提交申请(simple) → 主管审批(inspection) → 终审确认(simple) |
| 多阶段巡检 | 预热检查(inspection) → 运行测试(simple) → 关机检查(inspection) |
| 复合流程 | 备料(simple) → 质检(inspection) → 打包(simple) → 终检(inspection) |

## 前端交互

- 任务卡片：橙色角标 + 步骤摘要（`A→B→C` 或 `A→B→C→等5项`）
- 待办按钮：`完成/备注/跳过/中断`
- 创建表单：子任务编辑器（SubTaskEditor），每步可选择类型并配置对应参数

## 后端 extra 格式

```json
{
  "steps": [
    { "name": "第一步", "kind": "simple" },
    { "name": "第二步", "kind": "inspection", "extra": { "check_items": [...] } },
    { "name": "第三步", "kind": "simple" }
  ]
}
```
