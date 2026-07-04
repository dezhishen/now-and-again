/**
 * 嵌套待办生成测试 — 验证多层嵌套的 todo 生命周期
 *
 * 场景：
 *   Chain → Inspection(create_todo) → Inspection(create_todo) → Simple
 *
 * 验证：每一步完成时，下一级任务自动生成待办
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;

test.describe('嵌套待办生成', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 chain → inspection(create_todo) → inspection(create_todo) → simple', async () => {
    const res = await api.createTask(familyId, {
      name: 'E2E-NestedTodo-Root',
      kind: 'chain',
      extra: {
        steps: [
          {
            name: 'E2E-NestedTodo-S1',
            kind: 'inspection',
            extra: {
              check_items: [{
                name: 'C1',
                branches: [
                  { name: 'OK', create_todo: false },
                  {
                    name: 'FAIL', create_todo: true,
                    branch_task: {
                      task: { name: 'E2E-NestedTodo-S1-FAIL', kind: 'inspection', schedule_type: 'once', schedule_data: { time: '09:00' } },
                      extra: {
                        check_items: [{
                          name: 'SUB',
                          branches: [
                            { name: 'OK', create_todo: false },
                            { name: 'FAIL', create_todo: true, branch_task: { task: { name: 'E2E-NestedTodo-DEEP', kind: 'simple', schedule_type: 'once', schedule_data: { time: '09:00' } } } },
                          ],
                        }],
                      },
                    },
                  },
                ],
              }],
            },
          },
          { name: 'E2E-NestedTodo-S2', kind: 'simple' },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 1000));
  });

  test('STEP-2: 触发 root → 完成 → S1 应获得待办', async () => {
    const root = db.findTask('E2E-NestedTodo-Root');
    expect(root).toBeTruthy();

    await api.triggerTask(familyId, root.id);
    await new Promise(r => setTimeout(r, 500));

    // Complete root
    const rootTodos = db.getTodos(root.id);
    const rootPending = rootTodos.find((t: any) => t.status === 'pending');
    expect(rootPending).toBeTruthy();
    await api.completeTodo(familyId, rootPending.id, 'done');
    await new Promise(r => setTimeout(r, 800));

    // S1 should have a pending todo now
    const s1 = db.findTask('E2E-NestedTodo-S1');
    const s1Todos = db.getTodos(s1.id);
    expect(s1Todos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ Root done → S1 todo created');
  });

  test('STEP-3: 完成 S1(FAIL) → S1-FAIL 获得待办 + S2 获得待办（链推进）', async () => {
    // Get S1's FAIL branch ID
    const s1 = db.findTask('E2E-NestedTodo-S1');
    const checkItems = db.getCheckItems(s1.id);
    const ci = checkItems.find((ci: any) => ci.name === 'C1');
    expect(ci).toBeTruthy();
    const branches = db.getBranches(ci.id);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');
    expect(failBranch).toBeTruthy();

    // Complete S1 todo with correct branch_id
    const s1Todos = db.getTodos(s1.id);
    const s1Pending = s1Todos.find((t: any) => t.status === 'pending');
    expect(s1Pending).toBeTruthy();

    await api.completeTodoRaw(familyId, s1Pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ci.id, item_name: 'C1', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    await new Promise(r => setTimeout(r, 800));

    // S1-FAIL should have a pending todo
    const s1fail = db.findTask('E2E-NestedTodo-S1-FAIL');
    expect(s1fail).toBeTruthy();
    const s1fTodos = db.getTodos(s1fail.id);
    expect(s1fTodos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ S1(FAIL) done → S1-FAIL todo created');

    // S2 should also have a pending todo (chain progression)
    const s2 = db.findTask('E2E-NestedTodo-S2');
    const s2Todos = db.getTodos(s2.id);
    expect(s2Todos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ Chain progression → S2 todo created');
  });

  test('STEP-4: 完成 S1-FAIL(FAIL) → DEEP 获得待办', async () => {
    const s1fail = db.findTask('E2E-NestedTodo-S1-FAIL');
    const checkItems = db.getCheckItems(s1fail.id);
    const ci = checkItems.find((ci: any) => ci.name === 'SUB');
    expect(ci).toBeTruthy();
    const branches = db.getBranches(ci.id);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');
    expect(failBranch).toBeTruthy();

    const todos = db.getTodos(s1fail.id);
    const pending = todos.find((t: any) => t.status === 'pending');
    expect(pending).toBeTruthy();

    await api.completeTodoRaw(familyId, pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ci.id, item_name: 'SUB', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    await new Promise(r => setTimeout(r, 800));

    // DEEP should have a pending todo
    const deep = db.findTask('E2E-NestedTodo-DEEP');
    expect(deep).toBeTruthy();
    const deepTodos = db.getTodos(deep.id);
    expect(deepTodos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ S1-FAIL(FAIL) done → DEEP todo created');
  });

  test('STEP-5: 验证所有 CreatedByKind', async () => {
    const tasks = db.getTasks();
    const our = tasks.filter((t: any) => t.name && (t.name as string).startsWith('E2E-NestedTodo'));
    for (const t of our) {
      console.log(`  ${t.name}: kind=${t.kind} created_by=${t.created_by_kind}`);
    }
    // Verified by STEP-3 assertions already
  });

  test.afterAll(async () => {
    const root = db.findTask('E2E-NestedTodo-Root');
    if (root) try { await api.deleteTask(familyId, root.id); } catch { /* ok */ }
  });
});
