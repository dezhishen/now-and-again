# Now & Again

> *"Life is just a mix of 'Now' (one-off) and 'Again' (recurring)."*
>
> 家庭事务管理平台 — Web UI + CLI + RESTful API，三端统一。

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)
[![GitHub Release](https://img.shields.io/github/v/release/dezhishen/now-and-again?include_prereleases)](https://github.com/dezhishen/now-and-again/releases)
[![Docker Build](https://img.shields.io/github/actions/workflow/status/dezhishen/now-and-again/docker.yml?label=docker)](https://github.com/dezhishen/now-and-again/actions)

---

## 📖 名字的由来

生活中的琐事只有两种：

- **Now（此刻）**：临时起意、只做一次的事 — 取快递、给绿植换盆、预约体检。
- **Again（再次）**：循环往复、刻在生活节律里的事 — 每两周换四件套、每天铲猫砂、每月大扫除。

**Now & Again** 把它们统一管理起来，让你无论在手机、电脑还是命令行终端，都能随手处理这些生活碎片。

---

## 🎯 项目背景

家庭事务分散在各个角落：手机提醒、冰箱贴纸条、微信群里的"别忘了"……Now & Again 把它们全部收拢到一个平台：

- **Web UI** — 家庭成员在手机/电脑上查看任务大屏，一目了然
- **CLI (`na`)** — 极简命令行，一条 `na todo done` 就能完成待办，轻松融入终端工作流
- **Go SDK** — 外部工具（脚本、CI、定时任务）直接调用，无需手写 HTTP 请求
- **AI 集成** — 搭配 AI 助手，通过自然语言即可管理家庭事务

最重要的是 **易于融入现有场景**，不必打开任何界面。

---

## 🔮 未来场景：给 AI 一个状态机

Now & Again 不是另一个待办 App，而是 **给 AI 的家庭事务状态机**。

AI 理解自然语言、感知上下文，但家务需要状态管理——什么该做、谁来做、做了没有。Now & Again 提供**确定性的状态管理**，让 AI 不需要自己维护复杂的任务拓扑和进度，只需调用简单的 API 就能查询状态、推进流程。

### 洗衣→晾晒→收纳 全自动衔接 🧺

> 你早上把衣服丢进洗衣机就走了。

洗衣机完成洗涤 → 检测器调用 `fulfill` API → "洗衣"待办自动完成 → chain 插件自动推进到"晾晒"。你在办公室收到提醒，AI 跳过已完成的步骤，整条任务链自动推进。

**人驱动 AI → AI 驱动状态机 → 设备反馈结果。** Now & Again 是其中"状态机"这一环。

→ 详见 [未来场景](doc/vision.md)

---

## 🚀 快速开始

### 前置要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | ≥ 1.25 | Backend + CLI |
| Node.js | ≥ 18 | Frontend |
| pnpm | ≥ 10 | Frontend 包管理 |

### 一键启动

```bash
git clone https://github.com/dezhishen/now-and-again.git && cd now-and-again
make dev
```

> Windows 用户请使用 WSL 或 Git Bash，或分别启动后端和前端（详见 [Docker 部署](doc/deployment/docker.md)）。

### Docker 部署

```bash
docker compose up -d
```

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `NA_PORT` | `8080` | HTTP 端口 |
| `NA_ADMIN_DEFAULT_PASSWORD` | (随机) | 初始管理员密码 |
| `NA_DATA_DIR` | `./data` | 数据根目录 |
| `NA_JWT_SECRET` | (自动生成) | JWT 签名密钥 |

> 后端强制使用 UTC。客户端在 API 边界自动做 UTC ↔ 本地时区转换，对用户透明。

---

## 🗺️ 路线图

| 版本 | 目标 |
|------|------|
| v1.0.x | 数据模型、认证、家庭/小组、任务调度、插件系统、待办、ICS、模板、Web/CLI/SDK |
| v1.1.0 🚧 | fulfill API、Hook 插件体系、通知渠道（邮件/企微/钉钉） |
| v1.2.0 | 家庭事务统计、耗时趋势、自动化建议 |

→ 详见 [完整路线图](doc/roadmap.md)

---

## 🚧 开发状态

核心模块全部完成，详见 [核心特性](doc/features.md)。

当前重点：
- **v1.1.0 外部交互与通知** — fulfill API、Hook 插件
- **CLI action-id 状态机** — AI 友好的多轮交互模式

---

## 📚 文档

| 文档 | 说明 |
|------|------|
| [核心特性](doc/features.md) | 完整功能列表与插件系统介绍 |
| [未来场景](doc/vision.md) | AI + 智能家居 + 任务链自动化 |
| [架构设计](doc/architecture/overview.md) | 系统架构、分层设计 |
| [API 文档](doc/api/endpoints.md) | 73 个 RESTful 端点 |
| [数据库 Schema](doc/database/schema.md) | 数据模型与表结构 |
| [Docker 部署](doc/deployment/docker.md) | Docker 一键部署 |
| [教程索引](doc/tutorial/README.md) | 模板、巡检、家庭、日历、AI 接入 |
| [路线图](doc/roadmap.md) | 版本规划与里程碑 |
| [CLI 使用](cli/README.md) | 命令行工具安装与命令参考 |

---

## 📄 License

MIT © Now & Again Contributors
