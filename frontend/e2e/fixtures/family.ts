/**
 * Family fixture — ensure a test family exists.
 *
 * Usage:
 *   import { ensureFamily } from '@fixtures/family';
 *   const familyId = await ensureFamily(page, '测试家庭');
 */
import { Page } from '@playwright/test';
import { FamilyManagePage } from '../pages/families';
import { ensureLoggedIn } from './auth';
import * as api from '../utils/api';

/**
 * Ensure a family exists and enter it.
 * Returns the family ID.
 * If the family already exists, just enter it.
 */
export async function ensureFamily(
  page: Page,
  name: string = '测试家庭',
): Promise<string> {
  await ensureLoggedIn(page);

  // First try via API: check if family already exists
  const listRes = await api.listMyFamilies();
  if (listRes.status === 200 && Array.isArray(listRes.data)) {
    const existing = listRes.data.find((f: any) => f.name === name);
    if (existing) {
      api.setFamilyId(existing.id);
      // Navigate to family if not already there
      if (!page.url().includes('/family')) {
        await page.goto(`/family?id=${existing.id}`);
        await page.waitForTimeout(500);
      }
      return existing.id;
    }
  }

  // Create new family via UI
  const famPage = new FamilyManagePage(page);
  await famPage.goto();
  await famPage.createFamily(name);

  // Wait for creation to complete
  await page.waitForTimeout(1000);

  // Get the ID from API
  const res = await api.listMyFamilies();
  if (res.status === 200 && Array.isArray(res.data)) {
    const created = res.data.find((f: any) => f.name === name);
    if (created) {
      api.setFamilyId(created.id);
      await famPage.enterFamily(name);
      return created.id;
    }
  }

  throw new Error('Failed to create or find family: ' + name);
}
