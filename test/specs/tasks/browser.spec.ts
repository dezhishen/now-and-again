import { test, expect, type Page } from '@playwright/test';
import { db } from '../../utils/db';

function uniqueName(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
}

async function loginAndEnterFamily(page: Page) {
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
    await familyCards.first().click();
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
    expect(task?.created_by_kind).toBe('simple');

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
    expect(['simple', 'inspection']).toContain(task?.created_by_kind);

    await triggerTask(page, taskName);
    await openDashboard(page);

    await expectTodoVisible(page, taskName);

    const todos = db.getTodos(task.id);
    expect(todos.length).toBeGreaterThan(0);
    expect(todos.some((t: any) => t.status === 'pending')).toBeTruthy();
  });

  test('chain: 创建+触发，并验证 UI 与 DB', async ({ page }) => {
    const taskName = uniqueName('E2E-Chain');
    await createTask(page, 'chain', taskName);

    const task = db.findTask(taskName);
    expect(task).toBeTruthy();
    expect(task?.kind).toBe('chain');
    expect(task?.created_by_kind).toBe('chain');

    await triggerTask(page, taskName);
    await openDashboard(page);

    await expectTodoVisible(page, taskName);

    const todos = db.getTodos(task.id);
    expect(todos.length).toBeGreaterThan(0);
    expect(todos.some((t: any) => t.status === 'pending')).toBeTruthy();
  });
});
