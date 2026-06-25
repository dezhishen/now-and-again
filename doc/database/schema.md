# 数据库 Schema

> 由 GORM AutoMigrate 自动管理。21 张表，UUID 主键由 Go 端 `BeforeCreate` 生成。

## 表清单

| # | 表名 | 说明 | 核心字段 |
|---|------|------|---------|
| 1 | `users` | 用户 | username, email, password_hash, display_name, is_admin |
| 2 | `families` | 家庭组 | name, invite_code, created_by |
| 3 | `family_members` | 家庭成员 | family_id, user_id, role(owner/admin/member) |
| 4 | `sub_groups` | 分工小组 | family_id, name, created_by |
| 5 | `sub_group_members` | 小组成员 | sub_group_id, user_id |
| 6 | `task_type_models` | 调度类型 | code, name, category(now/again), icon, default_priority |
| 7 | `tasks` | 核心任务 | family_id, task_type_id, title, status, priority, due_date |
| 8 | `task_assignees` | 任务分配 | task_id, user_id |
| 9 | `task_dependencies` | 任务依赖 | blocked_task_id, blocker_task_id, dependency_type |
| 10 | `task_chains` | 事项链模板 | family_id, name, created_by |
| 11 | `task_chain_steps` | 链步骤 | chain_id, sort_order, task_type_id, assigned_role |
| 12 | `task_logs` | 操作日志 | task_id, user_id, action, detail(JSON) |
| 13 | `inspections` | 巡检记录 | family_id, title, status, created_by |
| 14 | `inspection_items` | 巡检项 | inspection_id, check_point, result, generated_task_id |
| 15 | `notification_channels` | 通知渠道 | code(email/push/wechat_webhook/webhook), name, config |
| 16 | `notification_templates` | 通知模板 | family_id, task_type_id, trigger_event, channel_code, title_tmpl |
| 17 | `user_channel_configs` | 用户渠道配置 | user_id, channel_code, destination, quiet_start/end |
| 18 | `notifications` | 投递记录 | user_id, task_id, channel_code, status(pending/sent/failed) |
| 19 | `refresh_token_models` | 刷新令牌 | user_id, token_hash(SHA-256), expires_at, revoked |
| 20 | `api_key_models` | API 密钥 | user_id, name, key_hash(SHA-256), key_prefix, expires_at |

## 核心索引

| 表 | 索引 | 用途 |
|----|------|------|
| `users` | UNIQUE(username), UNIQUE(email) | 登录/注册唯一性 |
| `families` | UNIQUE(invite_code) | 邀请码查找 |
| `family_members` | UNIQUE(family_id, user_id) | 防重复加入 |
| `tasks` | (family_id, status), (due_date) | 家庭看板 + 到期扫描 |
| `task_assignees` | UNIQUE(task_id, user_id) | 防重复分配 |
| `task_dependencies` | UNIQUE(blocked_task_id, blocker_task_id) | 防重复依赖 |
| `task_logs` | (task_id, created_at) | 任务时间线 |
| `refresh_token_models` | UNIQUE(token_hash) | 令牌查找 |
| `api_key_models` | UNIQUE(key_hash) | API Key 查找 |

## 调度类型

调度类型由 `backend/internal/scheduler/handlers/` 下的 `ScheduleHandler` 接口实现 + `init()` 注册。Seed 时自动同步到 `task_type_models` 表。

| code | name | category | 说明 |
|------|------|----------|------|
| `one_off` | 一次性 | now | 完成即归档 |
| `recurring_daily` | 每日循环 | again | 完成后自动重置，次日到期 |
| `inspection_driven` | 巡检驱动 | now | 巡检发现问题生成，高优先级 |

> 新增调度类型：在 `scheduler/handlers/` 下创建文件，实现 `ScheduleHandler` 接口，`init()` 中调用 `scheduler.Register()`。
