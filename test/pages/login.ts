/**
 * Login page (`/login`).
 */
import { Page, Locator } from '@playwright/test';

export class LoginPage {
  readonly usernameInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly registerLink: Locator;
  readonly heading: Locator;

  constructor(public page: Page) {
    this.usernameInput = page.getByPlaceholder('输入用户名');
    this.passwordInput = page.getByPlaceholder('输入密码');
    this.loginButton = page.getByRole('button', { name: '登录' });
    this.registerLink = page.getByRole('link', { name: '注册' });
    this.heading = page.getByRole('heading', { name: '登录' });
  }

  async goto(): Promise<void> {
    await this.page.goto('/login');
    await this.heading.waitFor({ state: 'visible' });
  }

  /** Fill and submit the login form. */
  async login(username: string, password: string): Promise<void> {
    await this.usernameInput.fill(username);
    await this.passwordInput.fill(password);
    await this.loginButton.click();
    // Wait for redirect away from login page
    await this.page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 10000 });
  }
}
