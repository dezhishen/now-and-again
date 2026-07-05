# CLI 命令行工具使用指南

> `na` 是 Now & Again 的命令行工具。部署到 AI 助手后，你只需要用自然语言说话，AI 会帮你调用它。

---

## 1. 安装（由管理员一次性完成）

管理员在服务器上安装 `na`，家庭其他成员**无需安装任何东西**。

```bash
curl -fsSL https://github.com/dezhishen/now-and-again/releases/latest/download/na_linux_amd64.tar.gz | tar xz
sudo mv na /usr/local/bin/
```

> 支持平台：`linux_amd64` · `linux_arm64` · `darwin_amd64` · `darwin_arm64` · `windows_amd64`
> 所有版本：[GitHub Releases](https://github.com/dezhishen/now-and-again/releases)

## 2. 初始化（管理员一次性完成）

管理员为 AI 助手完成配置：

```bash
na init -u admin -p 12345678 --server http://localhost:8080
```

完成后配置保存在 `~/.na.yaml`，AI 助手之后的所有调用无需再传凭据。

## 3. 接入 AI 助手（OpenClaw / 其他）

将 `na-tools.yaml` 配置到 AI 框架，AI 就能直接调用 `na` 来完成家庭事务。

```yaml
# 添加到 AI 框架的工具配置中
tools:
  - name: na
    path: /usr/local/bin/na
    skills: cli/skills/na-tools.yaml
```

配置完成后，家人只需要用自然语言说话即可。

---

## 4. 日常使用（通过 AI 对话）

以下是家庭成员通过 AI 助手完成日常事务的完整场景。
**所有命令均由 AI 自动执行，用户只需要说话。**

### 查看今天的待办

> 👤 **家人**：帮我看看今天有什么事要做
>
> 🤖 **AI**：`na todo list` → 列出 3 项：倒垃圾、遛狗、取快递

### 完成待办

> 👤 **家人**：倒垃圾已经完成了
>
> 🤖 **AI**：`na todo done --id <短ID>` → ✅ 已完成

### 创建每天的任务

> 👤 **家人**：帮我创建一个每天晚上8点倒垃圾的任务
>
> 🤖 **AI**：`na task create --name "倒垃圾" --schedule daily --data '{"time":"20:00"}'` → ✅ 已创建

### 创建每周的任务

> 👤 **家人**：每周一、三、五早上7点遛狗
>
> 🤖 **AI**：`na task create --name "遛狗" --schedule weekly --data '{"days":[1,3,5],"time":"07:00"}'` → ✅ 已创建

### 创建每月的任务

> 👤 **家人**：每月1号和15号大扫除
>
> 🤖 **AI**：`na task create --name "大扫除" --schedule monthly --data '{"days":[1,15],"time":"09:00"}'` → ✅ 已创建

### 从模板创建

> 👤 **家人**：每周打扫一下客厅
>
> 🤖 **AI**：`na template use --code weekly_cleaning --params '{"area_name":"客厅"}'` → ✅ 已创建

### 创建家庭并邀请家人

> 👤 **家人**：帮我创建一个新家庭叫"温馨小家"
>
> 🤖 **AI**：`na family create --name "温馨小家"` → 邀请码: ABCD1234

拿到邀请码后，对方的管理员只需运行一次 `na family join --code ABCD1234`，之后对方全家人也能通过 AI 对话管理这个家庭。

### 一次说完多件事

> 👤 **家人**：帮我创建三个任务——每天早上8点喂猫，每周六下午3点洗澡，每月1号交电费
>
> 🤖 **AI**：依次调用 3 次 `na task create`，全部完成 ✅

---

## 5. 工具速查（管理员参考）

管理员配置 AI 时需要的命令一览。**家庭成员无需了解这些。**

| 场景 | AI 调用的命令 |
|------|-------------|
| 查看今日待办 | `na todo list` 或 `na daily` |
| 完成待办 | `na todo done --id <短ID> [--remark <备注>]` |
| 跳过待办 | `na todo skip --id <短ID>` |
| 查看所有任务 | `na task list` |
| 创建每天任务 | `na task create --name <名称> --schedule daily --data '{"time":"09:00"}'` |
| 创建每周任务 | `na task create --name <名称> --schedule weekly --data '{"days":[1,3,5],"time":"10:00"}'` |
| 创建每月任务 | `na task create --name <名称> --schedule monthly --data '{"days":[1,15],"time":"08:00"}'` |
| 创建间隔任务 | `na task create --name <名称> --schedule interval --data '{"days":14,"time":"12:00"}'` |
| 创建一次性任务 | `na task create --name <名称> --schedule once --data '{"date":"2026-12-31","time":"18:00"}'` |
| 从模板创建 | `na template use --code <模板代码> --params '<JSON>'` |
| 查看可用模板 | `na template list` |
| 创建家庭 | `na family create --name <名称>` |
| 加入家庭 | `na family join --code <邀请码>` |
| 查看我的家庭 | `na family list` |
| 切换活跃家庭 | `na family select` |
| 删除任务 | `na task delete --id <ID>` |

> 调度类型说明：
> - `daily` — 每天指定时间
> - `weekly` — 指定星期（1=周一 ... 7=周日）
> - `monthly` — 指定日期（1-31）
> - `interval` — 每隔 N 天
> - `once` — 一次性任务
>
> 待办 ID 支持短前缀（3位即可），不需要完整 UUID。
> 所有命令自动使用活跃家庭，无需传 `--family-id`。
> 配置存储在 `~/.na.yaml`，不依赖环境变量。
