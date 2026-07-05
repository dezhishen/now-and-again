# 家庭内快速使用模板创建任务（图文）

这份文档演示一个完整流程：

1. 进入家庭
2. 从模板快速创建任务
3. 生成待办并完成事项

适合首次上手或给家人做简短培训。

## 0. 环境准备

在仓库根目录执行：

```bash
# 可选：重置数据库，获得干净演示数据
make db-reset

# 启动后端和前端
make dev
```

默认地址：

- 前端：http://localhost:5173
- 后端：http://localhost:8080

## 1. 进入家庭后打开任务页

登录后进入家庭，在左侧导航选择“任务”。

![进入家庭任务页](images/template-quickstart/01-family-task-entry.png)

## 2. 从模板创建

点击“从模板创建”，打开模板选择弹窗。

![模板选择弹窗](images/template-quickstart/02-template-picker.png)

## 3. 填写模板参数并预览

以“每日打卡”为例，填写“打卡项目”（如：晨会打卡），然后点击“预览”。

![模板预览](images/template-quickstart/03-template-preview.png)

## 4. 填充到任务表单并创建

点击“填充到任务表单”后，系统会把模板渲染结果自动带入任务表单，确认后点击“创建”。

![任务表单已预填](images/template-quickstart/04-prefilled-task-form.png)

## 5. 生成待办并完成

在任务卡片点击“生成”，再到“首页”查看待办卡片并点击“完成”。

![待办进行中](images/template-quickstart/05-dashboard-pending-todo.png)

完成后，待办会从列表消失（或变为已完成状态）。

![待办完成后](images/template-quickstart/06-dashboard-after-done.png)

## 6. 复杂嵌套任务：任务链（Chain）

对于需要多步骤串联执行的场景，可以使用“任务链”类型。每一步可用不同的任务类型，前一步完成后自动激活下一步。

### 6.1 手动创建任务链

在任务页点击“创建任务”，类型选择“任务链”，填写任务名称后添加步骤：

- 每个步骤可以设置独立名称、类型（简单任务/巡检任务）
- 巡检型步骤还可以配置检查项与分支
- 步骤顺序可以上下拖拽调整

![创建任务链表单](images/template-quickstart/07-chain-create-form.png)

### 6.2 任务链卡片

创建后的任务链会显示橙色角标和步骤摘要（如 `水电检查 → 门窗检查`），便于快速识别。

![任务链卡片](images/template-quickstart/08-chain-task-card.png)

### 6.3 触发待办并推进

点击“生成”后，系统会按顺序生成待办。当前的待办卡片会显示：

- 任务链名称
- 当前步骤摘要
- ✅ 完成 / 📝 备注 / ⏭️ 跳过 / ⛔ 中断 四个操作按钮

![任务链待办进行中](images/template-quickstart/09-chain-todo.png)

### 6.4 自动推进下一步

完成当前步骤后，系统会自动激活下一步的待办：

1. 点击 ✅ 完成第一个步骤（如“水电检查”）
2. 第二个步骤（如“门窗检查”）会立即出现在待办列表中
3. 依次完成后任务链才算全部完成

![步骤推进后下一步待办自动出现](images/template-quickstart/10-chain-step-progress.png)

## 附：如何重新生成本文截图

本文截图由 Playwright 自动生成，用例文件：

- `test/specs/tasks/template-doc-capture.spec.ts`

执行命令：

```bash
cd test
npx playwright test tasks/template-doc-capture.spec.ts --project=chromium
```

截图输出目录：

- `doc/tutorial/images/template-quickstart/`
