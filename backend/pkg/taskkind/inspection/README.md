# taskkind: inspection

## 类型

**巡检任务** — 包含检查项（check items）和分支（branches）的复合任务。每个检查项下有多个分支选项，完成巡检时选择结果，异常分支可自动生成跟进子任务。

## 数据模型

```
巡检任务 (Task, kind=inspection)
  └── 检查项 (CheckItem)
        ├── 分支: 正常 (不生成跟进)
        ├── 分支: 异常 (create_todo: true → 自动生成跟进任务)
        └── 分支: ... (更多自定义分支)
```

| 表 | 说明 |
|---|------|
| `check_items` | 检查项，关联到巡检任务 |
| `check_item_branches` | 检查项的分支选项 |
| `inspection_results` | 巡检完成时的审计记录 |

## 适用场景

| 场景 | 示例 |
|------|------|
| 设备巡检 | 检查机房温度、服务器状态 → 异常自动报修 |
| 卫生检查 | 检查厨房/卫生间 → 不合格生成打扫任务 |
| 安全巡检 | 检查消防设施 → 缺失生成采购工单 |
| 质量检查 | 生产线质检 → 不合格生成返工任务 |

## 前端交互

巡检待办与其他任务不同：

1. 待办卡片显示"巡检"徽章 + 检查项摘要
2. 完成巡检时打开检查项弹窗，逐项选择结果
3. 分支选择基于 `create_todo` 标记互斥：
   - 无 `create_todo` 的分支（如"正常"）与有 `create_todo` 的分支互斥
   - 多个有 `create_todo` 的分支可同时选中
   - 点击已选中的分支可取消选择
4. 完成后，异常分支自动生成跟进待办

## 模板配置

### 必填字段

```yaml
kind: inspection                 # ← 必须为 inspection
```

### extra_schema 结构

```yaml
extra_schema:
  check_items:
    - name: "检查项名称"          # 必填
      branches:
        - name: "正常"           # 无 create_todo → 正常结果，无跟进
        - name: "异常"           # create_todo: true → 异常结果
          create_todo: true
          branch_task:           # 生成的跟进任务配置
            task:
              name: "{{.area_name}} - 设备异常处理"
              schedule_type: once
              kind: simple       # 跟进任务类型（通常为 simple）
              group_id: ""       # 可选：指定小组
              location_id: ""    # 可选：指定地点
            extra: null          # 跟进任务的 extra 数据
```

### 完整示例

```yaml
version: 1
provider:
  code: builtin
  name: "官方模板"
templates:
  - code: daily_inspection
    name: "每日巡检"
    description: "标准每日设备巡检模板"
    kind: inspection
    icon: "🔍"
    parameters:
      - key: area_name
        label: "区域名称"
        type: string
        required: true
        placeholder: "例如：A区"
    task_defaults:
      name: "{{.area_name}} - 每日巡检"
      schedule_type: daily
      schedule_data:
        time: "09:00"
      enabled: true
    extra_schema:
      check_items:
        - name: "设备运行状态"
          branches:
            - name: "正常"
            - name: "异常"
              create_todo: true
              branch_task:
                task:
                  name: "{{.area_name}} - 设备异常处理"
                  schedule_type: once
                  kind: simple
                extra: null
        - name: "环境卫生"
          branches:
            - name: "合格"
            - name: "不合格"
              create_todo: true
              branch_task:
                task:
                  name: "{{.area_name}} - 卫生整改"
                  schedule_type: once
                  kind: simple
                extra: null
```

### branch_task 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `task.name` | string | 跟进任务名称，支持 `{{.param}}` 模板 |
| `task.schedule_type` | string | `once` / `daily` / `weekly` 等 |
| `task.kind` | string | 跟进任务类型，通常为 `simple` |
| `task.group_id` | string | 可选，指派到特定小组 |
| `task.location_id` | string | 可选，关联特定地点 |
| `extra` | object | 跟进任务的 extra 数据 |