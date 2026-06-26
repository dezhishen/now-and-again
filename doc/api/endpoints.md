# API 文档

> 共 67 个端点。公开路由无鉴权，受保护路由需要 JWT 或 API Key（自动校验 Scope）。

## 系统初始化

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/system/status` | 无 | 返回 `{initialized: bool}` |
| POST | `/api/setup` | 无 | 创建第一个管理员（仅未初始化时可用） |

> 首次运行服务时自动创建默认管理员账户（username: `admin`），密码由 `ADMIN_DEFAULT_PASSWORD` 环境变量控制。

## 认证

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/auth/register` | 无 | 注册新用户 |
| POST | `/api/auth/login` | 无 | 登录 → access_token + refresh_token(cookie) |
| POST | `/api/auth/refresh` | Cookie | 刷新 access_token |
| POST | `/api/auth/logout` | Cookie | 登出，撤销 refresh_token |

## 图片

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/images/:id` | 无 | 301 重定向到实际文件（`/uploads/{filename}`） |

> 图片表统一管理所有文件，业务表只存 `image_id`，支持未来扩展 S3/OSS 等存储后端。

## 用户

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/users/me` | JWT/APIKey | 获取当前用户 |
| PUT | `/api/users/me` | JWT/APIKey | 更新当前用户 |
| GET | `/api/admin/users` | JWT(admin) | 管理员查看所有用户 |

## 管理面板

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/admin/settings` | JWT | 获取所有系统设置 |
| PUT | `/api/admin/settings` | JWT | 批量更新设置（JSON: `{"key":"value"}`） |

> 系统设置当前支持 `storage.type`（默认 `"local"`），未来可扩展为 `"s3"`、`"minio"` 等。

## 家庭

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families` | JWT/APIKey | 创建家庭（每个用户仅能创建一个） |
| POST | `/api/families/join` | JWT/APIKey | 通过邀请码加入 |
| GET | `/api/families/:family_id` | JWT/APIKey | 家庭详情 |
| PATCH | `/api/families/:family_id` | JWT/APIKey | 修改家庭名称（owner/admin） |
| DELETE | `/api/families/:family_id` | JWT/APIKey | 删除家庭（仅 owner） |
| GET | `/api/users/me/families` | JWT/APIKey | 我的家庭列表（含封面缩略图） |
| GET | `/api/families/:family_id/members` | JWT/APIKey | 成员列表 |
| PUT | `/api/families/:family_id/members/:user_id/role` | JWT/APIKey | 修改成员角色（owner/admin） |
| DELETE | `/api/families/:family_id/members/:user_id` | JWT/APIKey | 移除成员（owner/admin） |
| POST | `/api/families/:family_id/leave` | JWT/APIKey | 退出家庭 |

## 家庭审核

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/families/:family_id/join-requests` | JWT/APIKey | 待审核申请列表 |
| PUT | `/api/families/:family_id/join-requests` | JWT/APIKey | 审核申请（action: `active`/`rejected`） |

## 小组

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:family_id/groups` | JWT/APIKey | 创建小组 |
| GET | `/api/families/:family_id/groups` | JWT/APIKey | 小组列表 |
| POST | `/api/groups/:group_id/join` | JWT/APIKey | 加入小组 |
| POST | `/api/groups/:group_id/leave` | JWT/APIKey | 离开小组 |
| GET | `/api/groups/:group_id/members` | JWT/APIKey | 小组成员列表 |
| DELETE | `/api/groups/:group_id/members/:user_id` | JWT/APIKey | 移除小组成员 |
| GET | `/api/groups/:group_id/join-requests` | JWT/APIKey | 小组待审核申请 |
| PUT | `/api/groups/:group_id/join-requests` | JWT/APIKey | 审核小组申请 |

## 户型图

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:family_id/floor-plans` | JWT/APIKey | 上传户型图（multipart: file + label + is_cover） |
| GET | `/api/families/:family_id/floor-plans` | JWT/APIKey | 列出所有楼层 |
| GET | `/api/floor-plans/:plan_id` | JWT/APIKey | 获取单层详情（含地点列表） |
| PUT | `/api/floor-plans/:plan_id/cover` | JWT/APIKey | 设为封面 |
| DELETE | `/api/floor-plans/:plan_id` | JWT/APIKey | 删除楼层及图片 |

## 地点标记

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/floor-plans/:plan_id/locations` | JWT/APIKey | 创建地点（name + point + color） |
| GET | `/api/floor-plans/:plan_id/locations` | JWT/APIKey | 地点列表 |
| PUT | `/api/locations/:location_id` | JWT/APIKey | 更新地点 |
| DELETE | `/api/locations/:location_id` | JWT/APIKey | 删除地点 |

## API Key

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/users/me/api-keys` | JWT | 创建 API Key（返回 raw_key 仅一次） |
| GET | `/api/users/me/api-keys` | JWT | API Key 列表 |
| DELETE | `/api/users/me/api-keys/:key_id` | JWT | 撤销 API Key |

> Scope 说明：创建时可传 `scopes[]`，支持精确 scope（如 `family:read`）和快捷组（`read`/`write`/`admin`）。详细定义见 `shared/scopes/scopes.go`。

## 任务模板

| 方法 | 路径 | 鉴权 | Scope | 说明 |
|------|------|------|-------|------|
| POST | `/api/families/:family_id/tasks` | JWT/APIKey | task:write | 创建任务模板 |
| GET | `/api/families/:family_id/tasks` | JWT/APIKey | task:read | 任务模板列表 |
| PUT | `/api/tasks/:task_id` | JWT/APIKey | task:write | 更新任务模板 |
| DELETE | `/api/tasks/:task_id` | JWT/APIKey | task:write | 删除任务模板 |
| POST | `/api/tasks/:task_id/trigger` | JWT/APIKey | task:write | 手动生成待办 |
| GET | `/api/tasks/:task_id/logs` | JWT/APIKey | task:read | 操作日志（?type=user 过滤用户日志） |

> 任务统一模型：`kind` 字段区分类型 — `"simple"` 普通任务 / `"branched"` 分支任务。未来可扩展 `"chain"` 等类型。
> 分支任务通过 `branches` 数组配置选项，每个分支可选是否创建跟进任务。

### 创建普通任务

```json
{
  "name": "取快递",
  "schedule_type": "once",
  "schedule_data": {"date": "2026-06-28", "time": "18:00"},
  "kind": "simple"
}
```

### 创建分支任务

```json
{
  "name": "厨房安全巡检",
  "schedule_type": "daily",
  "schedule_data": {"time": "21:00"},
  "kind": "branched",
  "branches": [
    {"name": "正常", "create_todo": false},
    {"name": "漏水", "create_todo": true, "todo_name": "修复漏水-{name}"}
  ]
}
```

## 待办

| 方法 | 路径 | 鉴权 | Scope | 说明 |
|------|------|------|-------|------|
| GET | `/api/families/:family_id/todos?status=pending` | JWT/APIKey | task:read | 待办列表 |
| PUT | `/api/todos/:todo_id` | JWT/APIKey | task:write | 完成/跳过待办 |

### 完成分支任务的待办

```json
{
  "status": "done",
  "branch_name": "漏水"
}
```
> 选择分支后，系统自动创建独立的一次性跟进任务和待办。

## ICS 日历订阅

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:family_id/ics-feeds` | JWT | 创建 ICS 订阅（auth_type: api_key/basic） |
| GET | `/api/families/:family_id/ics-feeds` | JWT | 订阅列表 |
| GET | `/api/ics-feeds/:feed_id` | JWT | 查看订阅详情 |
| PUT | `/api/ics-feeds/:feed_id` | JWT | 更新订阅 |
| DELETE | `/api/ics-feeds/:feed_id` | JWT | 删除订阅 |
| GET | `/api/ics/:token.ics` | 无(JWT)/自定义 | 公开 ICS 端点，支持 API Key（?key=）或 Basic Auth |

## HTTP 响应规范

| 状态码 | 格式 | 说明 |
|--------|------|------|
| 200 | `{"success":true,"data":...}` | 成功 |
| 201 | 同上 | 创建成功 |
| 301 | 重定向 | 图片服务重定向 |
| 400 | `{"success":false,"error":"..."}` | 参数错误 |
| 401 | 同上 | 未认证 → 触发自动刷新 |
| 404 | 同上 | 资源不存在 |
| 500 | 同上 | 服务器错误 |

## 认证方式

```
JWT:        Authorization: Bearer <access_token>
API Key:    X-API-Key: na_xxxxxxxx...  或  Authorization: Bearer na_xxxxxxxx...
Refresh:    Cookie: refresh_token=<opaque> (httpOnly, Secure, SameSite)
```
