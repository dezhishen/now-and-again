# 路线图

## v1.0.x — 已完成 ✅

家庭事务管理的基础设施全部就位。

| 模块 | 说明 |
|------|------|
| 🗄️ 数据模型 | 24 张表，GORM AutoMigrate |
| 🔐 认证体系 | JWT + Refresh Token + API Key（Scope 权限控制） |
| 👤 用户管理 | 注册/登录/个人中心/管理员面板 |
| 👨‍👩‍👧 家庭管理 | 创建/加入/邀请码/成员管理/审核 |
| 👥 小组管理 | 创建/加入/审核/成员管理 |
| 📍 地点管理 | 一级实体 + locationkind 插件系统(indoor)，可选关联户型图 |
| 🏠 户型图 | 多楼层上传/Canvas绘制/地点关联 |
| 🔧 任务调度 | Gocron 引擎，once/daily/weekly/monthly/yearly/interval |
| 📋 任务系统 | taskkind 插件系统(simple/inspection/chain)，支持复合任务编排与巡检异常跟进 |
| ✅ 待办管理 | 快速完成/备注完成/跳过，巡检分支选择 |
| 📅 ICS 订阅 | 标准 iCalendar (RFC 5545)，API Key/Basic Auth，可导入日历 App |
| 🖥️ 日历大屏 | embed 标签嵌入，支持自动刷新 |
| 📋 任务模板 | 插件式 Provider（内置 YAML + HTTP 订阅），系统/家庭双级别 |
| ⚠️ 错误处理 | 统一 ErrorCode + 字段级 details + 前端可折叠 ErrorDisplay |
| 🖥️ Web 前端 | Vue 3 + i18n 中英文 + 暗色模式 + 响应式自适应布局 |
| 🧩 Go SDK | 独立模块，高层封装（模板建任务/待办备注/短ID解析），CLI 与外部工具共用 |
| 💻 CLI 工具 | Cobra 框架，交互式初始化，一键日常，短ID匹配，GitHub Releases 分发 |
| 🤖 AI 助手集成 | OpenClaw 接入，自然语言管理家庭事务 |
| 🐳 Docker | 多阶段构建 + UPX 压缩 (2.9MB)，推送到 GHCR |
| 🔗 接口契约约束 | 编译期确保 SDK 与 Backend 实现同一套 `contracts` 接口，两端一致性由编译器保证 |
| 🧪 自动化测试 + CI | Playwright E2E 覆盖主流程，GitHub Actions 自动执行，PR 门禁保证回归质量 |

---

## v1.1.0 — 外部交互（进行中）

让 Now & Again 从"人工驱动"走向"决策驱动"。核心目标：**降低人的重复劳动 — 人能不动手就不动手，能只决策不执行就只决策**。

核心思路：对外提供 `fulfill` API，供外部系统主动完成待办；对内引入 Hook 插件体系，在待办生命周期中插入外部通知。两者结合，让"外部驱动 + 主动通知"替代"人手动操作"。

| 能力 | 说明 |
|------|------|
| 🔗 `fulfill` API | `POST /api/tasks/:task_id/fulfill` — 外部检测器主动完成待办，触发链式推进 |
| 🪝 Hook 插件体系 | 对标 `taskkind` 的插件化设计，支持 `webhook`、`mqtt`、`detector` 等外部交互插件 |
| 🎛️ 任务级 Hook 配置 | 每个任务可自定义 webhook URL，完成/生成时回调外部系统 |
| 🏗️ 前置基础（v1.0.x 已就绪） | CLI、Go SDK、API Key + Scope 权限体系，外部系统安全调用的基础 |
| 📡 通知渠道（基于 Hook 扩展） | 邮件 → 企业微信 → 钉钉 → Push，逐步覆盖 |

### 实现方式

```
外部检测器 / IoT 设备 / AI Agent
        │
        │ ① POST /api/tasks/:task_id/fulfill
        │    (API Key 鉴权 + Scope 校验)
        ▼
┌───────────────────┐    ② 查找今日 pending 待办
│  TaskService      │ ──▶ ┌───────────────┐
│  .Fulfill()       │     │ FindPending    │
└────────┬──────────┘     │ TodoForTask    │
         │                └───────────────┘
         │ ③ 复用完整完成流程
         ▼
┌───────────────────┐
│  TodoService      │ ──▶ repo.CompleteTodo (幂等)
│  .CompleteTodo()  │ ──▶ CreateUserLog
└────────┬──────────┘ ──▶ kind.Handler.OnTodo (chain 推进)
         │             ──▶ scheduler.MarkCompleted
         │             ──▶ hook.Fire(HookTodoCompleted) ← 未来插件
         ▼
    待办完成，任务链推进到下一步
```

> 核心思路：`Fulfill` 是外部入口，之后**完全复用**现有的 `CompleteTodo` 流程——日志记录、chain 推进、调度器联动全部自动生效。Hook 插件体系作为后续扩展，对标 `taskkind` 的注册模式，在完成流程中插入 webhook 等外部通知逻辑。

> **典型场景**：轮询检测器发现洗衣机洗涤完成 → 调用 `fulfill` API 自动完成"洗衣"待办 → chain 插件自动推进到"晒衣"步骤。

---

## 未来版本

| 版本 | 类比 | 目标 | 说明 |
|------|------|------|------|
| v1.2.0 | 系统知道你要干什么 | 理解习惯，辅助决策 | 家庭事务统计、耗时趋势、自动化建议 |
