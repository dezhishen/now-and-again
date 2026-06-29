# API 文档

> 共 59 个端点。公开路由无鉴权，受保护路由需要 JWT 或 API Key。

## 系统初始化

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/system/status` | 无 | 返回 `{initialized: bool}` |
| POST | `/api/setup` | 无 | 创建第一个管理员（仅未初始化时可用） |

## 认证

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/auth/register` | 无 | 注册新用户 |
| POST | `/api/auth/login` | 无 | 登录获取 access_token + refresh_token(cookie) |
| POST | `/api/auth/refresh` | Cookie | 刷新 access_token |
| POST | `/api/auth/logout` | Cookie | 登出 |

## 图片

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/images/:id` | 无 | 301 重定向到实际文件 |

## 用户

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/users/me` | JWT/APIKey | 获取当前用户 |
| PUT | `/api/users/me` | JWT/APIKey | 更新当前用户 |
| GET | `/api/users/me/families` | JWT/APIKey | 我的家庭列表 |
| GET | `/api/admin/users` | JWT(admin) | 管理员查看所有用户 |

## 管理面板

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/admin/settings` | JWT | 获取系统设置 |
| PUT | `/api/admin/settings` | JWT | 更新设置 |

## 家庭

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families` | JWT/APIKey | 创建家庭 |
| POST | `/api/families/join` | JWT/APIKey | 通过邀请码加入 |
| GET | `/api/families/:family_id` | JWT/APIKey | 家庭详情 |
| PATCH | `/api/families/:family_id` | JWT/APIKey | 修改家庭名称 (owner/admin) |
| DELETE | `/api/families/:family_id` | JWT/APIKey | 删除家庭 (仅 owner) |
| GET | `/api/families/:family_id/members` | JWT/APIKey | 成员列表 |
| PUT | `/api/families/:family_id/members/:user_id/role` | JWT/APIKey | 修改成员角色 (owner/admin) |
| DELETE | `/api/families/:family_id/members/:user_id` | JWT/APIKey | 移除成员 (owner/admin) |
| POST | `/api/families/:family_id/leave` | JWT/APIKey | 退出家庭 |

## 家庭审核

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/families/:family_id/join-requests` | JWT/APIKey | 待审核申请 |
| PUT | `/api/families/:family_id/join-requests` | JWT/APIKey | 审核申请 (active/rejected) |

## 小组

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:family_id/groups` | JWT/APIKey | 创建小组 |
| GET | `/api/families/:family_id/groups` | JWT/APIKey | 小组列表 |
| POST | `/api/groups/:group_id/join` | JWT/APIKey | 加入小组 |
| POST | `/api/groups/:group_id/leave` | JWT/APIKey | 离开小组 |
| GET | `/api/groups/:group_id/members` | JWT/APIKey | 小组成员 |
| DELETE | `/api/groups/:group_id/members/:user_id` | JWT/APIKey | 移除成员 |
| GET | `/api/groups/:group_id/join-requests` | JWT/APIKey | 待审核申请 |
| PUT | `/api/groups/:group_id/join-requests` | JWT/APIKey | 审核申请 |

## API Key

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/users/me/api-keys` | JWT/APIKey | 创建 API Key |
| GET | `/api/users/me/api-keys` | JWT/APIKey | 列出 API Key |
| DELETE | `/api/users/me/api-keys/:key_id` | JWT/APIKey | 撤销 API Key |

## 户型图

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:family_id/floor-plans` | JWT/APIKey | 上传户型图 (multipart: file + label + is_cover) |
| GET | `/api/families/:family_id/floor-plans` | JWT/APIKey | 户型图列表 |
| GET | `/api/floor-plans/:plan_id` | JWT/APIKey | 户型图详情 |
| PUT | `/api/floor-plans/:plan_id/cover` | JWT/APIKey | 设为封面 |
| DELETE | `/api/floor-plans/:plan_id` | JWT/APIKey | 删除户型图 |

## 地点（一级实体）

> Location 属于 Family，可选关联到 FloorPlan。kind 插件化（indoor 等）。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/families/:family_id/locations` | JWT/APIKey | 家庭所有地点 |
| POST | `/api/families/:family_id/locations` | JWT/APIKey | 创建地点 (name + kind + color, 可选 floor_plan_id) |
| GET | `/api/floor-plans/:plan_id/locations` | JWT/APIKey | 户型图关联地点 |
| PUT | `/api/locations/:location_id` | JWT/APIKey | 更新地点（可设置/取消户型图关联） |
| DELETE | `/api/locations/:location_id` | JWT/APIKey | 删除地点（被任务引用时拒绝） |

## 任务

> 请求体：`{ task: {...}, extra: {...} }`，与 GET 返回格式对称。
> 任务类型插件化：simple / inspection。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:family_id/tasks` | JWT/APIKey | 创建任务 |
| GET | `/api/families/:family_id/tasks` | JWT/APIKey | 任务列表 |
| GET | `/api/tasks/:task_id` | JWT/APIKey | 获取任务 (with_extra=true 返回插件数据) |
| PUT | `/api/tasks/:task_id` | JWT/APIKey | 更新任务（含 extra 数据时触发插件 OnUpdate） |
| PUT | `/api/tasks/:task_id/enabled` | JWT/APIKey | 启用/禁用任务（不触发插件，仅更新调度器） |
| DELETE | `/api/tasks/:task_id` | JWT/APIKey | 删除任务 |
| GET | `/api/tasks/:task_id/logs` | JWT/APIKey | 操作日志 |
| POST | `/api/tasks/:task_id/trigger` | JWT/APIKey | 手动生成待办 |

### 创建巡检任务示例

```json
{
  "task": {
    "name": "厨房安全巡检",
    "schedule_type": "daily",
    "schedule_data": {"time": "21:00"},
    "kind": "inspection"
  },
  "extra": {
    "check_items": [
      {
        "name": "燃气阀门",
        "branches": [
          {"name": "正常", "create_todo": false},
          {"name": "异常", "create_todo": true, "todo_name": "修复燃气阀门"}
        ]
      }
    ]
  }
}
```

## 待办

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/families/:family_id/todos` | JWT/APIKey | 待办列表 (?status=pending) |
| GET | `/api/todos/:todo_id` | JWT/APIKey | 待办详情 (with_extra=true) |
| PUT | `/api/todos/:todo_id` | JWT/APIKey | 完成/跳过待办 |

## 日历

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/families/:family_id/calendar` | JWT/APIKey | 日历视图 |

## ICS 订阅

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:family_id/ics-feeds` | JWT/APIKey | 创建 ICS 订阅 |
| GET | `/api/families/:family_id/ics-feeds` | JWT/APIKey | 订阅列表 |
| GET | `/api/ics-feeds/:feed_id` | JWT/APIKey | 订阅详情 |
| PUT | `/api/ics-feeds/:feed_id` | JWT/APIKey | 更新订阅 |
| DELETE | `/api/ics-feeds/:feed_id` | JWT/APIKey | 删除订阅 |
| GET | `/api/ics/:token` | 无 | ICS 日历端点（支持 API Key / Basic Auth） |

## 认证方式

- JWT: `Authorization: Bearer <access_token>`
- API Key: `X-API-Key: na_xxxx` 或 `Authorization: Bearer na_xxxx`
- Refresh: Cookie `refresh_token` (httpOnly)

## 统一错误响应

所有非 2xx 响应遵循统一格式：

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "summary": "参数校验失败",
    "details": [
      { "field": "name", "message": "不能为空" }
    ]
  }
}
```

### 错误码

| Code | HTTP | 含义 | details |
|------|------|------|---------|
| `BAD_REQUEST` | 400 | 请求格式错误 | — |
| `VALIDATION_ERROR` | 400 | 字段校验失败 | `FieldError[]` |
| `UNAUTHORIZED` | 401 | 未认证 | — |
| `FORBIDDEN` | 403 | 无权限 | — |
| `NOT_FOUND` | 404 | 资源不存在 | — |
| `CONFLICT` | 409 | 数据冲突 | — |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 | — |

### 后端实现

- `pkg/types/common.go` — `ErrorCode`、`APIError`、`FieldError` 类型定义
- `internal/handler/handlers.go` — `apiError()`/`validationError()`/`notFound()` 等统一响应辅助函数
- 所有 Handler 中 `bindJSON` 错误统一调用 `validationError()`，返回结构化字段级错误

### 前端处理

- `types/index.ts` — `ApiRequestError` 类，继承 `Error`，含 `code`/`summary`/`details`
- `api/client.ts` — 解析 `APIError` 并抛出 `ApiRequestError`
- `composables/useErrorHandler.ts` — 按 `ErrorCode` 映射 i18n 摘要
- `components/ErrorDisplay.vue` — 可折叠错误展示组件（400 琥珀色 / 500 红色）
