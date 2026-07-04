/**
 * Task create/edit dialog (modal).
 * Appears when clicking "创建任务" or "编辑".
 */
import { Page, Locator } from '@playwright/test';

export class TaskEditorDialog {
  readonly dialog: Locator;
  readonly heading: Locator;
  readonly kindSelect: Locator;
  readonly nameInput: Locator;
  readonly scheduleTypeSelect: Locator;
  readonly scheduleTimeInput: Locator;
  readonly locationSelect: Locator;
  readonly groupSelect: Locator;
  readonly submitButton: Locator;
  readonly cancelButton: Locator;
  readonly closeButton: Locator;

  constructor(public page: Page) {
    this.dialog = page.locator('[role="dialog"]').first();
    this.heading = page.getByRole('heading', { name: /创建任务|编辑任务/ });
    this.kindSelect = page.getByRole('combobox').first();
    this.nameInput = page.getByPlaceholder('输入任务名称');
    this.scheduleTypeSelect = page.getByRole('combobox').nth(1);
    this.scheduleTimeInput = page.locator('input[type="time"]').first();
    this.locationSelect = page.getByRole('combobox').nth(2);
    this.groupSelect = page.getByRole('combobox').nth(3);
    this.submitButton = page.getByRole('button', { name: /^(创建|保存)$/ });
    this.cancelButton = page.getByRole('button', { name: '取消' });
    this.closeButton = page.getByRole('button', { name: '✕' });
  }

  /** Wait for dialog to appear. */
  async waitForOpen(): Promise<void> {
    await this.heading.waitFor({ state: 'visible', timeout: 5000 });
  }

  /** Select task kind from the dropdown. */
  async selectKind(kind: '任务' | '巡检' | '任务链'): Promise<void> {
    await this.kindSelect.click();
    await this.page.getByRole('option', { name: kind }).click();
  }

  /** Fill the task name. */
  async fillName(name: string): Promise<void> {
    await this.nameInput.fill(name);
  }

  /** Select schedule type. */
  async selectScheduleType(type: string): Promise<void> {
    await this.scheduleTypeSelect.click();
    await this.page.getByRole('option', { name: type }).click();
  }

  /** Fill all basic fields and submit. */
  async createSimpleTask(name: string, kind: '任务' | '巡检' | '任务链' = '任务'): Promise<void> {
    if (kind !== '任务') await this.selectKind(kind);
    await this.fillName(name);
    await this.submitButton.click();
  }
}
