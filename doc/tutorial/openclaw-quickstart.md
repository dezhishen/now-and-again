# OpenClaw / AI 助手接入指南

> 将 Now & Again 接入 OpenClaw 等 AI 助手，全家只需用自然语言管理事务。

---

## 1. 前置准备（管理员一次性完成）

### 1.1 安装 `na` 工具

```bash
curl -fsSL https://github.com/dezhishen/now-and-again/releases/latest/download/na_linux_amd64.tar.gz | tar xz
sudo mv na /usr/local/bin/
```

### 1.2 初始化配置

```bash
na init -u admin -p 12345678 --server http://localhost:8080
```

完成后配置保存在 `~/.na.yaml`，后续无需再传凭据。

---

## 2. 接入 OpenClaw

### 2.1 配置工具

在 OpenClaw 的工具目录中添加 `na.yaml`：

```yaml
# ~/.openclaw/tools/na.yaml
tools:
  - name: na
    path: /usr/local/bin/na
    skills: cli/skills/na-tools.yaml
```

重启 OpenClaw 即可生效。

---

## 3. 创建家庭

通过 AI 对话完成。管理员一次性创建好家庭，之后全家人都在同一个家庭空间内协作。

> 👤 **你**：帮我创建一个家庭，叫"温馨小家"
>
> 🤖 **AI**：`na family create --name "温馨小家"` → 已创建，邀请码: ABCD1234

如果已有家庭，AI 会选择当前活跃的家庭：

> 👤 **你**：我现在有哪些家庭？
>
> 🤖 **AI**：`na family list` → 2 个家庭：温馨小家 ★、公司项目

---

## 4. 查看可用模板

> 👤 **你**：有哪些任务模板可以用？
>
> 🤖 **AI**：`na template list` → 8 个模板：每日巡检、每日打卡、日常大扫除、绿植养护、定期缴费检查、换洗床品、冰箱清理、垃圾分类

---

## 5. 从模板创建任务

> 👤 **你**：帮我每周打扫客厅
>
> 🤖 **AI**：`na template use --code weekly_cleaning --params '{"area_name":"客厅"}'` → ✅ 已从模板创建: 客厅 - 日常大扫除

> 👤 **你**：每天打卡晨会
>
> 🤖 **AI**：`na template use --code simple_daily_check --params '{"check_item":"晨会"}'` → ✅ 已从模板创建: 每日晨会

---

## 6. 手动创建任务

根据你的描述，AI 自动选择合适的调度类型：

> 👤 **你**：每天晚上8点倒垃圾
>
> 🤖 **AI**：`na task create --name "倒垃圾" --schedule daily --data '{"time":"20:00"}'` → ✅

> 👤 **你**：每周一、三、五早上7点遛狗
>
> 🤖 **AI**：`na task create --name "遛狗" --schedule weekly --data '{"days":[1,3,5],"time":"07:00"}'` → ✅

> 👤 **你**：每月1号和15号大扫除，上午9点
>
> 🤖 **AI**：`na task create --name "大扫除" --schedule monthly --data '{"days":[1,15],"time":"09:00"}'` → ✅

> 👤 **你**：每两周换一次床单
>
> 🤖 **AI**：`na task create --name "换床单" --schedule interval --data '{"days":14,"time":"10:00"}'` → ✅

> 👤 **你**：12月31号下午6点取快递，就这一次
>
> 🤖 **AI**：`na task create --name "取快递" --schedule once --data '{"date":"2026-12-31","time":"18:00"}'` → ✅

### 一次说完多个任务

> 👤 **你**：帮我创建三个任务——每天早上8点喂猫，每周六下午3点给猫洗澡，每月1号驱虫
>
> 🤖 **AI**：依次创建，全部完成 ✅

---

## 7. 日常待办

### 查看今天的待办

> 👤 **你**：今天有什么事要做？
>
> 🤖 **AI**：`na todo list` →
> ```
> 📋 待办 (3项):
>   1. [a1b2c3] 每日倒垃圾  ⏰ 07-05 20:00
>   2. [d4e5f6] 遛狗        ⏰ 07-05 07:00
>   3. [e7f8g9] 每日大扫除  ⏰ 07-05 09:00
> ```

### 完成待办

> 👤 **你**：倒垃圾已完成，备注"分类投放"
>
> 🤖 **AI**：`na todo done --id a1b2 --remark "分类投放"` → ✅ 已完成: 每日倒垃圾

> 👤 **你**：遛狗今天跳过吧
>
> 🤖 **AI**：`na todo skip --id d4e5` → ⏭️ 已跳过

---

## 8. 查看和管理任务

> 👤 **你**：我现在有哪些任务？
>
> 🤖 **AI**：`na task list` → 列出所有任务及调度信息

> 👤 **你**：删除取快递的任务
>
> 🤖 **AI**：`na task delete --id a3b4` → ✅

---

## 9. 邀请家人

管理员创建家庭后，拿到邀请码。其他家庭的管理员运行一次即可接入：

```bash
na family join --code ABCD1234
```

之后两个家庭的所有成员都能通过自己的 AI 助手管理事务。

---

## 附：AI 调用的命令一览

家庭成员无需了解以下内容，仅供管理员调试参考。

| 场景 | AI 实际执行的命令 |
|------|------------------|
| 查看今日待办 | `na todo list` |
| 完成待办 | `na todo done --id <短ID> [--remark <备注>]` |
| 跳过待办 | `na todo skip --id <短ID>` |
| 查看任务 | `na task list` |
| 每天 | `na task create --name <名> --schedule daily --data '{"time":"09:00"}'` |
| 每周 | `na task create --name <名> --schedule weekly --data '{"days":[1,3,5],"time":"10:00"}'` |
| 每月 | `na task create --name <名> --schedule monthly --data '{"days":[1,15],"time":"08:00"}'` |
| 间隔 | `na task create --name <名> --schedule interval --data '{"days":14,"time":"12:00"}'` |
| 一次性 | `na task create --name <名> --schedule once --data '{"date":"2026-12-31","time":"18:00"}'` |
| 从模板 | `na template use --code <模板代码> --params '<JSON>'` |
| 查看模板 | `na template list` |
| 创建家庭 | `na family create --name <名称>` |
| 加入家庭 | `na family join --code <邀请码>` |

> 调度类型：`daily` / `weekly` / `monthly` / `interval` / `once`
> 待办 ID 支持短前缀（3位即可）
