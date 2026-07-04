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

    const trigRes = await api.triggerTask(familyId, root.id);
    expect([200, 201]).toContain(trigRes.status);
    await new Promise(r => setTimeout(r, 500));

    // Complete root
    const rootTodos = db.getTodos(root.id);
    const rootPending = rootTodos.find((t: any) => t.status === 'pending');
    expect(rootPending).toBeTruthy();
    const compRes = await api.completeTodo(familyId, rootPending.id, 'done');
    expect([200, 201]).toContain(compRes.status);
    await new Promise(r => setTimeout(r, 800));

    // Root todo should now be done
    const rootTodosAfter = db.getTodos(root.id);
    expect(rootTodosAfter.filter((t: any) => t.status === 'done').length).toBe(1);

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

    const compS1Res = await api.completeTodoRaw(familyId, s1Pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ci.id, item_name: 'C1', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    expect([200, 201]).toContain(compS1Res.status);
    await new Promise(r => setTimeout(r, 800));

    // S1 todo should now be done
    const s1TodosAfter = db.getTodos(s1.id);
    expect(s1TodosAfter.filter((t: any) => t.status === 'done').length).toBe(1);

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

    const compFailRes = await api.completeTodoRaw(familyId, pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ci.id, item_name: 'SUB', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    expect([200, 201]).toContain(compFailRes.status);
    await new Promise(r => setTimeout(r, 800));

    // S1-FAIL todo should now be done
    const s1fTodosAfter = db.getTodos(s1fail.id);
    expect(s1fTodosAfter.filter((t: any) => t.status === 'done').length).toBe(1);

    // DEEP should have a pending todo
    const deep = db.findTask('E2E-NestedTodo-DEEP');
    expect(deep).toBeTruthy();
    const deepTodos = db.getTodos(deep.id);
    expect(deepTodos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ S1-FAIL(FAIL) done → DEEP todo created');
  });

  test('STEP-5: 验证所有 CreatedByKind + inspection_results + chain_steps + GetExtra 叶子', async () => {
    // Assert CreatedByKind for every task in the tree
    db.assertCreatedByKind('E2E-NestedTodo-Root', 'chain');
    db.assertCreatedByKind('E2E-NestedTodo-S1', 'chain');
    db.assertCreatedByKind('E2E-NestedTodo-S1-FAIL', 'inspection');
    db.assertCreatedByKind('E2E-NestedTodo-DEEP', 'inspection');
    db.assertCreatedByKind('E2E-NestedTodo-S2', 'chain');

    // Verify leaf tasks' is_root and schedule_type
    const root = db.findTask('E2E-NestedTodo-Root');
    expect(root.is_root).toBe(1);
    const s2 = db.findTask('E2E-NestedTodo-S2');
    expect(s2.is_root).toBe(0);
    expect(s2.schedule_type).toBeTruthy();
    const deep = db.findTask('E2E-NestedTodo-DEEP');
    expect(deep.is_root).toBe(0);
    expect(deep.kind).toBe('simple');

    // Verify inspection_results recorded for S1 and S1-FAIL
    const s1 = db.findTask('E2E-NestedTodo-S1');
    const s1Results = db.getInspectionResults(s1.id);
    expect(s1Results.length).toBeGreaterThanOrEqual(1);
    expect(s1Results[0].item_name).toBe('C1');

    const s1fail = db.findTask('E2E-NestedTodo-S1-FAIL');
    const s1fResults = db.getInspectionResults(s1fail.id);
    expect(s1fResults.length).toBeGreaterThanOrEqual(1);
    expect(s1fResults[0].item_name).toBe('SUB');

    // Verify chain_steps table for root: all fields
    const steps = db.getChainSteps(root.id);
    expect(steps.length).toBe(2);
    expect(steps[0].name).toBe('E2E-NestedTodo-S1');
    expect(steps[0].kind).toBe('inspection');
    expect(steps[0].child_task_id).toBeTruthy();
    expect(steps[1].name).toBe('E2E-NestedTodo-S2');
    expect(steps[1].kind).toBe('simple');
    expect(steps[1].child_task_id).toBeTruthy();

    // Verify GetExtra recursively reaches all leaf nodes
    const token = api.getToken();
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra ?? body?.extra ?? null;

    // S1: inspection → check_items → FAIL → branch_task(inspection) → check_items → FAIL → branch_task(simple)
    const s1Step = extra.steps[0];
    expect(s1Step.name).toBe('E2E-NestedTodo-S1');
    expect(s1Step.kind).toBe('inspection');
    const s1Extra = s1Step.extra;
    expect(s1Extra.check_items[0].name).toBe('C1');
    const s1Fail = s1Extra.check_items[0].branches.find((b: any) => b.name === 'FAIL');
    expect(s1Fail.create_todo).toBe(true);
    expect(s1Fail.branch_task.task.name).toBe('E2E-NestedTodo-S1-FAIL');
    expect(s1Fail.branch_task.task.kind).toBe('inspection');

    // Recursively into S1-FAIL
    const s1fExtra = s1Fail.branch_task.extra;
    expect(s1fExtra.check_items[0].name).toBe('SUB');
    const s1fFail = s1fExtra.check_items[0].branches.find((b: any) => b.name === 'FAIL');
    expect(s1fFail.branch_task.task.name).toBe('E2E-NestedTodo-DEEP');
    expect(s1fFail.branch_task.task.kind).toBe('simple');
    expect(s1fFail.branch_task.extra ?? null).toBeNull(); // simple 叶子

    // S2: simple leaf
    const s2Step = extra.steps[1];
    expect(s2Step.name).toBe('E2E-NestedTodo-S2');
    expect(s2Step.kind).toBe('simple');
    expect(s2Step.extra ?? null).toBeNull();

    console.log('  ✅ All 5 tasks CreatedByKind + inspection_results + chain_steps + GetExtra 递归到叶子');
  });

  test.afterAll(async () => {
    const root = db.findTask('E2E-NestedTodo-Root');
    if (root) try { await api.deleteTask(familyId, root.id); } catch { /* ok */ }
  });
});
