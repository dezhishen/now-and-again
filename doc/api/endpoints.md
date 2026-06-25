# API 文档

> 共 51 个端点。公开路由无鉴权，受保护路由需要 JWT 或 API Key。

## 系统初始化

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/system/status` | 无 | 返回 `{initialized: bool}` |
| POST | `/api/setup` | 无 | 创建第一个管理员（仅未初始化时可用） |

## 认证

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/auth/register` | 无 | 注册新用户 |
| POST | `/api/auth/login` | 无 | 登录 → access_token + refresh_token(cookie) |
| POST | `/api/auth/refresh` | Cookie | 刷新 access_token |
| POST | `/api/auth/logout` | Cookie | 登出，撤销 refresh_token |

## 调度类型

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/task-types` | 无 | 获取所有已注册的调度类型 |

## 用户

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/users/me` | JWT/APIKey | 获取当前用户 |
| PUT | `/api/users/me` | JWT/APIKey | 更新当前用户 |
| GET | `/api/admin/users` | JWT(admin) | 管理员查看所有用户 |

## 家庭

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families` | JWT/APIKey | 创建家庭 |
| POST | `/api/families/join` | JWT/APIKey | 通过邀请码加入 |
| GET | `/api/families/:id` | JWT/APIKey | 家庭详情 |
| GET | `/api/users/me/families` | JWT/APIKey | 我的家庭列表 |
| GET | `/api/families/:id/members` | JWT/APIKey | 成员列表 |
| PUT | `/api/families/:id/members/:uid/role` | JWT/APIKey | 修改成员角色 |
| DELETE | `/api/families/:id/members/:uid` | JWT/APIKey | 移除成员 |
| POST | `/api/families/:id/leave` | JWT/APIKey | 退出家庭 |

## 小组

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:id/subgroups` | JWT/APIKey | 创建小组 |
| GET | `/api/families/:id/subgroups` | JWT/APIKey | 小组列表 |
| POST | `/api/subgroups/:id/members` | JWT/APIKey | 添加成员 |
| DELETE | `/api/subgroups/:id/members/:uid` | JWT/APIKey | 移除成员 |

## 任务

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:id/tasks` | JWT/APIKey | 创建任务 |
| GET | `/api/families/:id/tasks` | JWT/APIKey | 任务列表(?status=&assignee_id=&page=&page_size=) |
| GET | `/api/tasks/:id` | JWT/APIKey | 任务详情 |
| PATCH | `/api/tasks/:id` | JWT/APIKey | 更新任务(状态/标题/优先级等) |
| POST | `/api/tasks/:id/assignees` | JWT/APIKey | 设置负责人 |
| POST | `/api/tasks/:id/dependencies` | JWT/APIKey | 添加依赖 |
| DELETE | `/api/tasks/:id/dependencies/:dep_id` | JWT/APIKey | 移除依赖 |

## 事项链

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:id/chains` | JWT/APIKey | 创建链模板 |
| GET | `/api/families/:id/chains` | JWT/APIKey | 链模板列表 |
| GET | `/api/chains/:id` | JWT/APIKey | 链详情 |
| POST | `/api/chains/:id/steps` | JWT/APIKey | 添加步骤 |
| PUT | `/api/chains/:id/steps/reorder` | JWT/APIKey | 重排步骤 |
| DELETE | `/api/chains/:id/steps/:step_id` | JWT/APIKey | 删除步骤 |
| POST | `/api/chains/:id/start` | JWT/APIKey | 启动链(实例化为任务) |

## 巡检

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/families/:id/inspections` | JWT/APIKey | 创建巡检 |
| GET | `/api/families/:id/inspections` | JWT/APIKey | 巡检列表 |
| GET | `/api/inspections/:id` | JWT/APIKey | 巡检详情 |
| POST | `/api/inspections/:id/items` | JWT/APIKey | 添加巡检项 |
| PATCH | `/api/inspections/:id/items/:item_id` | JWT/APIKey | 更新巡检项 |
| POST | `/api/inspections/:id/complete` | JWT/APIKey | 完成巡检 |

## 日志

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/tasks/:id/logs` | JWT/APIKey | 任务操作日志 |
| POST | `/api/tasks/:id/comments` | JWT/APIKey | 添加评论 |

## 通知

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/users/me/notifications` | JWT/APIKey | 我的通知(?page=&page_size=) |
| PUT | `/api/users/me/channel-configs` | JWT/APIKey | 配置通知渠道 |
| GET | `/api/families/:id/notification-templates` | JWT/APIKey | 通知模板列表 |
| PUT | `/api/families/:id/notification-templates` | JWT/APIKey | 更新通知模板 |

## API Key

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/users/me/api-keys` | JWT | 创建 API Key (返回 raw_key 仅一次) |
| GET | `/api/users/me/api-keys` | JWT | API Key 列表 |
| DELETE | `/api/users/me/api-keys/:key_id` | JWT | 撤销 API Key |

## HTTP 响应规范

| 状态码 | 格式 | 说明 |
|--------|------|------|
| 200 | `{"success":true,"data":...}` | 成功 |
| 201 | 同上 | 创建成功 |
| 400 | `{"success":false,"error":"..."}` | 参数错误 |
| 401 | 同上 | 未认证 → 触发自动刷新 |
| 404 | 同上 | 资源不存在 |

## 认证方式

```
JWT:        Authorization: Bearer <access_token>
API Key:    X-API-Key: na_xxxxxxxx...  或  Authorization: Bearer na_xxxxxxxx...
Refresh:    Cookie: refresh_token=<opaque> (httpOnly, Secure, SameSite)
```
