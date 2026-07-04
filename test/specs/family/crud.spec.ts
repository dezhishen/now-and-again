import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';

test.describe('家庭管理', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
  });

  test('STEP-1: 创建家庭并获取列表', async () => {
    // Check if family already exists (can only create one per user)
    let res = await api.listMyFamilies();
    const families = (res.data?.data || res.data) as any[];
    if (!families || families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      // 200=already exists, 201=created — both are fine
      expect([200, 201]).toContain(cr.status);
    }

    res = await api.listMyFamilies();
    expect(res.status).toBe(200);
    const list = res.data?.data || res.data;
    expect(Array.isArray(list)).toBe(true);
    expect(list.length).toBeGreaterThan(0);
  });

  test('STEP-2: 浏览器登录后跳转到家庭页面', async ({ page }) => {
    await page.goto('/login');
    await page.getByPlaceholder('输入用户名').fill('admin');
    await page.getByPlaceholder('输入密码').fill('12345678');
    await page.getByRole('button', { name: '登录' }).click();

    // Wait for Vue router navigation (home → family)
    await page.waitForTimeout(2000);

    const url = page.url();
    const isValid = url.includes('/family') || url.includes('/families') || url === 'http://localhost:5173/';
    expect(isValid).toBe(true);
  });
});
