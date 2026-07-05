import { test, expect } from '@playwright/test';
import { LoginPage } from '../../pages/login';

test.describe('登录模块', () => {

  test('STEP-1: 登录页面元素完整', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await expect(loginPage.usernameInput).toBeVisible();
    await expect(loginPage.passwordInput).toBeVisible();
    await expect(loginPage.loginButton).toBeVisible();
  });

  test('STEP-2: admin/12345678 登录成功', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', '12345678');
    await expect(page).not.toHaveURL(/\/login/);
  });

  test('STEP-3: 错误密码显示错误', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.usernameInput.fill('admin');
    await loginPage.passwordInput.fill('wrongpassword');
    await loginPage.loginButton.click();
    await expect(
      page
        .locator('text=用户名或密码错误')
        .or(page.locator('text=错误'))
        .or(page.locator('text=invalid credentials'))
        .or(page.locator('text=请先登录'))
    ).toBeVisible({ timeout: 5000 });
  });
});
