# Now & Again CLI (`na`)

## 安装

```bash
cd cli && go build -o na . && sudo mv na /usr/local/bin/
```

## 初始化（仅需一次）

```bash
# 通过用户名密码登录，自动创建 API Key 并保存配置
na init -u admin -p 12345678

# 或使用已有的 API Key
na init --key na_key_xxxxxxxxxx
```

配置自动保存到 `~/.na.yaml`，之后所有命令无需再传凭据。

## 命令参考

### `na family` — 家庭管理

| 命令 | 说明 |
|------|------|
| `family list` | 列出我的家庭 |
| `family create --name <name>` | 创建新家庭（自动设为活跃） |
| `family join --code <code>` | 通过邀请码加入家庭 |

### `na task` — 任务管理（使用活跃家庭）

| 命令 | 说明 |
|------|------|
| `task list` | 列出活跃家庭的所有任务 |
| `task create --name <name> --schedule <type> --data <json>` | 创建任务 |
| `task delete --id <id>` | 删除任务 |

### `na todo` — 待办管理（使用活跃家庭）

| 命令 | 说明 |
|------|------|
| `todo list` | 列出所有待办 |
| `todo done --id <id> [--remark <text>]` | 完成待办（可附带备注） |
| `todo skip --id <id>` | 跳过待办 |

### `na template` — 模板管理

| 命令 | 说明 |
|------|------|
| `template list` | 列出可用模板 |
| `template use --code <code> --params <json>` | 从模板创建任务 |

### 调度类型示例

```bash
# 每天
na task create --name "倒垃圾" --schedule daily --data '{"time":"09:00"}'

# 每周一三五
na task create --name "周报" --schedule weekly --data '{"days":[1,3,5],"time":"10:00"}'

# 每月1号和15号
na task create --name "大扫除" --schedule monthly --data '{"days":[1,15],"time":"08:00"}'

# 每14天
na task create --name "换床单" --schedule interval --data '{"days":14,"time":"09:00"}'

# 一次性
na task create --name "取快递" --schedule once --data '{"date":"2026-06-28","time":"18:00"}'
```

### 从模板创建

```bash
# 列出模板
na template list

# 使用模板
na template use --code weekly_cleaning --params '{"area_name":"客厅"}'
```

## 全局选项

| 选项 | 说明 |
|------|------|
| `-c, --config <path>` | 指定配置文件路径（默认 `~/.na.yaml`） |
| `--server <url>` | 服务器地址（覆盖配置文件） |
| `--token <key>` | API Key（覆盖配置文件） |
| `-o, --output table\|json\|yaml` | 输出格式（默认 table） |

### 使用自定义配置文件

```bash
# 指定配置文件
na -c /path/to/custom-config.yaml task list

# 初始化到自定义配置文件
na -c /path/to/custom-config.yaml init -u admin -p 12345678
```

## 时区说明

所有时间在 CLI 中输入时使用**本地时间**，SDK 自动转换为 UTC 再发送给服务端。
显示时（如 `todo list` 的截止时间）自动从 UTC 转回本地时区。
无需手动处理时区转换。当前时区从操作系统自动检测。

> 默认配置存储在 `~/.na.yaml`，可通过 `-c` 指定其他路径。不需要设置环境变量。
