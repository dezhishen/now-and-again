# Backend ↔ CLI 联动契约

> ⚠️ **Vibecoding 重要提示**：backend 和 CLI 通过 `shared/contracts` 接口实现**编译期强制同步**。

---

## 强制联动机制（Go 接口）

```
                     shared/contracts/
                    (Go 接口 — 唯一真理源)
                    ┌─────────────────────┐
                    │  UserContract        │
                    │  FamilyContract      │
                    │  TaskContract        │
                    │  ChainContract       │
                    │  InspectionContract  │
                    │  LogContract         │
                    │  NotificationContract│
                    └─────────┬───────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
    backend/internal/service/        cli/internal/client/
    ┌─────────────────────┐         ┌─────────────────────┐
    │ var _ UserContract  │         │ var _ UserContract  │
    │   = (*UserService)  │         │   = (*UserClient)   │
    │         (nil)       │         │         (nil)       │
    └─────────────────────┘         └─────────────────────┘
         编译期断言                      编译期断言
```

**向接口添加新方法 → backend 和 CLI 同时编译失败 → 强制同步实现。**

---

## 新增方法标准流程

```
1. shared/contracts/contracts.go → 在接口中添加方法签名
         ↓
2. 编译 backend → ❌ 失败（service 未实现）
   编译 cli     → ❌ 失败（client 未实现）
         ↓
3. backend/internal/service/xxx_service.go → 添加实现
   cli/internal/client/xxx_client.go       → 添加实现
         ↓
4. 编译通过 ✅
```

---

## 接口 ↔ 实现对照表

| 接口 (`shared/contracts`) | Backend 实现 | CLI 实现 | 断言位置 |
|------|-------------|---------|---------|
| `UserContract` | `service.UserService` | `client.UserClient` | `contracts.go` |
| `FamilyContract` | `service.FamilyService` | `client.FamilyClient` | 同上 |
| `SubGroupContract` | `service.SubGroupService` | `client.SubGroupClient` | 同上 |
| `TaskContract` | `service.TaskService` | `client.TaskClient` | 同上 |
| `ChainContract` | `service.ChainService` | `client.ChainClient` | 同上 |
| `InspectionContract` | `service.InspectionService` | `client.InspectionClient` | 同上 |
| `LogContract` | `service.LogService` | `client.LogClient` | 同上 |
| `NotificationContract` | `service.NotificationService` | `client.NotificationClient` | 同上 |

---

## REST API ↔ CLI 命令对照

| API 路由 | CLI 命令 | 调用 Client 方法 |
|------|---------|---------|
| `POST /api/auth/login` | `na login` | `UserClient.Login` |
| `POST /api/families` | `na family create` | `FamilyClient.Create` |
| `GET /api/families/:id/tasks` | `na task list` | `TaskClient.List` |
| `POST /api/families/:id/tasks` | `na task create` | `TaskClient.Create` |
| `PATCH /api/tasks/:id` | `na task update` | `TaskClient.Update` |
| `POST /api/chains/:id/start` | `na chain start` | `ChainClient.Start` |
| `POST /api/families/:id/inspections` | `na inspect start` | `InspectionClient.Create` |

---

## Vibecoding Checklist

### 修改 contract 接口（最强约束）

- [ ] 在 `shared/contracts/contracts.go` 添加/修改方法
- [ ] backend 编译 → 修复 `service/xxx_service.go`
- [ ] CLI 编译 → 修复 `client/xxx_client.go`
- [ ] handler 层适配（`handler/xxx_handler.go`）
- [ ] routes.go 注册新路由（如需）

### 修改 shared/types 字段

- [ ] frontend `src/types/index.ts` TypeScript 类型同步更新

### 新增 TaskType / NotificationChannel

- [ ] `backend/internal/repository/db.go` 的 `Seed()` 函数

### 新增通知渠道实现

- [ ] `backend/internal/notifier/` 下新建文件 + `RegisterNotifier()`

---

## 开发顺序（Contract-First）

1. **shared/contracts** → 定义接口
2. **编译确认失败** → 两端都需要改
3. **backend service** → 业务逻辑
4. **CLI client** → HTTP 调用
5. **handler adapter** → Gin 路由
6. **frontend** → UI

---

## HTTP 响应规范

| 状态码 | JSON 格式 | 说明 |
|--------|----------|------|
| 200 | `{"success":true,"data":...}` | 成功 |
| 201 | 同上 | 创建成功 |
| 400 | `{"success":false,"error":"..."}` | 参数错误 |
| 401 | 同上 | 未认证 → CLI 提示 `na login` |
| 404 | 同上 | 资源不存在 |

---

## Token 管理

- 后端返回：`{"token":"...", "expires_at":1234567890}`
- CLI 存储：`~/.na.yaml` → `token` 字段
- CLI 环境变量：`NA_TOKEN`
- CLI 发送：`Authorization: Bearer <token>`
