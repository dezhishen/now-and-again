/**
 * Family dashboard page (`/family`).
 * Shows todo cards and overview.
 */
import { Page, Locator } from '@playwright/test';
import { BasePage } from './base';

export class DashboardPage extends BasePage {
  // ─── Tabs ────────────────────────────────────────────────────────
  get todoTab(): Locator {
    return this.page.getByRole('button', { name: /待办/ });
  }

  get overviewTab(): Locator {
    return this.page.getByRole('button', { name: /概览/ });
  }

  // ─── Todo Cards ──────────────────────────────────────────────────
  get emptyState(): Locator {
    return this.page.getByText('暂无待办事项');
  }

  /** Get a todo card by task name. */
  todoCard(taskName: string): Locator {
    return this.page.locator('[class*="todo"]').filter({ hasText: taskName }).first();
  }

  /** Click "快速完成" on a todo card by task name. */
  async quickComplete(taskName: string): Promise<void> {
    const card = this.todoCard(taskName);
    await card.getByRole('button', { name: /快速完成/ }).click();
    await this.waitForLoad();
  }

  /** Click "备注" on a todo card. */
  async remarkTodo(taskName: string): Promise<void> {
    const card = this.todoCard(taskName);
    await card.getByRole('button', { name: /备注/ }).click();
  }

  /** Click "跳过" on a todo card. */
  async skipTodo(taskName: string): Promise<void> {
    const card = this.todoCard(taskName);
    await card.getByRole('button', { name: /跳过/ }).click();
    await this.waitForLoad();
  }

  /** Assert a todo card with the given task name is visible. */
  async expectTodoVisible(taskName: string): Promise<void> {
    await this.todoCard(taskName).waitFor({ state: 'visible', timeout: 5000 });
  }
}
