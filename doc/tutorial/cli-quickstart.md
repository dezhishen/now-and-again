# CLI 命令行工具使用指南

> `na` 是 Now & Again 的命令行工具。一键初始化、无需记 UUID、交互式处理待办。

---

## 1. 安装

### 从 GitHub Releases 安装（推荐）

```bash
# Linux / macOS (amd64)
curl -fsSL https://github.com/dezhishen/now-and-again/releases/latest/download/na_linux_amd64.tar.gz | tar xz
sudo mv na /usr/local/bin/

# macOS (Apple Silicon)
curl -fsSL https://github.com/dezhishen/now-and-again/releases/latest/download/na_darwin_arm64.tar.gz | tar xz
sudo mv na /usr/local/bin/

# 指定版本
curl -fsSL https://github.com/dezhishen/now-and-again/releases/download/v0.0.2-beta-001/na_linux_amd64.tar.gz | tar xz
```

> 📦 所有版本：[GitHub Releases](https://github.com/dezhishen/now-and-again/releases)
>
> 支持平台：`linux_amd64` · `linux_arm64` · `darwin_amd64` · `darwin_arm64` · `windows_amd64`

验证安装：

```bash
na --help
```

输出：

```
Now & Again CLI (na) — 家庭事务管理，一行命令搞定。

快速上手:
  na init -u <用户名> -p <密码>   一次性初始化
  na daily                         查看并处理今天的待办
  na template list                  查看可用模板
  na task list                      查看任务
```

## 2. 初始化（仅需一次）

```bash
na init -u admin -p 12345678
```

输出：

```
✓ Initialized successfully
  Server: http://localhost:8080
  Family: 我的家
```

这一步完成了：
- 🔐 登录验证 → 自动创建 API Key
- 🏠 自动选择第一个家庭
- 💾 配置保存到 `~/.na.yaml`

之后所有命令**无需再传凭据**。

> 如果服务器不在本地：
> ```bash
> na init -u admin -p 12345678 --server https://your-server.com
> ```

## 3. 添加到 OpenClaw / AI Agent

将 `na-tools.yaml` 配置到你的 AI 框架中，AI 就能直接调用 `na` 命令。

### OpenClaw 配置

```yaml
# ~/.openclaw/tools/na.yaml
tools:
  - name: na
    path: /usr/local/bin/na
    skills: cli/skills/na-tools.yaml
```

### 对话示例

> **你**：帮我看看今天有什么要做的
>
> **AI**：`na todo list` → 📋 待办 (3项): 1. 倒垃圾  2. 遛狗  3. 取快递
>
> **你**：完成倒垃圾
>
> **AI**：`na todo done --id <短ID>` → ✅ 已完成: 每日倒垃圾
>
> **你**：创建每天上午9点喂猫的任务
>
> **AI**：`na task create --name "喂猫" --schedule daily --data '{"time":"09:00"}'` → ✅ 已创建

## 4. 创建第一个家庭

```bash
# 如果还没有家庭
na family create --name "我的家"
```

输出：

```
Family created: 我的家 (invite code: ABCD1234)
```

将邀请码发给家人，对方用 `na family join --code ABCD1234` 即可加入。

## 5. 从模板创建任务

```bash
# 查看可用模板
na template list

# 使用"每日打卡"模板
na template use --code simple_daily_check --params '{"check_item":"晨会"}'
```

输出：

```
📋 可用模板 (8个):
  🔍 daily_inspection    inspection  每日巡检
  ✅ simple_daily_check  simple      每日打卡
  🧹 weekly_cleaning     inspection  日常大扫除
  ...

✅ 已从模板创建: 每日晨会 (abc123)
```

## 6. 手动创建任务

支持 5 种调度类型：

```bash
# 每天
na task create --name "倒垃圾" --schedule daily --data '{"time":"20:00"}'

# 每周一三五
na task create --name "遛狗" --schedule weekly --data '{"days":[1,3,5],"time":"07:00"}'

# 每月1号和15号
na task create --name "大扫除" --schedule monthly --data '{"days":[1,15],"time":"09:00"}'

# 每14天
na task create --name "换床单" --schedule interval --data '{"days":14,"time":"10:00"}'

# 一次性
na task create --name "取快递" --schedule once --data '{"date":"2026-12-31","time":"18:00"}'

# 查看任务
na task list
```

输出：

```
📋 任务列表 (5项):
  ✅ [a1b2c3] 倒垃圾      daily
  ✅ [d4e5f6] 遛狗        weekly
  ✅ [a7b8c9] 大扫除      monthly
  ✅ [d0e1f2] 换床单      interval
  ✅ [a3b4c5] 取快递      once
```

## 7. 一键日常（推荐）

`na daily` 是最常用的命令，交互式处理今天的所有待办：

```bash
na daily
```

输出：

```
🏠 已自动选择家庭: 我的家

📋 待办 (3项):

   1. 每日倒垃圾  ⏰ 07-05 20:00
   2. 遛狗        ⏰ 07-05 07:00
   3. 每日大扫除  ⏰ 07-05 09:00

💡 输入编号完成 (1-3)，s 跳过当前，q 退出

→ 1
📝 备注 (回车跳过): 已完成
✅ 已完成: 每日倒垃圾
   📝 已完成

→ s

→ 3
📝 备注 (回车跳过):
✅ 已完成: 每日大扫除

→ q
👋
```

### 快捷完成

```bash
# 直接完成第 3 项
na daily --done 3
```

## 8. 完成待办

```bash
# 查看待办（显示短ID方便复制）
na todo list

# 用短ID完成（3位即可，不需要完整UUID）
na todo done --id a1b2

# 完成并添加备注
na todo done --id a1b2 --remark "已完成✅"

# 跳过某个待办
na todo skip --id c3d4
```

输出：

```
📋 待办 (2项):

   1. [a1b2c3] 每日倒垃圾  ⏰ 07-05 20:00
   2. [d4e5f6] 遛狗        ⏰ 07-05 07:00

💡 使用 na todo done --id abc123 完成（支持短ID，至少3位）
💡 或使用 na daily 交互式处理
```

## 9. 删除/管理任务

```bash
# 删除任务
na task delete --id abc123
```

## 10. 命令速查表

| 命令 | 说明 |
|------|------|
| `na init -u <用户> -p <密码>` | 一次性初始化 |
| `na daily` | 交互式处理今日待办（⭐ 最常用） |
| `na daily --done 3` | 直接完成第 3 项待办 |
| `na family create --name <名称>` | 创建家庭 |
| `na family list` | 查看我的家庭 |
| `na family join --code <邀请码>` | 加入家庭 |
| `na template list` | 查看可用模板 |
| `na template use --code <码> --params <JSON>` | 从模板创建任务 |
| `na task list` | 查看所有任务 |
| `na task create --name <名> --schedule <类型> --data <JSON>` | 创建任务 |
| `na task delete --id <ID>` | 删除任务 |
| `na todo list` | 查看待办 |
| `na todo done --id <短ID> [--remark <备注>]` | 完成待办 |
| `na todo skip --id <短ID>` | 跳过待办 |

> 所有命令自动使用活跃家庭，无需传 `--family-id`。
> 配置存储在 `~/.na.yaml`，不需要环境变量。
