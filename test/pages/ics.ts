/**
 * ICS subscription page (family tab: /family/*\/ics).
 */
import { Page, Locator } from '@playwright/test';
import { BasePage } from './base';

export class IcsPage extends BasePage {
  // ─── Navigation ──────────────────────────────────────────────────
  get icsTab(): Locator {
    return this.page.getByRole('button', { name: /📅/ });
  }

  // ─── Feed List ───────────────────────────────────────────────────
  get feedList(): Locator {
    return this.page.locator('.card').filter({ has: this.page.locator('code') });
  }

  get createFeedButton(): Locator {
    return this.page.getByRole('button', { name: /创建.*订阅|ICS/ });
  }

  get embedButton(): Locator {
    return this.page.getByRole('button', { name: /🖥️|embed|Embed/ });
  }

  // ─── Create/Edit Form ────────────────────────────────────────────
  get feedNameInput(): Locator {
    return this.page.getByPlaceholder(/名称|Name/);
  }

  get feedDescInput(): Locator {
    return this.page.getByPlaceholder(/描述|Description/);
  }

  get saveFeedButton(): Locator {
    return this.page.getByRole('button', { name: /保存|创建/ }).last();
  }

  // ─── Embed Dialog ────────────────────────────────────────────────
  get embedDialog(): Locator {
    return this.page.locator('.card').filter({ has: this.page.locator('code') });
  }

  get embedCode(): Locator {
    return this.page.locator('code');
  }

  async goto(familyId: string): Promise<void> {
    await this.page.goto(`/family/${familyId}/ics`);
    await this.page.waitForLoadState('networkidle');
  }

  /** Click the ICS tab from the sidebar navigation. */
  async openIcsTab(): Promise<void> {
    await this.icsTab.click();
    await this.page.waitForTimeout(500);
  }

  /** Open the embed dialog. */
  async openEmbedDialog(): Promise<void> {
    await this.embedButton.click();
    await this.page.waitForTimeout(500);
  }

  /** Get the ICS URL text displayed for the first feed. */
  async getFirstFeedUrl(): Promise<string> {
    const codeBlocks = this.page.locator('code.text-primary');
    const first = codeBlocks.first();
    await first.waitFor({ state: 'visible', timeout: 5000 });
    return (await first.textContent()) || '';
  }

  /** Create a new ICS feed with the given name. */
  async createFeed(name: string, authType: 'api_key' | 'basic' = 'api_key'): Promise<void> {
    await this.createFeedButton.click();
    await this.page.waitForTimeout(300);
    await this.feedNameInput.fill(name);
    // Select auth type
    if (authType === 'basic') {
      await this.page.getByRole('radio', { name: /Basic Auth/ }).click();
      await this.page.getByPlaceholder(/密码|Password/).fill('test1234');
    }
    await this.saveFeedButton.click();
    await this.page.waitForTimeout(1000);
  }
}
