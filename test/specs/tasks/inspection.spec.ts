/**
 * Inspection 任务 — 简单测试（仅嵌套 simple）
 *
 * 场景：inspection → simple（无嵌套 inspection）
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-巡检-简单';

test.describe('Inspection (简单: 仅引 simple)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 inspection，子任务仅 simple', async () => {
    const res = await api.createTask(familyId, {
      name: ROOT, kind: 'inspection',
      extra: {
        check_items: [{
          name: '检查项',
          branches: [
            { name: '合格', result: 'pass', create_todo: false },
            { name: '不合格', result: 'fail', create_todo: true, branch_task: { task: { name: ROOT + '-不合格处理', kind: 'simple' } } },
          ],
        }],
      },
    });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 800));
  });

  test('STEP-2: 验证根 CreatedByKind + 子任务 Kind/CreatedByKind', async () => {
    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === ROOT);
    expect(root).toBeTruthy();
    expect(root.kind).toBe('inspection');
    expect(root.created_by_kind).toBe('simple');
    console.log('  ✅ Root: kind=inspection, created_by=simple');

    const sub = tasks.find((t: any) => t.name === ROOT + '-不合格处理');
    expect(sub).toBeTruthy();
    expect(sub.kind).toBe('simple');
    expect(sub.created_by_kind).toBe('inspection');
    console.log('  ✅ Sub: kind=simple, created_by=inspection');
  });

  test('STEP-3: 验证 GetExtra 不递归（子任务 simple 无 extra）', async () => {
    const token = api.getToken();
    const root = db.findTask(ROOT);
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra || body?.extra;
    expect(extra.check_items).toBeTruthy();
    const failBranch = extra.check_items[0].branches.find((b: any) => b.name === '不合格');
    expect(failBranch.branch_task).toBeTruthy();
    expect(failBranch.branch_task.extra).toBeFalsy(); // simple has no extra
    console.log('  ✅ branch_task.extra = null (simple leaf)');
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
  });
});
