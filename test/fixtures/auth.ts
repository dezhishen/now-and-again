/**
 * Auth fixture — reusable login / logout steps.
 *
 * Usage:
 *   import { ensureLoggedIn } from '@fixtures/auth';
 *   test.beforeAll(async () => { await ensureLoggedIn(page); });
 */
import { Page } from '@playwright/test';
import { LoginPage } from '../pages/login';
import * as api from '../utils/api';

const DEFAULT_USER = 'admin';
const DEFAULT_PASS = '12345678';

/**
 * Ensure a user is logged in via the browser UI.
 * If already on a dashboard/family page, skip login.
 */
export async function ensureLoggedIn(
  page: Page,
  username: string = DEFAULT_USER,
  password: string = DEFAULT_PASS,
): Promise<void> {
  const currentUrl = page.url();

  // Already logged in if we see the main navigation
  const navVisible = await page.getByRole('button', { name: /首页/ }).isVisible().catch(() => false);
  if (navVisible) return;

  // On login page — perform login
  if (currentUrl.includes('/login')) {
    const loginPage = new LoginPage(page);
    await loginPage.login(username, password);
    return;
  }

  // Navigate to login and do it
  await page.goto('/login');
  const loginPage = new LoginPage(page);
  await loginPage.login(username, password);
}

/**
 * Login via API only (no browser UI). Faster for setup-only scenarios.
 * Returns the auth token string.
 */
export async function loginViaApi(
  username: string = DEFAULT_USER,
  password: string = DEFAULT_PASS,
): Promise<string> {
  const res = await api.login(username, password);
  if (res.status !== 200) {
    throw new Error(`Login failed: ${res.status} ${JSON.stringify(res.data)}`);
  }
  return res.token;
}
