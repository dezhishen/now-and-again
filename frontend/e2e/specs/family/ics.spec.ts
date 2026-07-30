/**
 * ICS 订阅测试：进入家庭 → 访问 ICS 页面
 */
import { test, expect } from '@playwright/test';

test.describe('ICS 订阅', () => {

  test('ICS 页面可访问', async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('输入用户名').fill('admin');
    await page.getByPlaceholder('输入密码').fill('12345678');
    await page.locator('[data-testid="login-submit"]').click();
    await page.waitForTimeout(3000);

    const enter = page.locator('[data-testid="family-enter-btn"]').first();
    if (await enter.count() > 0) await enter.click();
    await page.waitForTimeout(2000);

    await page.goto('/family/ics');
    await page.waitForLoadState('networkidle');
    expect(page.url()).toContain('/family/ics');
  });
});
