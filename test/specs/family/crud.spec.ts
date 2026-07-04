import { test, expect } from '@playwright/test';

test.describe('家庭管理', () => {

  test('浏览器登录→创建家庭→进入', async ({ page }) => {
    // Login
    await page.goto('/login');
    await page.getByPlaceholder('输入用户名').fill('admin');
    await page.getByPlaceholder('输入密码').fill('12345678');
    await page.getByRole('button', { name: '登录' }).click();
    await page.waitForTimeout(2000);

    // Should be on families page
    const url = page.url();
    expect(url).toContain('/families');

    // If no family, create one
    const noFamily = page.locator('text=暂无家庭');
    if (await noFamily.count() > 0) {
      await page.getByRole('button', { name: /创建/ }).first().click();
      await page.waitForTimeout(500);
      await page.locator('input[placeholder*="家庭"]').first().fill('E2E测试家庭');
      await page.getByRole('button', { name: /创建|确定/ }).click();
      await page.waitForTimeout(1500);
    }

    // Enter family
    const enterBtn = page.locator('button:has-text("进入")').first();
    if (await enterBtn.count() > 0) {
      await enterBtn.click();
      await page.waitForTimeout(2000);
    }

    // Should be on family page (dashboard)
    expect(page.url()).toContain('/family');
  });
});
