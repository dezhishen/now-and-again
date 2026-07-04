/**
 * 独立巡检嵌套待办测试 — 非 chain 场景
 *
 * 场景：inspection → inspection → simple（无 chain 参与）
 * 验证：CreatedByKind="simple" 的根巡检任务能正确触发嵌套待办生成
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;

test.describe('独立巡检嵌套待办', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 inspection → inspection → simple（独立，非 chain）', async () => {
    const res = await api.createTask(familyId, {
      name: 'E2E-Standalone',
      kind: 'inspection',
      extra: {
        check_items: [{
          name: '检查',
          branches: [
            { name: 'OK', create_todo: false },
            {
              name: 'FAIL', create_todo: true,
              branch_task: {
                task: { name: 'E2E-Standalone-L2', kind: 'inspection', schedule_type: 'once', schedule_data: { time: '09:00' } },
                extra: {
                  check_items: [{
                    name: '子检查',
                    branches: [
                      { name: 'OK', create_todo: false },
                      { name: 'FAIL', create_todo: true, branch_task: { task: { name: 'E2E-Standalone-DEEP', kind: 'simple', schedule_type: 'once', schedule_data: { time: '09:00' } } } },
                    ],
                  }],
                },
              },
            },
          ],
        }],
      },
    });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 1000));
  });

  test('STEP-2: 验证根巡检 CreatedByKind = "simple"', async () => {
    const root = db.findTask('E2E-Standalone');
    expect(root).toBeTruthy();
    expect(root.kind).toBe('inspection');
    // User-created task: CreatedByKind is "simple" (DB default)
    expect(root.created_by_kind).toBe('simple');
    console.log('  ✅ Root: kind=inspection, created_by=simple (user-created)');
  });

  test('STEP-3: 触发并完成根巡检(FAIL) → 验证 L2 获得待办', async () => {
    const root = db.findTask('E2E-Standalone');
    // Trigger todo
    await api.triggerTask(familyId, root.id);
    await new Promise(r => setTimeout(r, 500));

    // Complete as FAIL
    const rootTodos = db.getTodos(root.id);
    const pending = rootTodos.find((t: any) => t.status === 'pending');
    expect(pending).toBeTruthy();

    const checkItems = db.getCheckItems(root.id);
    const ci = checkItems.find((c: any) => c.name === '检查');
    const branches = db.getBranches(ci.id);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');

    await api.completeTodoRaw(familyId, pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ci.id, item_name: '检查', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    await new Promise(r => setTimeout(r, 800));

    // L2 should have a pending todo
    const l2 = db.findTask('E2E-Standalone-L2');
    expect(l2).toBeTruthy();
    const l2Todos = db.getTodos(l2.id);
    expect(l2Todos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ Root(FAIL) → L2 todo created');
  });

  test('STEP-4: 完成 L2(FAIL) → 验证 DEEP 获得待办', async () => {
    const l2 = db.findTask('E2E-Standalone-L2');
    const checkItems = db.getCheckItems(l2.id);
    const ci = checkItems.find((c: any) => c.name === '子检查');
    const branches = db.getBranches(ci.id);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');

    const l2Todos = db.getTodos(l2.id);
    const pending = l2Todos.find((t: any) => t.status === 'pending');
    expect(pending).toBeTruthy();

    await api.completeTodoRaw(familyId, pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ci.id, item_name: '子检查', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    await new Promise(r => setTimeout(r, 800));

    // DEEP should have a pending todo
    const deep = db.findTask('E2E-Standalone-DEEP');
    expect(deep).toBeTruthy();
    const deepTodos = db.getTodos(deep.id);
    expect(deepTodos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ L2(FAIL) → DEEP todo created');
  });

  test.afterAll(async () => {
    const root = db.findTask('E2E-Standalone');
    if (root) try { await api.deleteTask(familyId, root.id); } catch { /* ok */ }
  });
});
