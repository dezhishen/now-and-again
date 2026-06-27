# Now & Again CLI (`na`)

## 安装

```bash
cd cli && go build -o na . && sudo mv na /usr/local/bin/
```

## 认证

```bash
# 登录获取 token
na login -u admin -p secret

# 设置环境变量（推荐）
export NA_TOKEN=<access_token>
export NA_SERVER=http://localhost:8080
```

## 命令参考

### `na family` — 家庭管理

| 命令 | 说明 | 示例 |
|------|------|------|
| `family list` | 列出我的家庭 | `na family list` |
| `family create --name <name>` | 创建家庭 | `na family create --name "我的家"` |
| `family join --code <code>` | 通过邀请码加入 | `na family join --code ABC123` |

### `na task` — 任务管理

| 命令 | 说明 | 示例 |
|------|------|------|
| `task list --family-id <id>` | 列出任务 | `na task list --family-id abc` |
| `task create` | 创建任务 | 见下方调度类型 |
| `task todo --family-id <id>` | 列出待办 | `na task todo --family-id abc` |
| `task done --id <id> --status done` | 完成/跳过待办 | `na task done --id xyz --status done` |
| `task toggle --id <id> --enable false` | 启用/禁用任务 | `na task toggle --id xyz --enable false` |
| `task delete --id <id>` | 删除任务 | `na task delete --id xyz` |

### 调度类型

```bash
# 一次性
na task create --family-id <id> --name "取快递" --schedule once \
  --data '{"date":"2026-06-28","time":"18:00"}'

# 每天
na task create --family-id <id> --name "倒垃圾" --schedule daily \
  --data '{"time":"09:00"}'

# 每周 (1=周一...7=周日)
na task create --family-id <id> --name "周报" --schedule weekly \
  --data '{"days":[1,3,5],"time":"10:00"}'

# 每月
na task create --family-id <id> --name "大扫除" --schedule monthly \
  --data '{"days":[1,15],"time":"08:00"}'

# 间隔天数
na task create --family-id <id> --name "换床单" --schedule interval \
  --data '{"days":14,"time":"09:00"}'
```

## 全局选项

| 选项 | 环境变量 | 默认值 | 说明 |
|------|---------|--------|------|
| `--server` | `NA_SERVER` | `http://localhost:8080` | API 服务器地址 |
| `--token` | `NA_TOKEN` | — | 认证 Token |
| `--output` / `-o` | — | `table` | 输出格式: table/json/yaml |
| `--config` | — | `~/.na.yaml` | 配置文件路径 |
