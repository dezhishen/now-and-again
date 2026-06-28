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
