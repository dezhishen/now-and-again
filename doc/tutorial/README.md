# 教程

> 图文结合的实践指南，帮助你快速上手 Now & Again 的各项功能。

---

## 📋 教程列表

| 教程 | 说明 | 截图 |
|------|------|------|
| [从模板创建任务并完成待办](template-quickstart.md) | 基于标准模板快速创建任务，支持调步器自动生成待办，含任务链多步骤推进 | 10 张 |
| [家庭管理入门](family-quickstart.md) | 创建家庭、首页概况、邀请成员、角色权限管理 | 3 张 |
| [巡检任务使用指南](inspection-quickstart.md) | 创建巡检、检查项与分支判定、异常跟进、完成巡检 | 5 张 |
| [日历大屏与 ICS 订阅](calendar-quickstart.md) | 日历视图、全屏大屏、ICS 订阅源配置、嵌入外部网页 | 2 张 |

## 📸 如何重新生成本文截图

所有教程截图由 Playwright 自动生成，用例文件：

- 模板与任务链截图：`test/specs/tasks/template-doc-capture.spec.ts`
- 家庭/巡检/日历截图：`test/specs/tutorial-capture.spec.ts`

执行命令：

```bash
cd test
npx playwright test template-doc-capture.spec.ts tutorial-capture.spec.ts --project=chromium
```

截图输出目录：

- `doc/tutorial/images/`
