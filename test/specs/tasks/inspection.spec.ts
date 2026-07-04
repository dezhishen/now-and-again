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

  test('STEP-3: 验证 GetExtra 叶子节点（check_item → branches → branch_task 全链路）', async () => {
    const token = api.getToken();
    const root = db.findTask(ROOT);
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra ?? body?.extra ?? null;

    // 验证 check_items
    expect(extra.check_items.length).toBe(1);
    const ci = extra.check_items[0];
    expect(ci.name).toBe('检查项');

    // 验证每个 branch 的完整数据
    expect(ci.branches.length).toBe(2);

    // 合格分支: 不应创建子任务
    const okBranch = ci.branches.find((b: any) => b.name === '合格');
    expect(okBranch).toBeTruthy();
    expect(okBranch.create_todo).toBe(false);
    expect(okBranch.branch_task).toBeFalsy(); // 无子任务

    // 不合格分支: 应创建 simple 子任务
    const failBranch = ci.branches.find((b: any) => b.name === '不合格');
    expect(failBranch).toBeTruthy();
    expect(failBranch.create_todo).toBe(true);
    expect(failBranch.branch_task).toBeTruthy();
    expect(failBranch.branch_task.task.name).toBe(ROOT + '-不合格处理');
    expect(failBranch.branch_task.task.kind).toBe('simple');
    expect(failBranch.branch_task.extra ?? null).toBeNull(); // simple leaf: no nested extra

    // 验证 DB: check_items + branches 表
    const dbCheckItems = db.getCheckItems(root.id);
    expect(dbCheckItems.length).toBe(1);
    expect(dbCheckItems[0].name).toBe('检查项');

    const dbBranches = db.getBranches(dbCheckItems[0].id);
    expect(dbBranches.length).toBe(2);
    const dbFail = dbBranches.find((b: any) => b.name === '不合格');
    expect(dbFail).toBeTruthy();
    expect(dbFail.create_todo).toBe(1); // SQLite stores bool as 0/1
    expect(dbFail.branch_task_id).toBeTruthy();
    const dbOk = dbBranches.find((b: any) => b.name === '合格');
    expect(dbOk).toBeTruthy();
    expect(dbOk.create_todo).toBe(0);

    console.log('  ✅ All leaves verified: check_item → 2 branches → branch_task name/kind + DB tables');
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
  });
});
