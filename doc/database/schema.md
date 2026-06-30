# 数据库 Schema

> GORM AutoMigrate 管理，UUID 主键由 `BaseModel.BeforeCreate` 生成。

## 表清单

| 表名 | 说明 | 核心字段 |
|------|------|---------|
| `users` | 用户 | display_name, email, phone, avatar_url |
| `accounts` | 账户 | user_id, provider, username, password_hash |
| `roles` | 角色 | name, description |
| `user_roles` | 用户角色关联 | user_id, role_id |
| `families` | 家庭 | name, invite_code, created_by |
| `family_members` | 家庭成员 | family_id, user_id, role, status |
| `family_groups` | 家庭小组 | family_id, name, description |
| `family_group_members` | 小组成员 | group_id, user_id, role, status |
| `refresh_token_models` | 刷新令牌 | user_id, token_hash, expires_at |
| `api_key_models` | API Key | user_id, name, key_prefix, key_hash, scopes |
| `images` | 图片文件 | storage_type, file_path, original_name, mime_type, size |
| `floor_plans` | 户型图 | family_id, label, image_id, is_cover, width, height |
| `locations` | 地点（一级实体） | family_id, floor_plan_id(可选), kind, name, color |
| `system_settings` | 系统配置 | key (PK), value |
| `tasks` | 任务模板 | family_id, name, schedule_type, schedule_data, enabled, kind, display_summary, group_id, location_id, parent_task_id, is_root |
| `todos` | 待办事项 | task_id, family_id, location_id, status, branch_name, remark, due_start, due_date |
| `task_logs` | 操作日志 | task_id, todo_id, status, message, log_type, operator_id |
| `check_items` | 巡检检查项 | task_id, name, sort_order |
| `check_item_branches` | 检查项分支 | check_item_id, name, create_todo, branch_task_id, sort_order |
| `inspection_results` | 巡检结果 | task_id, todo_id, family_id, item_name, branch_name, created_by |
| `ics_feeds` | ICS 订阅 | family_id, name, filter_days, auth_type, api_key_id, access_token |
| `task_templates` | 任务模板 | family_id(NULL=系统级), provider_code, template_code, name, kind, icon, parameters, task_defaults, extra_schema |
| `task_template_subscriptions` | 模板订阅源 | family_id(NULL=系统级), provider_code, url, name, auto_refresh, refresh_interval_hours |

## 核心索引

| 表 | 索引 |
|----|------|
| `accounts` | UNIQUE(username) |
| `users` | UNIQUE(email) |
| `families` | UNIQUE(invite_code) |
| `family_members` | UNIQUE(family_id, user_id) |
| `family_group_members` | UNIQUE(group_id, user_id) |
| `refresh_token_models` | UNIQUE(token_hash) |
| `api_key_models` | UNIQUE(key_hash), UNIQUE(key_prefix) |
| `tasks` | (is_root, family_id), family_id, group_id, location_id, parent_task_id |
| `locations` | family_id, floor_plan_id |
| `task_templates` | UNIQUE(family_id, provider_code, template_code) |
| `task_template_subscriptions` | (family_id, provider_code) |

## 角色与权限

| 角色 | 说明 |
|------|------|
| `admin` | 系统管理员 |
| `user` | 普通用户 |

## 家庭成员角色

| 角色 | 说明 |
|------|------|
| `owner` | 家庭所有者 |
| `admin` | 家庭管理员 |
| `member` | 普通成员 |

## 成员状态

| 状态 | 说明 |
|------|------|
| `active` | 已加入 |
| `pending` | 待审核 |
| `rejected` | 已拒绝 |

## 任务模板

### task_templates

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | char(36) PK | UUID |
| `family_id` | char(36) nullable | NULL = 系统级，非 NULL = 家庭级 |
| `provider_code` | varchar(32) | 来源（builtin/http/family） |
| `template_code` | varchar(64) | 模板唯一标识 |
| `name` | varchar(128) | 模板名称 |
| `description` | varchar(512) | 描述 |
| `kind` | varchar(16) | 任务类型（simple/inspection） |
| `icon` | varchar(32) | 图标 emoji |
| `sort_order` | int | 排序 |
| `enabled` | bool | 是否启用 |
| `parameters` | text | JSON：参数定义列表 |
| `task_defaults` | text | JSON：任务默认字段（支持 Go template） |
| `extra_schema` | text | JSON：任务类型专属 extra 字段 |
| `version` | varchar(32) | 版本号 |
| `metadata` | text | Provider 元数据（如源 URL） |

索引：`(provider_code, template_code)` 联合索引

### task_template_subscriptions

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | char(36) PK | UUID |
| `family_id` | char(36) nullable | NULL = 系统级订阅 |
| `provider_code` | varchar(32) | Provider 标识（http） |
| `url` | varchar(2048) | 订阅 URL |
| `name` | varchar(128) | 显示名称 |
| `auto_refresh` | bool | 是否自动刷新 |
| `refresh_interval_hours` | int | 刷新间隔（小时） |
| `enabled` | bool | 是否启用 |
