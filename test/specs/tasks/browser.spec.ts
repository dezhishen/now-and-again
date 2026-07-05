import { test, expect, type Page } from '@playwright/test';
import { db } from '../../utils/db';

function uniqueName(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
}

async function ensureAdminAccount() {
  try {
    await fetch('http://localhost:8080/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        display_name: '管理员',
        username: 'admin',
        email: 'admin@local.test',
        password: '12345678',
      }),
    });
  } catch {
    // Ignore setup failure here; login step will surface actionable error.
  }
}

async function loginAndEnterFamily(page: Page) {
  await ensureAdminAccount();
  await page.goto('/login');
  await page.locator('[data-testid="login-username"]').fill('admin');
  await page.locator('[data-testid="login-password"]').fill('12345678');
  await page.locator('[data-testid="login-submit"]').click();
  await expect(page).not.toHaveURL(/\/login/);

  if (page.url().includes('/family')) return;

  await expect(page.getByRole('heading', { name: /家庭|Home/ })).toBeVisible();

  await page.waitForTimeout(400);

  let enterButtons = page.locator('[data-testid="family-enter-btn"]');
  let familyCards = page.locator('[data-testid="family-card"]');

  if (await enterButtons.count() === 0 && await familyCards.count() === 0) {
    const createToggle = page.locator('[data-testid="family-create-toggle"]').first();
    const createToggleFallback = page.getByRole('button', { name: /创建家庭|创建/ }).first();
    if (await createToggle.count() > 0 || await createToggleFallback.count() > 0) {
      if (await createToggle.count() > 0) await createToggle.click();
      else await createToggleFallback.click();

      const familyName = uniqueName('E2E家庭');
      const familyNameInput = page.locator('[data-testid="family-name-input"]').first();
      if (await familyNameInput.count() > 0) await familyNameInput.fill(familyName);
      else await page.getByPlaceholder(/家庭名称|Family name/).first().fill(familyName);

      const familyCreateSubmit = page.locator('[data-testid="family-create-submit"]').first();
      if (await familyCreateSubmit.count() > 0) await familyCreateSubmit.click();
      else await page.getByRole('button', { name: /创建|Create/ }).first().click();

      await page.waitForTimeout(800);
      enterButtons = page.locator('[data-testid="family-enter-btn"]');
      familyCards = page.locator('[data-testid="family-card"]');
    }
  }

  if (await enterButtons.count() > 0) {
    for (let i = 0; i < 3 && !page.url().includes('/family'); i++) {
      try {
        await page.locator('[data-testid="family-enter-btn"]').first().click({ timeout: 5000 });
      } catch {
        await page.waitForTimeout(300);
      }
    }
  }

  if (!page.url().includes('/family') && await familyCards.count() > 0) {
    for (let i = 0; i < 3 && !page.url().includes('/family'); i++) {
      try {
        await familyCards.first().click({ timeout: 5000 });
      } catch {
        await page.waitForTimeout(300);
      }
    }
  }

  if (!page.url().includes('/family') && await page.locator('[data-testid="family-enter-btn"]').count() > 0) {
    await page.locator('[data-testid="family-enter-btn"]').first().click();
  }

  await expect(page).toHaveURL(/\/family/);
}

async function openTasks(page: Page) {
  const taskNav = page.locator('[data-testid="family-nav-tasks"]').first();
  await expect(taskNav).toBeVisible();
  await taskNav.click();
  await expect(page.locator('[data-testid="task-create-btn"]')).toBeVisible();
}

async function openDashboard(page: Page) {
  const dashboardNav = page.locator('[data-testid="family-nav-dashboard"]').first();
  await expect(dashboardNav).toBeVisible();
  await dashboardNav.click();
  await expect(page.locator('[data-testid="todo-card"]').first().or(page.locator('text=暂无待办'))).toBeVisible();
}

async function expectTodoVisible(page: Page, taskName: string) {
  const todoByName = page.locator(`[data-testid="todo-card"][data-task-name="${taskName}"]`).first();
  if (await todoByName.count() === 0) {
    const showAllBtn = page.getByRole('button', { name: /显示全部|Show all|展开/ }).first();
    if (await showAllBtn.count() > 0) {
      await showAllBtn.click();
    }
  }
  await expect(page.locator(`[data-testid="todo-card"][data-task-name="${taskName}"]`).first()).toBeVisible();
}

async function createTask(page: Page, kind: 'simple' | 'inspection' | 'chain', taskName: string) {
  await page.locator('[data-testid="task-create-btn"]').click();
  await page.locator('[data-testid="task-kind"]').selectOption(kind);
  await page.locator('[data-testid="task-name"]').fill(taskName);

  if (kind === 'chain') {
    await page.locator('[data-testid="chain-add-step"]').click();
    await page.locator('[data-testid="subtask-name-input"]').first().fill('步骤一');
  }

  await page.locator('[data-testid="task-submit"]').click();

  const card = page.locator(`[data-testid="task-card"][data-task-name="${taskName}"]`).first();
  await expect(card).toBeVisible();
}

async function triggerTask(page: Page, taskName: string) {
  const card = page.locator(`[data-testid="task-card"][data-task-name="${taskName}"]`).first();
  await expect(card).toBeVisible();
  await card.locator('[data-testid="task-trigger-btn"]').click();
}

test.describe('任务创建与渲染（纯浏览器）', () => {
  test.beforeEach(async ({ page }) => {
    await loginAndEnterFamily(page);
    await openTasks(page);
  });

  test('simple: 创建+触发+完成，并验证 UI 与 DB', async ({ page }) => {
    const taskName = uniqueName('E2E-Simple');
    await createTask(page, 'simple', taskName);

    const task = db.findTask(taskName);
    expect(task).toBeTruthy();
    expect(task?.kind).toBe('simple');
    expect(task?.owner_kind).toBe('simple');

    await triggerTask(page, taskName);
    await openDashboard(page);

    await expectTodoVisible(page, taskName);
    const todoCard = page.locator(`[data-testid="todo-card"][data-task-name="${taskName}"]`).first();

    const todos = db.getTodos(task.id);
    expect(todos.length).toBeGreaterThan(0);
    expect(todos.some((t: any) => t.status === 'pending')).toBeTruthy();

    await todoCard.locator('[data-testid="todo-quick-done"]').click();
    await expect(todoCard).not.toBeVisible({ timeout: 10000 });

    const todosAfter = db.getTodos(task.id);
    expect(todosAfter.some((t: any) => t.status === 'done')).toBeTruthy();
  });

  test('inspection: 创建+触发，并验证 UI 与 DB', async ({ page }) => {
    const taskName = uniqueName('E2E-Inspect');
    await createTask(page, 'inspection', taskName);

    const task = db.findTask(taskName);
    expect(task).toBeTruthy();
    expect(task?.kind).toBe('inspection');
    expect(['simple', 'inspection']).toContain(task?.owner_kind);

    await triggerTask(page, taskName);
    await openDashboard(page);

    await expectTodoVisible(page, taskName);

    const todos = db.getTodos(task.id);
    expect(todos.length).toBeGreaterThan(0);
    expect(todos.some((t: any) => t.status === 'pending')).toBeTruthy();
  });

  test('inspection: 完成巡检后应写入结果并更新待办渲染', async ({ page }) => {
    const taskName = uniqueName('E2E-Inspect-Deep');

    await page.locator('[data-testid="task-create-btn"]').click();
    await page.locator('[data-testid="task-kind"]').selectOption('inspection');
    await page.locator('[data-testid="task-name"]').fill(taskName);
    await page.locator('[data-testid="check-item-add"]').click();
    await page.locator('[data-testid="check-item-name-input"]').first().fill('巡检项-A');
    await page.locator('[data-testid="task-submit"]').click();

    const task = db.findTask(taskName);
    expect(task).toBeTruthy();
    expect(task?.kind).toBe('inspection');

    const checkItems = db.getCheckItems(task.id);
    expect(checkItems.length).toBeGreaterThan(0);
    expect(checkItems.some((i: any) => i.name === '巡检项-A')).toBeTruthy();

    await triggerTask(page, taskName);
    await openDashboard(page);
    await expectTodoVisible(page, taskName);

    const todoCard = page.locator(`[data-testid="todo-card"][data-task-name="${taskName}"]`).first();
    await todoCard.locator('[data-testid="inspect-open-btn"]').click();
    await expect(page.getByRole('button', { name: '正常' }).first()).toBeVisible();

    await page.getByRole('button', { name: '正常' }).first().click();
    await page.locator('[data-testid="inspect-submit-btn"]').click();
    await expect(todoCard).not.toBeVisible({ timeout: 10000 });

    const results = db.getInspectionResults(task.id);
    expect(results.length).toBeGreaterThan(0);
    expect(results.some((r: any) => r.item_name === '巡检项-A' && r.branch_name === '正常')).toBeTruthy();

    const todosAfter = db.getTodos(task.id);
    expect(todosAfter.some((t: any) => t.status === 'done')).toBeTruthy();
  });

  test('chain: 创建+触发，并验证 UI 与 DB', async ({ page }) => {
    const taskName = uniqueName('E2E-Chain');
    await createTask(page, 'chain', taskName);

    const task = db.findTask(taskName);
    expect(task).toBeTruthy();
    expect(task?.kind).toBe('chain');
    expect(task?.owner_kind).toBe('chain');

    await triggerTask(page, taskName);
    await openDashboard(page);

    await expectTodoVisible(page, taskName);

    const todos = db.getTodos(task.id);
    expect(todos.length).toBeGreaterThan(0);
    expect(todos.some((t: any) => t.status === 'pending')).toBeTruthy();
  });

  test('chain: 子步骤任务应继承 owner_kind=chain', async ({ page }) => {
    const taskName = uniqueName('E2E-Chain-Owner');
    await createTask(page, 'chain', taskName);

    const root = db.findTask(taskName);
    expect(root).toBeTruthy();
    expect(root?.owner_kind).toBe('chain');

    const steps = db.getChainSteps(root.id);
    expect(steps.length).toBeGreaterThan(0);

    const firstStepTask = db.findTaskById(steps[0].child_task_id);
    expect(firstStepTask).toBeTruthy();
    expect(firstStepTask?.kind).toBe('simple');
    expect(firstStepTask?.owner_kind).toBe('chain');
  });

  test('chain: 完成第一步后应渲染并落库第二步待办', async ({ page }) => {
    const taskName = uniqueName('E2E-Chain-Next');

    await page.locator('[data-testid="task-create-btn"]').click();
    await page.locator('[data-testid="task-kind"]').selectOption('chain');
    await page.locator('[data-testid="task-name"]').fill(taskName);

    await page.locator('[data-testid="chain-add-step"]').click();
    await page.locator('[data-testid="subtask-name-input"]').nth(0).fill('链路-步骤1');
    await page.locator('[data-testid="chain-add-step"]').click();
    await page.locator('[data-testid="subtask-name-input"]').nth(1).fill('链路-步骤2');
    await page.locator('[data-testid="task-submit"]').click();

    const root = db.findTask(taskName);
    expect(root).toBeTruthy();
    const steps = db.getChainSteps(root.id);
    expect(steps.length).toBe(2);

    const step1 = db.findTaskById(steps[0].child_task_id);
    const step2 = db.findTaskById(steps[1].child_task_id);
    expect(step1).toBeTruthy();
    expect(step2).toBeTruthy();

    await triggerTask(page, taskName);
    await openDashboard(page);

    // Chain flow starts from root todo, then progresses to step1.
    await expectTodoVisible(page, taskName);
    const rootCard = page.locator(`[data-testid="todo-card"][data-task-name="${taskName}"]`).first();
    await rootCard.locator('[data-testid="chain-done-btn"]').click();
    await expect(rootCard).not.toBeVisible({ timeout: 10000 });

    await expectTodoVisible(page, step1.name);
    const step1Card = page.locator(`[data-testid="todo-card"][data-task-name="${step1.name}"]`).first();
    await step1Card.locator('[data-testid="todo-quick-done"]').click();
    await expect(step1Card).not.toBeVisible({ timeout: 10000 });

    await expectTodoVisible(page, step2.name);

    const step1Todos = db.getTodos(step1.id);
    const step2Todos = db.getTodos(step2.id);
    expect(step1Todos.some((t: any) => t.status === 'done')).toBeTruthy();
    expect(step2Todos.some((t: any) => t.status === 'pending')).toBeTruthy();
  });

  test('chain: 编辑 step(inspection) 时应渲染已保存检查项', async ({ page }) => {
    const taskName = uniqueName('E2E-Chain-Inspect-Edit');

    await page.locator('[data-testid="task-create-btn"]').click();
    await page.locator('[data-testid="task-kind"]').selectOption('chain');
    await page.locator('[data-testid="task-name"]').fill(taskName);

    await page.locator('[data-testid="chain-add-step"]').click();
    await page.locator('[data-testid="subtask-name-input"]').first().fill('chain-step-1');
    await page.locator('[data-testid="subtask-config-btn"]').first().click();

    await page.locator('[data-testid="subtask-kind-select"]').first().selectOption('inspection');
    await expect(page.locator('[data-testid="check-item-add"]')).toBeVisible();
    await page.locator('[data-testid="check-item-add"]').click();
    await page.locator('[data-testid="check-item-name-input"]').first().fill('检查项-A');
    await page.locator('[data-testid="subtask-confirm"]').click();

    await page.locator('[data-testid="task-submit"]').click();
    const card = page.locator(`[data-testid="task-card"][data-task-name="${taskName}"]`).first();
    await expect(card).toBeVisible();

    await card.locator('[data-testid="task-edit-btn"]').click();
    await page.locator('[data-testid="subtask-config-btn"]').first().click();

    const itemInput = page.locator('[data-testid="check-item-name-input"]').first();
    await expect(itemInput).toBeVisible();
    await expect(itemInput).toHaveValue('检查项-A');
  });
});
