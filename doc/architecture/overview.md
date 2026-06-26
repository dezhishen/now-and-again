# Now & Again — 系统架构

## 技术栈

| 层 | 技术 | 说明 |
|---|------|------|
| 前端 | Vue 3 + Vite + TypeScript + Pinia + Tailwind CSS + vue-i18n | SPA, pnpm |
| 后端 | Go + Gin + GORM | RESTful API，`backend/pkg/` 为可复用公共库 |
| CLI | Go + Cobra + Viper | 通过 HTTP / API Key 调用后端 |
| 数据库 | SQLite (开发) / PostgreSQL (生产) | GORM AutoMigrate |

## 分层架构

```
┌─────────────────────────────────────────────────┐
│  Frontend (Vue 3)    CLI (Cobra)                │  ← 客户端
├─────────────────────────────────────────────────┤
│  Handler (Gin)       Middleware (JWT/API Key)    │  ← HTTP 层
├─────────────────────────────────────────────────┤
│  Service                                        │  ← 业务层
├─────────────────────────────────────────────────┤
│  Repository (GORM)   Seed / Migration            │  ← 数据层
├─────────────────────────────────────────────────┤
│  SQLite / PostgreSQL                             │  ← 存储
└─────────────────────────────────────────────────┘

backend/pkg/   ← 类型定义、调度器、任务类型插件（CLI 可引用）
```

## 项目结构

```
backend/
  cmd/server/        入口
  pkg/               公共包（CLI 可直接引用）
    types/           共享 DTO
    contracts/       API 接口定义
    scheduler/       调度引擎 + 内置处理器
    taskkind/        任务类型插件系统（simple, inspection）
    scopes/          权限范围
  internal/          内部实现
    handler/         HTTP 处理器 (10 files)
    service/         业务逻辑 (8 files)
    repository/      数据访问 (12 files)
    middleware/      认证/鉴权
    config/          配置
  migrations/        数据库迁移

cli/
  cmd/               CLI 命令
  internal/client/   HTTP 客户端

frontend/
  src/               Vue 3 SPA

## 核心模块

### 1. 认证体系

```
首次运行 → 自动创建管理员 (admin + 随机密码/环境变量)
登录 → access_token (15min, 内存) + refresh_token (7d, httpOnly cookie)
401 → 自动 POST /api/auth/refresh → 新 token → 重试
refresh 也过期 → 跳转 /login

API Key (CLI/外部): X-API-Key: na_xxxx... 或 Bearer na_xxxx...
```

### 2. 家庭系统

- 每个用户仅能创建一个家庭，但可加入多个
- 家庭角色：owner / admin / member
- 加入需审核（owner/admin 审批）
- 被拒绝后可重新申请

### 3. 户型图系统

- 支持多楼层（floor_plans 表，每层一个记录）
- 上传图片或手动绘制（Canvas → PNG → 上传）
- 封面楼层：家庭列表卡片缩略图来源
- 地点标记：在户型图上点击标记位置，支持颜色区分
- 图片统一由 `images` 表管理，业务表仅存 `image_id`
- 图片访问：`GET /api/images/:id` → 301 → `/uploads/{filename}`
- 绘制工具：吸附网格（20px）、线段绘制（起点→终点）

### 4. 任务系统

- 任务类型：`simple`（普通任务）、`inspection`（巡检任务）
- 巡检任务：创建检查项清单，发现问题时自动生成一次性跟进任务
- 调度类型：once / daily / weekly / monthly / interval
- 待办状态：pending → done / skipped
- 完整操作日志（task_logs 表）

### 5. ICS 日历 & 大屏嵌入

- 标准 iCalendar 协议（RFC 5545），可导入 Apple/Google/Outlook 日历
- 认证方式：API Key（`?key=`）或 Basic Auth
- 大屏日历：`<embed>` 标签嵌入任意网页，支持 `?key=` 免登录 + `?refresh=N` 自动刷新

### 6. 图片存储

- `images` 表记录：storage_type, file_path, original_name, mime_type, size
- 当前支持 `local` 存储，架构预留 `s3`/`minio`/`oss` 扩展
- 存储类型通过管理面板的 `storage.type` 系统设置配置

### 7. 系统配置

- `system_settings` 表：key-value 键值对
- 管理面板可编辑（`/admin` → 存储配置 Tab）
- 默认配置在 Seed 阶段写入

### 8. Contract-First 开发

```
backend/pkg/contracts/  ← 定义接口
       │
       ├── backend/internal/service/ → 实现 (var _ 断言)
       
新增方法 → 编译失败 → 强制同步
```

## 前端路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/setup` | 初始化 | 系统首次设置（已有管理员则跳过） |
| `/login` | 登录 | |
| `/register` | 注册 | |
| `/` | 首页 | 家庭卡片列表（缩略图 + 收藏 + 创建/加入） |
| `/admin` | 管理面板 | 用户管理 + 存储配置 |
| `/api-keys` | API Key 管理 | 创建/撤销 API Key，Scope 配置 |
| `/family/:id` | 家庭 | 页签式导航（dashboard/tasks/groups/members/floor-plan/ics/settings） |
| `/calendar/:id` | 日历大屏 | 全屏日历视图，支持 `?key=` 免登录访问 + `?refresh=` 自动刷新 |

## 数据目录

`DATA_DIR` 环境变量控制数据存储位置：
```
$DATA_DIR/
├── now-and-again.db    # SQLite 数据库
└── uploads/             # 上传文件
```

开发环境默认：`../data`（项目根目录下的 `data/`）。

## 开发工作流

```bash
make dev              # 并行启动 backend (8080) + frontend (5173)
make dev-backend      # 仅后端
make dev-frontend     # 仅前端
make db-reset         # 删除 SQLite 数据库
```

## 目录结构

```
backend/
├── cmd/server/main.go        # 入口
├── pkg/                      # 公共包
│   ├── contracts/            # 接口定义
│   ├── scheduler/            # 调度引擎
│   ├── taskkind/             # 任务类型插件 (simple, inspection)
│   ├── scopes/               # 权限范围
│   └── types/                # DTO 定义
├── internal/
│   ├── config/               # 配置加载
│   ├── handler/               # HTTP 处理器 (10 files)
│   ├── middleware/            # JWT/CORS
│   ├── repository/            # 数据访问 (12 files)
│   └── service/               # 业务逻辑 (8 files)
├── migrations/               # (空，使用 GORM AutoMigrate)

frontend/
├── src/
│   ├── api/                  # API 客户端
│   ├── components/
│   ├── composables/          # 组合式函数
│   ├── i18n/                 # 国际化 (zh-CN, en)
│   ├── router/               # 路由
│   ├── stores/               # Pinia 状态管理
│   ├── styles/               # 全局样式
│   ├── types/                # TypeScript 类型
│   └── views/                # 页面组件 (12 views)
└── vite.config.ts            # Vite 配置（API + Uploads 代理）

cli/                          # Cobra CLI
```
