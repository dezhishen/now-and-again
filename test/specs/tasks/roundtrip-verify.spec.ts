/**
 * 综合验证测试 — round-trip + todo 生命周期
 *
 * 覆盖：
 *   1. simple: 创建→GET 一致性 + 待办生命周期
 *   2. inspection: 创建→GET 一致性 + 待办生命周期
 *   3. chain: 创建→GET 一致性 + 待办链推进
 *   4. inspection 嵌套: 创建→GET 递归 extra 一致性
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;

async function getTaskExtra(taskName: string): Promise<any> {
  const token = api.getToken();
  const task = db.findTask(taskName);
  if (!task) return null;
  const res = await fetch(`http://localhost:8080/api/tasks/${task.id}?with_extra=true`, {
    headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
  });
  const body = await res.json();
  return body?.data?.extra ?? body?.extra ?? null;
}

test.describe('综合验证：round-trip + todo 生命周期', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  // ═══════════════════════════════════════════════════════════════
  // listTasks: 验证列表端点数据完整性
  // ═══════════════════════════════════════════════════════════════

  test('listTasks-1: 创建任务后 listTasks 包含该任务且数据结构正确', async () => {
    await api.createTask(familyId, {
      name: 'E2E-Verify-List', kind: 'simple',
    });
    await new Promise(r => setTimeout(r, 500));

    const listRes = await api.listTasks(familyId);
    expect([200, 201]).toContain(listRes.status);

    const tasks: any[] = listRes.data?.data || listRes.data || [];
    expect(tasks.length).toBeGreaterThan(0);

    const found = tasks.find((t: any) => t.name === 'E2E-Verify-List');
    expect(found).toBeTruthy();
    expect(found.kind).toBe('simple');
    expect(found.schedule_type).toBeTruthy();
    expect(found.is_root).toBe(true);
    expect(found.enabled).toBe(true);
    console.log('  ✅ listTasks includes new task with correct shape');
  });

  // ═══════════════════════════════════════════════════════════════
  // Simple: round-trip + todo lifecycle
  // ═══════════════════════════════════════════════════════════════

  test('Simple-1: 创建→GET 一致性', async () => {
    await api.createTask(familyId, {
      name: 'E2E-Verify-Simple',
      kind: 'simple',
      scheduleType: 'daily',
      scheduleData: { time: '08:30' },
    });
    await new Promise(r => setTimeout(r, 500));

    const task = db.findTask('E2E-Verify-Simple');
    expect(task).toBeTruthy();
    expect(task.name).toBe('E2E-Verify-Simple');
    expect(task.kind).toBe('simple');
    expect(task.created_by_kind).toBe('simple');

    const extra = await getTaskExtra('E2E-Verify-Simple');
    expect(extra).toBeNull(); // simple has no extra
    console.log('  ✅ Simple round-trip: name/kind/extra match');
  });

  test('Simple-2: 待办生命周期（触发→生成→完成→消失）', async () => {
    const task = db.findTask('E2E-Verify-Simple');

    // Trigger
    const trigRes = await api.triggerTask(familyId, task.id);
    expect([200, 201]).toContain(trigRes.status);
    await new Promise(r => setTimeout(r, 500));

    // Should have 1 pending todo
    let todos = db.getTodos(task.id);
    const pending = todos.filter((t: any) => t.status === 'pending');
    expect(pending.length).toBe(1);
    console.log('  ✅ Triggered → 1 pending todo');

    // Complete
    const compRes = await api.completeTodo(familyId, pending[0].id, 'done');
    expect([200, 201]).toContain(compRes.status);
    await new Promise(r => setTimeout(r, 300));

    // Should have 1 done todo, 0 pending
    todos = db.getTodos(task.id);
    expect(todos.filter((t: any) => t.status === 'done').length).toBe(1);
    expect(todos.filter((t: any) => t.status === 'pending').length).toBe(0);
    console.log('  ✅ Completed → 1 done, 0 pending');
  });

  // ═══════════════════════════════════════════════════════════════
  // Inspection: round-trip + todo lifecycle
  // ═══════════════════════════════════════════════════════════════

  const INSP_EXTRA = {
    check_items: [
      { name: '检查项A', branches: [
        { name: 'OK', create_todo: false },
        { name: 'FAIL', create_todo: true, branch_task: { task: { name: 'E2E-Verify-Insp-Sub', kind: 'simple' } } },
      ]},
      { name: '检查项B', branches: [
        { name: 'OK', create_todo: false },
        { name: 'WARN', create_todo: false },
      ]},
    ],
  };

  test('Inspection-1: 创建→GET 一致性', async () => {
    await api.createTask(familyId, {
      name: 'E2E-Verify-Insp', kind: 'inspection', extra: INSP_EXTRA,
    });
    await new Promise(r => setTimeout(r, 800));

    const task = db.findTask('E2E-Verify-Insp');
    expect(task).toBeTruthy();
    expect(task.kind).toBe('inspection');
    expect(task.created_by_kind).toBe('simple');

    // Verify extra matches what was sent
    const extra = await getTaskExtra('E2E-Verify-Insp');
    expect(extra).toBeTruthy();
    expect(extra.check_items).toBeTruthy();
    expect(extra.check_items.length).toBe(2);
    expect(extra.check_items[0].name).toBe('检查项A');
    expect(extra.check_items[0].branches.length).toBe(2);
    expect(extra.check_items[0].branches[0].name).toBe('OK');
    expect(extra.check_items[0].branches[1].name).toBe('FAIL');
    expect(extra.check_items[1].name).toBe('检查项B');

    // FAIL branch should have branch_task populated with complete leaf data
    const failBranch = extra.check_items[0].branches[1];
    expect(failBranch.name).toBe('FAIL');
    expect(failBranch.create_todo).toBe(true);
    expect(failBranch.branch_task).toBeTruthy();
    expect(failBranch.branch_task.task.name).toBe('E2E-Verify-Insp-Sub');
    expect(failBranch.branch_task.task.kind).toBe('simple');
    expect(failBranch.branch_task.extra ?? null).toBeNull(); // simple has no extra (Go omitempty)

    // 检查项B: both branches should NOT have create_todo, no branch_task
    expect(extra.check_items[1].name).toBe('检查项B');
    expect(extra.check_items[1].branches.length).toBe(2);
    for (const b of extra.check_items[1].branches) {
      expect(b.create_todo).toBe(false);
      expect(b.branch_task).toBeFalsy();
    }
    console.log('  ✅ Inspection round-trip: check_items match, all branches verified to leaf');

    // Verify DB: check_items table has 2 rows with correct names
    const checkItems = db.getCheckItems(task.id);
    expect(checkItems.length).toBe(2);
    expect(checkItems[0].name).toBe('检查项A');
    expect(checkItems[1].name).toBe('检查项B');

    // 检查项A branches: OK=no task, FAIL=has task
    const branchesA = db.getBranches(checkItems[0].id);
    expect(branchesA.length).toBe(2);
    const failDbBranch = branchesA.find((b: any) => b.name === 'FAIL');
    expect(failDbBranch).toBeTruthy();
    expect(failDbBranch.create_todo).toBe(1);
    expect(failDbBranch.branch_task_id).toBeTruthy();
    const okDbBranch = branchesA.find((b: any) => b.name === 'OK');
    expect(okDbBranch).toBeTruthy();
    expect(okDbBranch.create_todo).toBe(0);
    expect(okDbBranch.branch_task_id).toBeFalsy();

    // 检查项B branches: all no task
    const branchesB = db.getBranches(checkItems[1].id);
    expect(branchesB.length).toBe(2);
    for (const b of branchesB) {
      expect(b.create_todo).toBe(0);
      expect(b.branch_task_id).toBeFalsy();
    }
    console.log('  ✅ DB: 2 check_items, 4 branches, all create_todo/branch_task_id verified at leaf');
  });

  test('Inspection-2: 待办生命周期（FAIL→子任务生成待办）', async () => {
    const task = db.findTask('E2E-Verify-Insp');

    // Trigger and get pending todo
    const trigRes = await api.triggerTask(familyId, task.id);
    expect([200, 201]).toContain(trigRes.status);
    await new Promise(r => setTimeout(r, 500));
    let todos = db.getTodos(task.id);
    const pending = todos.find((t: any) => t.status === 'pending');
    expect(pending).toBeTruthy();

    // Complete as FAIL on 检查项A
    const checkItems = db.getCheckItems(task.id);
    const ciA = checkItems.find((c: any) => c.name === '检查项A');
    const branches = db.getBranches(ciA.id);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');

    const compRes = await api.completeTodoRaw(familyId, pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ciA.id, item_name: '检查项A', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    expect([200, 201]).toContain(compRes.status);
    await new Promise(r => setTimeout(r, 800));

    // Root: 1 done, 0 pending
    todos = db.getTodos(task.id);
    expect(todos.filter((t: any) => t.status === 'done').length).toBe(1);
    expect(todos.filter((t: any) => t.status === 'pending').length).toBe(0);

    // Sub (E2E-Verify-Insp-Sub) should have 1 pending todo
    const sub = db.findTask('E2E-Verify-Insp-Sub');
    expect(sub).toBeTruthy();
    expect(sub.kind).toBe('simple');
    expect(sub.created_by_kind).toBe('inspection');
    const subTodos = db.getTodos(sub.id);
    expect(subTodos.filter((t: any) => t.status === 'pending').length).toBe(1);
    console.log('  ✅ Root done → Sub todo created');

    // Also verify inspection_results has 1 row
    const results = db.getInspectionResults(task.id);
    expect(results.length).toBe(1);
    expect(results[0].item_name).toBe('检查项A');
    expect(results[0].branch_name).toBe('FAIL');
    console.log('  ✅ inspection_result recorded');
  });

  // ═══════════════════════════════════════════════════════════════
  // Chain: round-trip + todo chain progression
  // ═══════════════════════════════════════════════════════════════

  const CHAIN_STEPS = [
    { name: 'E2E-Verify-Chain-S1', kind: 'simple' },
    { name: 'E2E-Verify-Chain-S2', kind: 'simple' },
  ];

  test('Chain-1: 创建→GET 一致性', async () => {
    await api.createTask(familyId, {
      name: 'E2E-Verify-Chain', kind: 'chain',
      extra: { steps: CHAIN_STEPS },
    });
    await new Promise(r => setTimeout(r, 800));

    const task = db.findTask('E2E-Verify-Chain');
    expect(task).toBeTruthy();
    expect(task.kind).toBe('chain');
    expect(task.created_by_kind).toBe('chain'); // SaveExtra updated it

    // Verify steps match (complete leaf data for every step)
    const extra = await getTaskExtra('E2E-Verify-Chain');
    expect(extra.steps.length).toBe(2);
    const expectedSteps = ['E2E-Verify-Chain-S1', 'E2E-Verify-Chain-S2'];
    for (let i = 0; i < 2; i++) {
      const s = extra.steps[i];
      expect(s.name).toBe(expectedSteps[i]);
      expect(s.kind).toBe('simple');
      expect(s.sort_order).toBe(i);
      expect(s.child_task_id).toBeTruthy();
      expect(s.extra ?? null).toBeNull();
    }
    console.log('  ✅ Chain round-trip: steps match (name/kind/sort_order/child_task_id)');

    // Verify DB: chain_steps has 2 rows
    const stepsInDb = db.getChainSteps(task.id);
    expect(stepsInDb.length).toBe(2);
    for (let i = 0; i < 2; i++) {
      expect(stepsInDb[i].name).toBe(expectedSteps[i]);
      expect(stepsInDb[i].kind).toBe('simple');
      expect(stepsInDb[i].sort_order).toBe(i);
      expect(stepsInDb[i].child_task_id).toBeTruthy();
    }
    console.log('  ✅ DB: 2 chain_steps with all fields verified');
  });

  test('Chain-2: 完整链推进（S1 完成→S2 生成待办）', async () => {
    const task = db.findTask('E2E-Verify-Chain');

    // Trigger root
    const trigRes = await api.triggerTask(familyId, task.id);
    expect([200, 201]).toContain(trigRes.status);
    await new Promise(r => setTimeout(r, 500));

    // Complete root (chain root → creates S1 todo)
    let rootTodos = db.getTodos(task.id);
    const rootPending = rootTodos.find((t: any) => t.status === 'pending');
    const compRootRes = await api.completeTodo(familyId, rootPending.id, 'done');
    expect([200, 201]).toContain(compRootRes.status);
    await new Promise(r => setTimeout(r, 800));

    // S1 should have 1 pending todo
    const s1 = db.findTask('E2E-Verify-Chain-S1');
    expect(s1.kind).toBe('simple');
    expect(s1.created_by_kind).toBe('chain');
    let s1Todos = db.getTodos(s1.id);
    const s1Pending = s1Todos.find((t: any) => t.status === 'pending');
    expect(s1Pending).toBeTruthy();
    console.log('  ✅ Root done → S1 todo created');

    // Complete S1 (done → creates S2 todo)
    const compS1Res = await api.completeTodo(familyId, s1Pending.id, 'done');
    expect([200, 201]).toContain(compS1Res.status);
    await new Promise(r => setTimeout(r, 800));

    // S1: 1 done, 0 pending
    s1Todos = db.getTodos(s1.id);
    expect(s1Todos.filter((t: any) => t.status === 'done').length).toBe(1);

    // S2 should have 1 pending todo
    const s2 = db.findTask('E2E-Verify-Chain-S2');
    const s2Todos = db.getTodos(s2.id);
    expect(s2Todos.filter((t: any) => t.status === 'pending').length).toBe(1);
    console.log('  ✅ S1 done → S2 todo created');

    // Complete S2 → no more steps
    const s2Pending = s2Todos.find((t: any) => t.status === 'pending');
    await api.completeTodo(familyId, s2Pending.id, 'done');
    await new Promise(r => setTimeout(r, 500));

    // Root: 1 done, S1: 1 done, S2: 1 done
    const final = db.getTodos(task.id);
    expect(final.filter((t: any) => t.status === 'done').length).toBe(1);
    const s2Final = db.getTodos(s2.id);
    expect(s2Final.filter((t: any) => t.status === 'done').length).toBe(1);
    console.log('  ✅ Chain complete: 3 tasks all done');
  });

  // ─── Cleanup ────────────────────────────────────────────────────

  test.afterAll(async () => {
    for (const name of ['E2E-Verify-Simple', 'E2E-Verify-Insp', 'E2E-Verify-Chain', 'E2E-Verify-List']) {
      const t = db.findTask(name);
      if (t) try { await api.deleteTask(familyId, t.id); } catch { /* ok */ }
    }
  });
});
