# Now & Again — 系统架构

## 技术栈

| 层 | 技术 |
|---|------|
| 前端 | Vue 3 + Vite + TypeScript + Pinia + Tailwind CSS + vue-i18n (pnpm) |
| 后端 | Go 1.25 + Gin + GORM (github.com/glebarez/sqlite, 纯 Go 无 CGO) |
| CLI | Go + Cobra，通过 HTTP / API Key 调用后端 |
| 数据库 | SQLite |

## 分层架构

```
┌──────────────────────────────────────────┐
│  Frontend (Vue 3 SPA)    CLI (Cobra)     │
├──────────────────────────────────────────┤
│  Handler (Gin)    Middleware (JWT/Scope) │
├──────────────────────────────────────────┤
│  Service                                 │
├──────────────────────────────────────────┤
│  Repository (GORM)                       │
├──────────────────────────────────────────┤
│  SQLite                                  │
└──────────────────────────────────────────┘

backend/pkg/ — 公共类型、调度器、插件系统（taskkind / tasktemplate / locationkind / scheduler）
```

## 项目结构

```
backend/
  cmd/server/main.go      入口
  internal/
    handler/               HTTP 处理器
    service/               业务逻辑（user/family/task/todo/log/floorplan/ics/apikey/calendar/image/task_template）
    repository/            数据访问 + AutoMigrate
    middleware/             JWT / API Key / Scope 鉴权
    config/                配置
    logger/                日志
  pkg/
    types/                 共享 DTO + model→DTO 转换
    model/                 共享 GORM 模型（BaseModel, TaskModel 等）
    contracts/             API 接口定义
    scheduler/             任务调度引擎 (gocron) + 类型注册表
    taskkind/              任务类型插件 (simple, inspection)
    tasktemplate/          任务模板系统 (Provider 接口 + 内置 + HTTP 订阅)
    locationkind/          地点类型插件 (indoor)
    timeutil/              时间工具
    scopes/                权限范围
  internal/
    webui/                 嵌入前端 dist（Go embed）

cli/
  cmd/                     CLI 命令
  internal/client/         HTTP 客户端

frontend/
  src/
    views/                 页面组件
    components/tasks/      任务卡片 + 插件组件
    components/locations/  地点插件注册
    composables/           useTaskKinds / useLocationKinds / useToast
    stores/                Pinia (auth)
    i18n/                  中/英多语言
    router/                Vue Router
```

## 核心设计

### 插件系统

| 系统 | 后端包 | 前端 composable | 现有类型 |
|------|--------|-----------------|---------|
| 任务类型 | `pkg/taskkind/` | `useTaskKinds` | simple, inspection, chain |
| 任务模板 | `pkg/tasktemplate/` | — | builtin, http |
| 地点类型 | `pkg/locationkind/` | `useLocationKinds` | indoor |
| 调度类型 | `pkg/scheduler/` | — | once, daily, weekly, monthly, interval |

新增类型只需实现接口并注册（`init()` 自动注册），无需修改任何现有代码。

taskkind 各类型实现细节请参考：
- [simple](../../backend/pkg/taskkind/simple/README.md)
- [inspection](../../backend/pkg/taskkind/inspection/README.md)
- [chain](../../backend/pkg/taskkind/chain/README.md)

### 任务类型委托系统

任务类型间通过 `OwnerKind` 字段实现复合委托：

```
TaskModel.OwnerKind  →  记录"当前由哪个 handler 持有编排语义"
  - 用户直建：DB 默认 "simple"
  - OwnerKind 为空或默认值时：按 Task.Kind 分发
  - OwnerKind 为非默认值时：优先分发到 OwnerKind 对应 handler

CompleteTodo 调度优先级：
  kind = ResolveDispatchKind(todo.Task.OwnerKind, todo.Task.Kind)
  taskManager.Get(kind).OnTodo()

复合 handler（如 chain）内部委托：
  chain.OnTodo → storage.LookupHandler(realKind).OnTodo()  // 真实逻辑
              → CreateTodo(nextStep)                        // 链推进
```

**约束**：主流程（`todo_service.go`、`task_service.go`）永远不引用插件内部类型名和 kind 值，只通过 `TaskStorage` 接口和 `OwnerKind` 字段名交互。

### 任务模板插件生命周期

```
Provider 接口
  ├─ Code() / Name() / Description()  ← 标识和展示信息
  ├─ Sync(ctx, storage)               ← 解析数据源 → Upsert 到 task_templates 表
  ├─ LastSyncAt() / SyncStatus()      ← 同步状态查询
  └─ TemplateStorage（注入给 Provider 的方法集合）
       ├─ UpsertTemplate(tmpl)        ← 写入/更新模板（按 provider_code + template_code 去重）
       ├─ DeleteTemplate(code)        ← 删除过时模板（系统级清理）
       ├─ FindByProvider(code)        ← 查询某 Provider 的所有模板
       ├─ ListSubscriptions(code)     ← 查询订阅源（HTTP Provider 获取 URL 列表）
       └─ DB()                        ← 返回 *gorm.DB
```

- **内置 Provider**：Go `embed.FS` 打包 YAML，启动时自动 Sync 到系统级模板
- **HTTP Provider**：从订阅 URL 拉取远程 YAML，支持系统和家庭两级订阅
- 主流程通过 `Provider` 接口调用，不做类型断言，Provider 完全可插拔

```
taskkind.Handler（接口，Kind 由插件扩展）
  ├─ Kind() string                     ← 返回类型标识
  ├─ SaveExtra(storage, task, extra)   ← 新建时持久化插件特有数据
  ├─ UpdateExtra(storage, task, extra) ← 更新时持久化（nil extra = 清空）
  ├─ DeleteExtra(storage, task)        ← 删除时清理插件数据
  ├─ OnTodo(storage, todo, extra)      ← 待办状态变更时触发（extra 来自请求）
  └─ GetExtra(storage, task)           ← 读取插件数据供前端展示

taskkind.TaskStorage（注入给 Handler 的方法集合）
  ├─ FindTaskByID(taskID)               ← 按 ID 查任务
  ├─ FindTaskByParentId(parentID)       ← 按 parent_id 查子任务
  ├─ CreateNoRootTask(task, extra)      ← 创建子任务并触发其 SaveExtra
  ├─ UpdateNoRootTask(task, extra)      ← 更新子任务并触发 UpdateExtra
  ├─ UpdateTaskFields(task)             ← 仅更新字段，不触发 Handler
  ├─ DeleteNonRootTask(id)              ← 递归删除子树，触发 DeleteExtra
  ├─ CreateTodo(taskID, displaySummary) ← 创建待办，返回 *TodoModel
  └─ DB() *gorm.DB                      ← 获取 DB 实例做自定义查询
```

- 主流程（`TaskService`）只处理 root 节点的行记录
- 插件通过 `TaskStorage` 调用主流程注入的方法，实现递归嵌套
- 主流程不引用插件内部的 model/结构体

### 统一错误处理

所有 API 非 2xx 响应遵循统一格式 `{ success, error: { code, summary, details? } }`。
前端通过 `ApiRequestError` 类 + `ErrorDisplay` 组件按 ErrorCode 区分展示（400 琥珀色 / 500 红色）。

### 插件注册模式

- **taskkind**：`TaskManager` struct 管理 Handler 注册表
- **scheduler**：`Registry` struct（含 `sync.RWMutex`）管理调度类型注册表
- **locationkind**：包级 `map[string]Handler` + `Register()` 函数
- **GORM 迁移注册**：插件通过 `model.RegisterModel()` 在 `init()` 中注册模型，`AutoMigrate` 动态发现，无需手动维护模型列表

### 认证

- JWT access_token (15min) + refresh_token (7d, httpOnly cookie)
- API Key 用于 CLI / 外部调用
- 401 → 自动 refresh → 重试
- Scope 鉴权中间件

### 家庭系统

- 角色：owner / admin / member
- 加入需审核
- 小组 (FamilyGroup) 用于任务分配

### 地点系统

- Location 是一级实体（属于 Family），可选关联 FloorPlan
- 户型图用于可视化，非必须
- 被任务引用时不允许删除

### 任务系统

- 调度类型：once / daily / weekly / monthly / interval
- 请求体统一为 `{ task, extra }` 格式
- 巡检任务：检查项 → 分支 → 异常时自动创建跟进子任务
- display_summary 字段：列表视图无需额外查询

### 图片存储

- 统一 images 表管理，业务表存 image_id
- 默认本地存储，可扩展 S3/OSS

### 任务模板系统

- **架构**: `pkg/tasktemplate/` — 插件式 Provider 接口，内置 + HTTP 两种实现
- **Provider 注册表**: `init()` 自注册，主流程通过接口调用，不做任何 Provider 特化
- **内置 Provider**: Go `embed.FS` 打包 YAML 模板文件，启动时自动同步到数据库
- **HTTP Provider**: 通过订阅 URL 拉取远程 YAML 模板，支持系统级和家庭级订阅
- **双级别**: 系统模板（admin 管理，所有家庭可见）+ 家庭模板（owner 管理，仅本家庭可见）
- **数据流**: Provider.Sync() → 解析 YAML → Upsert 到 `task_templates` 表 → 前端通过 API 查询
- **模板渲染**: Go `text/template` 填充参数，生成 `task_defaults` + `extra_schema` 用于预填任务表单

### 时区与时间处理

**核心原则：后端内部全部使用 UTC，客户端在 API 边界做转换。**

#### 后端

- `pkg/timeutil` — 禁止直接使用 `time.Now()`，必须使用 `timeutil.Now()`（始终返回 UTC）。
- 所有 `time.Time` 字段在数据库和 API 响应中均为 UTC，后端不提供任何时区配置选项。
- 调度器内部计算全部使用 `time.UTC`，时区转换完全由客户端负责。

#### Web 前端

- `composables/timezone.ts` 提供完整的 UTC↔local 转换：
  - `localTimeToUTC(localTime)` / `utcTimeToLocal(utcTime)` — "HH:MM" 时间互转
  - `localDateTimeToUTC(date, time)` / `utcDateTimeToLocal(date, time)` — 日期+时间互转
  - `requestToUTC(obj)` / `responseToLocal(obj)` — schedule_data 深层次递归转换
- Axios 拦截器在请求/响应时自动调用转换函数，对业务代码透明。

#### CLI + SDK

- `sdk/timezone.go` — Go 版本的时区转换，与前端逻辑完全对齐：
  - `localTimeToUTC` / `utcTimeToLocal` / `localDateTimeToUTC` / `utcDateTimeToLocal`
  - `scheduleDataToUTC` / `scheduleDataToLocal` — schedule_data map 的递归转换
- SDK 的 `NA` 结构体持有 `timezone *time.Location` 字段，默认 `time.Local`（操作系统时区）。
- `CreateTask` / `UpdateTask` 自动将 schedule_data 中的时间从 local→UTC 再发送。
- `FormatTime(t, layout)` 将 UTC 时间转为本地时区后格式化显示。
- 模板创建任务（`CreateTaskFromTemplate`）跳过转换，因为模板渲染结果已是 UTC。

> **用户视角**：在 Web UI 或 CLI 中输入时间时，始终使用本地时间，转换完全透明。

### 错误处理体系

- 前端统一使用 `useErrorHandler()` + `<ErrorDisplay>` 组件
- ErrorDisplay 支持三种展示模式：`toast`（居中自动消失）、`dialog`（模态弹窗）、`inline`（内联警告框）
- 展示模式通过插件式注册表 `displayModes` 按错误码映射，可通过 `registerDisplayMode()` 扩展
- 颜色按 severity 分级：`info`（蓝）、`warning`（琥珀）、`error`（红）、`success`（绿）
