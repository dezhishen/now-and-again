# Now & Again — 系统架构

## 技术栈

| 层 | 技术 |
|---|------|
| 前端 | Vue 3 + Vite + TypeScript + Pinia + Tailwind CSS + vue-i18n (pnpm) |
| 后端 | Go 1.25 + Gin + GORM (github.com/glebarez/sqlite, 纯 Go 无 CGO) |
| CLI | Go + Cobra，通过 HTTP / API Key 调用后端 |
| 数据库 | SQLite (默认) / PostgreSQL |

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
│  SQLite / PostgreSQL                     │
└──────────────────────────────────────────┘

backend/pkg/ — 公共类型、调度器、插件系统（CLI 直接引用）
```

## 项目结构

```
backend/
  cmd/server/main.go      入口
  internal/
    handler/               HTTP 处理器
    service/               业务逻辑（user/family/task/todo/log/floorplan/ics）
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
    locationkind/          地点类型插件 (indoor)
    scopes/                权限范围

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
| 任务类型 | `pkg/taskkind/` | `useTaskKinds` | simple, inspection |
| 地点类型 | `pkg/locationkind/` | `useLocationKinds` | indoor |
| 调度类型 | `pkg/scheduler/` | — | once, daily, weekly, monthly, interval |

新增类型只需实现接口并注册（`init()` 自动注册），无需修改任何现有代码。

### 任务插件生命周期

```
taskkind.Handler
  ├─ SaveExtra(task, extra)    ← 新建时持久化插件特有数据
  ├─ UpdateExtra(task, extra)  ← 更新时按 ID 精细化 diff（非全删全建）
  ├─ DeleteExtra(task)         ← 删除时清理插件数据
  ├─ OnComplete(todo, extra)   ← 待办完成时的业务逻辑
  └─ GetExtra(task)            ← 读取插件数据供前端展示

taskkind.TaskStorage（注入到插件的方法集合）
  ├─ CreateNoRootTask(task, extra)  ← 创建子任务并触发其 SaveExtra
  ├─ UpdateNoRootTask(task)         ← 更新子任务
  └─ DeleteNonRootTask(id)          ← 递归删除子任务树，触发 DeleteExtra
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

### 错误处理体系

- 前端统一使用 `useErrorHandler()` + `<ErrorDisplay>` 组件
- ErrorDisplay 支持三种展示模式：`toast`（居中自动消失）、`dialog`（模态弹窗）、`inline`（内联警告框）
- 展示模式通过插件式注册表 `displayModes` 按错误码映射，可通过 `registerDisplayMode()` 扩展
- 颜色按 severity 分级：`info`（蓝）、`warning`（琥珀）、`error`（红）、`success`（绿）
