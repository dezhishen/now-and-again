# taskkind: simple

## 类型

**普通任务** — 不包含任何额外数据结构，仅由任务本身的字段（名称、调度类型、调度数据）组成。

## 适用场景

| 场景 | 示例 |
|------|------|
| 每日打卡 | 晨会签到、每日阅读 |
| 周期提醒 | 每周换床单、每月大扫除 |
| 一次性任务 | 取快递、预约体检 |
| 简单待办 | 任何无需额外配置的待办事项 |

## 实现

`simple` 是任务类型插件的**最小实现**。所有生命周期方法（`SaveExtra`、`UpdateExtra`、`DeleteExtra`、`OnComplete`、`GetExtra`）均为空操作，不存储任何额外数据，也没有数据库表。

前端仅提供基础的任务卡片、完成/跳过/备注按钮，无任何类型特定的交互。

## 模板配置

### 参数（可选）

```yaml
parameters:
  - key: task_name
    label: "任务名称"
    type: string
    required: true
    placeholder: "例如：晨会"
```

### task_defaults

```yaml
task_defaults:
  name: "每日{{.task_name}}"       # 支持 Go template 语法引用参数
  schedule_type: daily             # once / daily / weekly / monthly / interval
  schedule_data:                   # 不同 schedule_type 对应不同格式
    time: "08:00"                  # daily: 每天执行时间
  enabled: true
```

### 完整示例

```yaml
version: 1
provider:
  code: builtin
  name: "自定义模板"
templates:
  - code: my_daily_check
    name: "每日打卡"
    description: "简单的每日打卡任务"
    kind: simple                   # ← 必须为 simple
    icon: "✅"
    parameters:
      - key: check_item
        label: "打卡项目"
        type: string
        required: true
        placeholder: "例如：晨会"
    task_defaults:
      name: "每日{{.check_item}}"
      schedule_type: daily
      schedule_data:
        time: "08:00"
      enabled: true
```
