/**
 * 家庭管理测试
 */
import { test, expect } from '@playwright/test';

test.describe('家庭', () => {
  test('创建家庭并进入', async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('输入用户名').fill('admin');
    await page.getByPlaceholder('输入密码').fill('12345678');
    await page.locator('[data-testid="login-submit"]').click();
    await page.waitForTimeout(3000);

    // Enter existing family or create one
    const enterBtns = page.locator('[data-testid="family-enter-btn"]');
    if (await enterBtns.count() > 0) {
      await enterBtns.first().click();
    } else {
      await page.locator('[data-testid="family-create-toggle"]').click();
      await page.locator('[data-testid="family-name-input"]').fill('E2E-' + Date.now());
      await page.locator('[data-testid="family-create-submit"]').click();
      await page.waitForTimeout(2000);
      const newEnter = page.locator('[data-testid="family-enter-btn"]').first();
      await newEnter.click();
    }
    await page.waitForTimeout(2000);
    expect(page.url()).toMatch(/\/family/);
  });
});
