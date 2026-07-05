import { test, expect } from '@playwright/test';
import { IcsPage } from '../../pages/ics';
import { ensureLoggedIn } from '../../fixtures/auth';

test.describe('ICS 订阅', () => {
  let icsPage: IcsPage;
  let familyId: string;

  test.beforeEach(async ({ page }) => {
    await ensureLoggedIn(page);
    icsPage = new IcsPage(page);

    // Extract family ID from URL or navigate to dashboard first
    const url = page.url();
    const match = url.match(/\/family\/([a-f0-9-]+)/);
    if (match) {
      familyId = match[1];
      await icsPage.goto(familyId);
    } else {
      // Go to dashboard to get family context
      await page.goto('/');
      await page.waitForTimeout(1000);
      const newMatch = page.url().match(/\/family\/([a-f0-9-]+)/);
      if (newMatch) {
        familyId = newMatch[1];
        await icsPage.goto(familyId);
      }
    }
  });

  test('创建 ICS 订阅后 URL 包含名称 slug', async () => {
    test.skip(!familyId, '需要先进入家庭');

    await icsPage.goto(familyId);
    await icsPage.openIcsTab();

    // Create a feed
    const feedName = 'E2E-Test-Calendar';
    await icsPage.createFeed(feedName);

    // Verify the displayed URL contains the feed name (as slug)
    const icsUrl = await icsPage.getFirstFeedUrl();
    // URL should contain the slugified name, not just UUID
    expect(icsUrl).toContain('/api/ics/');
    // Should have the slug path segment
    expect(icsUrl).toMatch(/\/api\/ics\/[a-f0-9-]+\/E2E-Test-Calendar\.ics/);
  });

  test('嵌入日历 URL 不包含函数源码', async () => {
    test.skip(!familyId, '需要先进入家庭');

    await icsPage.goto(familyId);
    await icsPage.openIcsTab();
    await icsPage.openEmbedDialog();

    // Get the embed code text
    const embedText = await icsPage.embedCode.first().textContent();
    expect(embedText).not.toBeNull();

    // Must NOT contain function source code (the familyId bug)
    expect(embedText).not.toContain('() =>');
    expect(embedText).not.toContain('auth.activeFamilyId');

    // Must contain a valid-looking URL with a UUID family ID
    expect(embedText).toMatch(/\/calendar\/[a-f0-9-]+\?key=/);
  });

  test('复制按钮存在于 ICS 订阅列表', async () => {
    test.skip(!familyId, '需要先进入家庭');

    await icsPage.goto(familyId);
    await icsPage.openIcsTab();

    // Create a feed first so we have something to copy
    await icsPage.createFeed('E2E-Copy-Test');

    // Verify copy button exists and is visible
    const copyBtn = icsPage.page.getByRole('button', { name: /复制|Copy/ }).first();
    await expect(copyBtn).toBeVisible();
  });
});
