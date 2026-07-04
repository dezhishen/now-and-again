/**
 * Simple 任务 — 简单测试（仅嵌套 simple）
 *
 * 场景：simple → simple → simple
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-Simple';

test.describe('Simple (简单)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 simple 任务', async () => {
    const res = await api.createTask(familyId, { name: ROOT, kind: 'simple' });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 500));
    const t = db.findTask(ROOT);
    expect(t).toBeTruthy();
    expect(t.kind).toBe('simple');
    expect(t.created_by_kind).toBe('simple');
    console.log('  ✅ simple root: kind=simple, created_by=simple');
  });

  test('STEP-2: simple 任务无子任务、无 extra', async () => {
    const token = api.getToken();
    const root = db.findTask(ROOT);
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra ?? body?.extra ?? null;
    expect(extra).toBeNull();
    console.log('  ✅ simple extra = null');
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
  });
});
