/**
 * 登录模块测试用例
 *
 * 测试目标：
 * - 验证登录页面元素完整性
 * - 验证 admin/12345678 登录成功
 * - 验证错误密码登录失败
 */
import { test, expect } from '@playwright/test';
import { LoginPage } from '../../pages/login';
import { loginViaApi } from '../../fixtures/auth';

test.describe('登录模块', () => {

  test('STEP-1: 登录页面显示用户名、密码输入框和登录按钮', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    await expect(loginPage.usernameInput).toBeVisible();
    await expect(loginPage.passwordInput).toBeVisible();
    await expect(loginPage.loginButton).toBeVisible();
    await expect(loginPage.registerLink).toBeVisible();
  });

  test('STEP-2: 使用 admin/12345678 登录，跳转到家庭管理页', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', '12345678');

    // Should redirect to /families or /family
    await expect(page).not.toHaveURL(/\/login/);
  });

  test('STEP-3: 错误密码登录应显示错误信息', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.usernameInput.fill('admin');
    await loginPage.passwordInput.fill('wrongpassword');
    await loginPage.loginButton.click();

    // Should stay on login page or show error
    await expect(page.locator('text=用户名或密码错误').or(page.locator('text=错误'))).toBeVisible({ timeout: 5000 });
  });

  test('STEP-4: API 登录获取 token', async () => {
    const token = await loginViaApi('admin', '12345678');
    expect(token).toBeTruthy();
    expect(typeof token).toBe('string');
    expect(token.length).toBeGreaterThan(10);
  });
});
