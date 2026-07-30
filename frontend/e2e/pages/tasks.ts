/**
 * Task list page (`/family/tasks`).
 * Shows all tasks with create/edit/delete controls.
 */
import { Page, Locator } from '@playwright/test';
import { BasePage } from './base';

export class TaskListPage extends BasePage {
  // ─── Buttons ─────────────────────────────────────────────────────
  get createTaskButton(): Locator {
    return this.page.getByRole('button', { name: /创建任务/ });
  }

  get createFromTemplateButton(): Locator {
    return this.page.getByRole('button', { name: /从模板创建/ });
  }

  get archivedButton(): Locator {
    return this.page.getByRole('button', { name: /已归档/ });
  }

  get disabledButton(): Locator {
    return this.page.getByRole('button', { name: /已禁用/ });
  }

  // ─── Task Cards ──────────────────────────────────────────────────
  get emptyState(): Locator {
    return this.page.getByText('暂无任务');
  }

  /** Get a task card by name. */
  taskCard(name: string): Locator {
    return this.page.locator('button', { hasText: name }).first();
  }

  /** Click "生成" to create a todo for the given task. */
  async generateTodo(taskName: string): Promise<void> {
    const card = this.taskCard(taskName);
    // The "生成" button is near the task card in its action row
    const generateBtn = this.page.getByRole('button', { name: '生成' }).first();
    await generateBtn.click();
    await this.page.waitForTimeout(1000);
  }

  /** Click "编辑" on a task card. */
  async editTask(taskName: string): Promise<void> {
    const row = this.taskCard(taskName).locator('..');
    await row.getByRole('button', { name: '编辑' }).click();
  }

  /** Click "删除" on a task card. */
  async deleteTask(taskName: string): Promise<void> {
    const row = this.taskCard(taskName).locator('..');
    await row.getByRole('button', { name: '删除' }).click();
  }

  /** Navigate to task list. */
  async goto(): Promise<void> {
    await this.navTasks.click();
    await this.page.waitForURL(/\/family\/tasks/);
    await this.waitForLoad();
  }
}
