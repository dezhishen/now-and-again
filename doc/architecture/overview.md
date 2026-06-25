# Now & Again — 系统架构

## 技术栈

| 层 | 技术 | 说明 |
|---|------|------|
| 前端 | Vue 3 + Vite + TypeScript + Pinia + Tailwind CSS + vue-i18n | SPA, pnpm |
| 后端 | Go + Gin + GORM | RESTful API |
| CLI | Go + Cobra + Viper | 通过 HTTP / API Key 调用后端 |
| 数据库 | SQLite (开发) / PostgreSQL (生产) | GORM AutoMigrate |
| 共享层 | `shared/types` + `shared/contracts` | backend/CLI 编译期强制同步 |

## 分层架构

```
┌─────────────────────────────────────────────────┐
│  Frontend (Vue 3)    CLI (Cobra)                │  ← 客户端
├─────────────────────────────────────────────────┤
│  Handler (Gin)       Middleware (JWT/API Key)    │  ← HTTP 层
├─────────────────────────────────────────────────┤
│  Service             Scheduler Engine            │  ← 业务层
├─────────────────────────────────────────────────┤
│  Repository (GORM)   Seed / Migration            │  ← 数据层
├─────────────────────────────────────────────────┤
│  SQLite / PostgreSQL                             │  ← 存储
└─────────────────────────────────────────────────┘
         ↑
  shared/contracts  ← 编译期强制同步
  shared/types      ← DTO 定义
```

## 核心模块

### 1. 调度引擎 (`backend/internal/scheduler/`)

```
ScheduleHandler 接口
├── Code() / Name() / Icon() / Category() / DefaultPriority()
├── OnCreate(task) → 任务创建时触发
├── OnComplete(task) → 任务完成时触发（now→归档, again→自动重置）
└── Reminders() → 到期前提醒时间

注册方式: handlers/ 下文件 + init() + scheduler.Register()
```

### 2. 通知引擎 (`backend/internal/notifier/`)

```
Notifier 接口 → email / push / webhook / wechat
注册方式: init() + RegisterNotifier()
```

### 3. 认证体系

```
首次访问 → /api/system/status → 未初始化 → /setup (创建管理员)
                          → 已初始化 → /login
登录 → access_token (15min, 内存) + refresh_token (7d, httpOnly cookie)
401 → 自动 POST /api/auth/refresh → 新 token → 重试
refresh 也过期 → 跳转 /login

API Key (CLI/外部): X-API-Key: na_xxxx... 或 Bearer na_xxxx...
```

### 4. Contract-First 开发

```
shared/contracts/  ← 定义接口
       │
       ├── backend/internal/service/ → 实现 (var _ 断言)
       └── cli/internal/client/      → 实现 (var _ 断言)
       
新增方法 → 两端编译失败 → 强制同步
```

## 开发工作流

```bash
make dev-backend     # :8080
make dev-frontend    # :5173 (代理到 :8080)
make check-contracts # 自动 fix-dupes → 编译检查
make ci              # 完整流水线
```
