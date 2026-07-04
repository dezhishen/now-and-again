/**
 * Family management page (`/families`).
 */
import { Page, Locator } from '@playwright/test';

export class FamilyManagePage {
  readonly heading: Locator;
  readonly createButton: Locator;
  readonly cancelButton: Locator;
  readonly nameInput: Locator;
  readonly submitButton: Locator;
  readonly inviteInput: Locator;
  readonly joinButton: Locator;

  constructor(public page: Page) {
    this.heading = page.getByRole('heading', { name: '家庭管理' });
    this.createButton = page.getByRole('button', { name: /创建家庭/ });
    this.cancelButton = page.getByRole('button', { name: '取消' });
    this.nameInput = page.getByPlaceholder('家庭名称');
    this.submitButton = page.getByRole('button', { name: '创建' });
    this.inviteInput = page.getByPlaceholder('输入邀请码加入');
    this.joinButton = page.getByRole('button', { name: '加入' });
  }

  async goto(): Promise<void> {
    await this.page.goto('/families');
    await this.heading.waitFor({ state: 'visible' });
  }

  /** Click "创建家庭", fill name, submit. */
  async createFamily(name: string): Promise<void> {
    await this.createButton.click();
    await this.nameInput.fill(name);
    await this.submitButton.click();
  }

  /** Click "进入" on the first family card. */
  async enterFamily(name: string): Promise<void> {
    const card = this.page.locator('button', { hasText: name }).first();
    await card.click();
    // Or click the enter button
    const enterBtn = this.page.getByRole('button', { name: /进入/ }).first();
    if (await enterBtn.isVisible()) {
      await enterBtn.click();
    }
    await this.page.waitForURL(/\/family/, { timeout: 10000 });
  }
}
