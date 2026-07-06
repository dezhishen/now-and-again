/**
 * 任务浏览器测试：创建、嵌套链、待办流程
 */
import { test, expect, type Page } from '@playwright/test';
import { db } from '../../utils/db';

function uname(p: string) { return `${p}-${Date.now()}`; }

async function loginEnter(page: Page) {
  await page.goto('/login');
  await page.locator('[data-testid="login-username"]').fill('admin');
  await page.locator('[data-testid="login-password"]').fill('12345678');
  await page.locator('[data-testid="login-submit"]').click();
  await page.waitForTimeout(3000);
  const enter = page.locator('[data-testid="family-enter-btn"]').first();
  if (await enter.count() > 0) { await enter.click(); await page.waitForTimeout(2000); }
}

test.describe('任务创建', () => {
  test.beforeEach(async ({ page }) => { await loginEnter(page); });

  test('simple: 创建→DB 验证', async ({ page }) => {
    const name = uname('E2E-Simple');
    await page.locator('[data-testid="family-nav-tasks"]').click();
    await page.locator('[data-testid="task-create-btn"]').click();
    await page.locator('[data-testid="task-name"]').fill(name);
    await page.locator('[data-testid="task-submit"]').click();
    await expect(page.locator(`[data-testid="task-card"][data-task-name="${name}"]`)).toBeVisible({ timeout: 10000 });
    const t = db.findTask(name);
    expect(t).toBeTruthy();
    expect(t.kind).toBe('simple');
  });

  test('inspection: 创建→DB 验证', async ({ page }) => {
    const name = uname('E2E-Inspect');
    await page.locator('[data-testid="family-nav-tasks"]').click();
    await page.locator('[data-testid="task-create-btn"]').click();
    await page.locator('[data-testid="task-kind"]').selectOption('inspection');
    await page.locator('[data-testid="task-name"]').fill(name);
    await page.locator('[data-testid="task-submit"]').click();
    const t = db.findTask(name);
    expect(t?.kind).toBe('inspection');
  });

  test('chain: 创建步骤→DB 验证子任务', async ({ page }) => {
    const name = uname('E2E-Chain');
    await page.locator('[data-testid="family-nav-tasks"]').click();
    await page.locator('[data-testid="task-create-btn"]').click();
    await page.locator('[data-testid="task-kind"]').selectOption('chain');
    await page.locator('[data-testid="task-name"]').fill(name);
    await page.locator('[data-testid="chain-add-step"]').click();
    await page.locator('[data-testid="subtask-name-input"]').first().fill('步骤1');
    await page.locator('[data-testid="chain-add-step"]').click();
    await page.locator('[data-testid="subtask-name-input"]').nth(1).fill('步骤2');
    await page.locator('[data-testid="task-submit"]').click();
    const t = db.findTask(name);
    expect(t?.kind).toBe('chain');
    const steps = db.getChainSteps(t.id);
    expect(steps.length).toBe(2);
    const sub1 = db.findTaskById(steps[0].child_task_id);
    const sub2 = db.findTaskById(steps[1].child_task_id);
    expect(sub1).toBeTruthy();
    expect(sub2).toBeTruthy();
    expect(sub1?.owner_kind).toBe('chain');
    expect(sub2?.owner_kind).toBe('chain');
  });
});

test.describe('待办流程', () => {
  test.beforeEach(async ({ page }) => { await loginEnter(page); });

  test('simple: 触发→完成→DB 验证 done', async ({ page }) => {
    const name = uname('E2E-Todo');
    await page.locator('[data-testid="family-nav-tasks"]').click();
    await page.locator('[data-testid="task-create-btn"]').click();
    await page.locator('[data-testid="task-name"]').fill(name);
    await page.locator('[data-testid="task-submit"]').click();
    await page.waitForTimeout(500);

    const t = db.findTask(name);
    expect(t).toBeTruthy();

    // Trigger from task card
    const taskCard = page.locator(`[data-testid="task-card"][data-task-name="${name}"]`).first();
    await taskCard.locator('[data-testid="task-trigger-btn"]').click();
    await page.waitForTimeout(1500);

    // Verify todo was created in DB
    const todos = db.getTodos(t.id);
    expect(todos.some((x: any) => x.status === 'pending')).toBeTruthy();
  });
});
