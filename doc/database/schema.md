# 数据库 Schema

> 由 GORM AutoMigrate 自动管理。UUID 主键由 Go 端 `BeforeCreate` 生成。

## 表清单

| # | 表名 | 说明 | 核心字段 |
|---|------|------|---------|
| 1 | `users` | 用户 | display_name, email, phone, avatar_url |
| 2 | `accounts` | 账户（登录凭据） | user_id, provider, username, password_hash |
| 3 | `roles` | 角色 | name, description |
| 4 | `user_roles` | 用户-角色关联 | user_id, role_id |
| 5 | `families` | 家庭 | name, invite_code, created_by |
| 6 | `family_members` | 家庭成员 | family_id, user_id, role, status, joined_at |
| 7 | `family_groups` | 家庭小组 | family_id, name, description, created_by |
| 8 | `family_group_members` | 小组成员 | group_id, user_id, role, status, joined_at |
| 9 | `refresh_token_models` | 刷新令牌 | user_id, token_hash, expires_at, revoked |
| 10 | `api_key_models` | API 密钥 | user_id, name, key_prefix, key_hash |
| 11 | `images` | 图片文件 | storage_type, file_path, original_name, mime_type, size |
| 12 | `floor_plans` | 户型图 | family_id, label, image_id, is_cover |
| 13 | `locations` | 地点标记 | floor_plan_id, name, point_x, point_y, color |
| 14 | `system_settings` | 系统配置 | key (PK), value |
| 15 | `task_templates` | 任务模板 | family_id, name, schedule_type, schedule_data, enabled, is_inspection, inspection_config, location_id, group_id |
| 16 | `todos` | 待办事项 | task_id, family_id, status, todo_type, inspection_result, due_start, due_date, assigned_to, completed_by |
| 17 | `task_logs` | 任务执行日志 | task_id, status, message |
| 18 | `ics_feeds` | ICS 日历订阅 | family_id, name, filter_days, auth_type, api_key_id, app_username, app_password_hash, access_token |

## 核心索引

| 表 | 索引 | 用途 |
|----|------|------|
| `accounts` | UNIQUE(username) | 登录唯一性 |
| `users` | UNIQUE(email) | 邮箱唯一性 |
| `families` | UNIQUE(invite_code) | 邀请码查找 |
| `family_members` | UNIQUE(family_id, user_id) | 防重复加入 |
| `family_group_members` | UNIQUE(group_id, user_id) | 防重复加入小组 |
| `refresh_token_models` | UNIQUE(token_hash) | 令牌查找 |
| `api_key_models` | UNIQUE(key_hash), UNIQUE(key_prefix) | API Key 查找 |

## 角色与权限

| 角色 | 说明 |
|------|------|
| `admin` | 系统管理员，可访问管理面板 |
| `user` | 普通用户 |

## 家庭成员角色

| 角色 | 说明 |
|------|------|
| `owner` | 家庭所有者，仅一个，可删除家庭 |
| `admin` | 家庭管理员，可管理成员和审核 |
| `member` | 普通成员 |

## 成员状态

| 状态 | 说明 |
|------|------|
| `active` | 已加入 |
| `pending` | 待审核 |
| `rejected` | 已拒绝（可重新申请） |

## 初始化

首次运行服务时自动创建默认管理员账户：
- 用户名：`admin`
- 密码：由环境变量 `ADMIN_DEFAULT_PASSWORD` 设置，未设置则随机生成并打印到控制台
