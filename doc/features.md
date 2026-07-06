# 核心特性

| 特性 | 说明 |
|------|------|
| 🔀 **Now & Again 双模式** | 一次性任务完成后归档；周期性任务自动计算下次到期日 |
| 🔍 **巡检驱动** | 检查项→分支→异常自动创建跟进子任务（可指定地点/小组） |
| 📋 **任务模板系统** | 插件式 Provider（内置 YAML + HTTP 远程订阅），Go template 渲染参数，系统/家庭双级别隔离 |
| 🧩 **插件化架构** | 任务类型(taskkind) + 任务模板(tasktemplate) + 调度类型(scheduler) + 地点类型(locationkind) 四插件系统，新增类型零侵入 |
| 🧱 **插件隔离规则** | 主流程只依赖 `taskkind` 公共语义（如 OwnerKind 分发），不依赖插件内部 kind 值与子表结构 |
| 📍 **地点独立管理** | 地点为一级实体，不强制绑定户型图，支持室内/户外等多种类型 |
| 👥 **家庭 + 小组分工** | 任务精确指派到小组/地点，巡检分支可独立配置 |
| 📋 **完整操作日志** | 全程记录创建/完成/跳过/巡检/跟进 |
| 🖥️ **三入口统一** | Web (Vue 3) · CLI (Cobra) · Go SDK · RESTful API — 共享数据契约 |
| 🤖 **AI 助手集成** | OpenClaw 等 AI 框架接入，自然语言管理家庭事务，零命令零代码 |
| 🔐 **API Key 权限** | 细粒度 Scope (read/write/admin)，CLI 自动创建和管理 |
| 📦 **轻量 CLI** | 单二进制 2.9MB（UPX 压缩），GitHub Releases 分发，交互式初始化 |
| 🌙 **暗色模式 + i18n** | 中英文切换 + 暗色/亮色主题 |

## 插件系统

| 插件系统 | 说明 | 扩展方式 |
|---------|------|---------|
| `taskkind` | 任务类型（simple / inspection / chain） | 实现 `Handler` 接口 + `init()` 注册 |
| `scheduler` | 调度类型（once / daily / weekly …） | 实现调度 `Handler` + 注册 |
| `tasktemplate` | 任务模板 Provider（YAML / HTTP 订阅） | 实现 `Provider` 接口 |
| `locationkind` | 地点类型（indoor …） | 同 taskkind 模式 |

> 插件隔离规则：主流程只依赖公共语义（如 `OwnerKind` 分发），不依赖插件内部的 kind 值与子表结构。新增插件不会破坏已有功能。
