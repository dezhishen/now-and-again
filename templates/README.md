# 任务模板说明

## 如何添加模板

1. 在 `templates/` 目录下创建或编辑 `.yaml` 文件
2. 推送到 GitHub `main` 分支
3. 系统自动从 `raw.githubusercontent.com` 拉取同步

> 管理员也可以在 `${DATA_DIR}/templates/` 下放置自定义 YAML 文件，通过管理面板手动刷新 `builtin` provider。

## YAML 结构

```yaml
version: 1                         # 固定为 1
provider:
  code: http                       # 固定为 http（GitHub 订阅）
  name: "官方模板"
  description: "模板描述"
templates:
  - code: template_code            # 唯一标识，建议英文小写+下划线，如 weekly_cleaning
    name: "模板名称"                # 显示名称
    description: "模板描述"          # 可选
    kind: simple | inspection      # 任务类型
    icon: "🧹"                     # 可选，emoji 图标
    sort_order: 1                  # 排序，数字越小越靠前
    enabled: true                  # 是否启用
    version: "1.0.0"              # 可选，版本号
    parameters: []                 # 参数列表，可选
    task_defaults: {}              # 任务默认值（支持 Go template）
    extra_schema: {}               # 类型特定配置（inspection 为 check_items）
```

## 参数类型

| 类型 | 说明 | 渲染方式 | 示例 default |
|------|------|---------|-------------|
| `string` | 文本输入 | 普通文本框 | `"客厅"` |
| `int` | 整数 | 数字输入 | `1` |
| `float` | 小数 | 数字输入 | `1.5` |
| `bool` | 布尔 | 复选框 | `true` |
| `time` | 时间 | 时间选择器 | — |
| `select` | 下拉选择 | 静态选项下拉 | `"option1"` |
| `array` | 数组 | 多行文本（每行一个值） | — |
| `location` | 地点 | 从系统地点列表选择 | 地点 ID |
| `group` | 小组 | 从系统小组列表选择 | 小组 ID |
| `schedule` | 调度 | 下拉 + 子字段展开 | `"daily"` |

### 参数定义

```yaml
parameters:
  - key: area_name                  # 唯一标识，模板中通过 {{.area_name}} 引用
    label: "区域名称"                # 显示标签
    type: string                    # 参数类型，见上表
    description: "需要巡检的区域"     # 可选，提示文字
    required: true                  # 是否必填
    default: "客厅"                  # 可选，默认值
    placeholder: "例如：厨房"        # 可选，占位提示
    options:                        # 仅 select 类型需要
      - label: "选项A"
        value: "a"
```

### array 类型

用户每行输入一个值，存储为 JSON 数组，Go template 可用 `range` 遍历：

```yaml
parameters:
  - key: rooms
    label: "房间列表"
    type: array
    required: true
    placeholder: "主卧\n次卧\n儿童房"

task_defaults:
  name: "{{range $i, $r := .rooms}}{{if $i}}、{{end}}{{$r}}{{end}} - 大扫除"
  # 渲染结果: "主卧、次卧、儿童房 - 大扫除"
```

### schedule 类型子字段

选择调度类型后自动展开对应子字段：

| 调度类型 | 子字段 | 模板引用 |
|---------|--------|---------|
| `once` | 无 | `{{.key}}` |
| `daily` | 时间 | `{{.key}}` + `{{.key_time}}` |
| `weekly` | 星期 + 时间 | `{{.key}}` + `{{.key_day}}` + `{{.key_time}}` |
| `monthly` | 日期 + 时间 | `{{.key}}` + `{{.key_day}}` + `{{.key_time}}` |
| `interval` | 小时间隔 | `{{.key}}` + `{{.key_hours}}` |

## task_defaults 写法

支持 Go template 语法引用参数：

```yaml
task_defaults:
  name: "{{.area_name}} - 每日巡检"    # 使用参数渲染任务名
  schedule_type: daily                 # once / daily / weekly / monthly / interval
  schedule_data:                       # 与 schedule_type 对应
    time: "09:00"                      # daily: 时间
    # day: 1                           # weekly: 0-6 (0=周日), monthly: 1-31
    # hours: 72                        # interval: 小时间隔
  enabled: true
  # group_id: ""                       # 可选，指定小组
  # location_id: ""                    # 可选，指定地点
```

## 全局标准参数

**所有模板使用时自动追加以下三个参数**，无需在 YAML 中定义：

| key | 类型 | 标签 | 说明 |
|-----|------|------|------|
| `_schedule` | schedule | 调度策略 | 默认 daily，用户可按需调整 |
| `_location` | location | 执行地点 | 从系统地点列表选择 |
| `_group` | group | 指派小组 | 从系统小组列表选择 |

> 如果在 YAML 中显式定义了同名参数，则使用自定义的 label/required/default，不会重复添加。

## extra_schema（inspection 专属）

巡检类型需配置检查项和分支。支持两种写法：

### 静态写法

检查项固定不变：

```yaml
extra_schema:
  check_items:
    - name: "设备运行状态"
      branches:
        - name: "正常"
        - name: "异常"
          create_todo: true
          branch_task:
            task:
              name: "设备异常处理"
              schedule_type: once
              kind: simple
            extra: null
```

### 动态写法（使用 `|` 嵌入 Go template）

根据参数动态生成检查项，用 `|`（YAML 字面量块）嵌入 Go template 渲染的 YAML：

```yaml
extra_schema: |
  check_items:
  {{range $room := .rooms}}
    - name: "{{$room}}床品"
      branches:
        - name: "已换洗"
        - name: "未换洗"
          create_todo: true
          branch_task:
            task:
              name: "{{$room}} - 换洗床品"
              schedule_type: once
              kind: simple
            extra: null
  {{end}}
```

> `|` 后的内容先经过 Go template 渲染，再解析为 YAML，最后转 JSON 传给前端。
> `{{range}}` 等控制流语法只能在这种动态写法中使用。

### branch_task 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `task.name` | string | 跟进任务名称，支持 `{{.param}}` 模板 |
| `task.schedule_type` | string | `once` / `daily` / `weekly` 等 |
| `task.kind` | string | 跟进任务类型，通常为 `simple` |
| `task.group_id` | string | 可选，指派到特定小组 |
| `task.location_id` | string | 可选，关联特定地点 |
| `extra` | object | 跟进任务的 extra 数据，通常为 `null` |

## 完整示例

参见同目录下的：
- `daily_inspection.yaml` — 巡检类模板示例（静态 extra_schema）
- `household.yaml` — 家庭事务模板集（含 `|` 动态写法）
