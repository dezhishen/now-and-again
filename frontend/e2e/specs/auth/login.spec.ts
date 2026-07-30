import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/login';

test.describe('登录', () => {
  test('登录页面', async ({ page }) => {
    const lp = new LoginPage(page);
    await lp.goto();
    await expect(lp.usernameInput).toBeVisible();
    await expect(lp.passwordInput).toBeVisible();
    await expect(lp.loginButton).toBeVisible();
  });

  test('admin 登录成功', async ({ page }) => {
    const lp = new LoginPage(page);
    await lp.goto();
    await lp.login('admin', '12345678');
  });

  test('错误密码提示', async ({ page }) => {
    const lp = new LoginPage(page);
    await lp.goto();
    await lp.usernameInput.fill('admin');
    await lp.passwordInput.fill('wrongpassword');
    await lp.loginButton.click();
    await page.waitForTimeout(1000);
    // Login page should still be showing after failed login
    expect(page.url()).toContain('/login');
  });
});
