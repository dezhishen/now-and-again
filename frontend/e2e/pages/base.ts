/**
 * Base page with common selectors shared across all pages.
 */
import { Page, Locator } from '@playwright/test';

export class BasePage {
  constructor(public page: Page) {}

  // ─── App Header ──────────────────────────────────────────────────
  get homeButton(): Locator {
    return this.page.locator('button', { hasText: 'Now & Again' });
  }

  get userMenuButton(): Locator {
    return this.page.locator('header button').filter({ hasText: /管/ });
  }

  // ─── Navigation (sidebar) ────────────────────────────────────────
  get navDashboard(): Locator {
    return this.page.getByRole('button', { name: /首页/ });
  }

  get navTasks(): Locator {
    return this.page.getByRole('button', { name: /任务/ });
  }

  get navTemplates(): Locator {
    return this.page.getByRole('button', { name: /任务模板/ });
  }

  get navGroups(): Locator {
    return this.page.getByRole('button', { name: /小组/ });
  }

  get navLocations(): Locator {
    return this.page.getByRole('button', { name: /地点/ });
  }

  get navMembers(): Locator {
    return this.page.getByRole('button', { name: /成员/ });
  }

  get navSettings(): Locator {
    return this.page.getByRole('button', { name: /设置/ });
  }

  get navLeave(): Locator {
    return this.page.getByRole('button', { name: /离开家庭/ });
  }

  // ─── Toast ───────────────────────────────────────────────────────
  get toast(): Locator {
    return this.page.locator('[class*="toast"]').first();
  }

  async waitForToast(text?: string): Promise<void> {
    if (text) {
      await this.page.locator(`text=${text}`).first().waitFor({ state: 'visible', timeout: 5000 });
    }
  }

  // ─── Breadcrumb ──────────────────────────────────────────────────
  get breadcrumbHome(): Locator {
    return this.page.getByRole('button', { name: /🏠 首页/ });
  }

  /** Wait for the page to finish loading (no more "加载中" text). */
  async waitForLoad(): Promise<void> {
    await this.page.locator('text=加载中').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
  }
}
