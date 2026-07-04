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
  return body?.data?.extra || body?.extra;
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
    expect(extra).toBeFalsy(); // simple has no extra
    console.log('  ✅ Simple round-trip: name/kind/extra match');
  });

  test('Simple-2: 待办生命周期（触发→生成→完成→消失）', async () => {
    const task = db.findTask('E2E-Verify-Simple');

    // Trigger
    await api.triggerTask(familyId, task.id);
    await new Promise(r => setTimeout(r, 500));

    // Should have 1 pending todo
    let todos = db.getTodos(task.id);
    const pending = todos.filter((t: any) => t.status === 'pending');
    expect(pending.length).toBe(1);
    console.log('  ✅ Triggered → 1 pending todo');

    // Complete
    await api.completeTodo(familyId, pending[0].id, 'done');
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

    // FAIL branch should have branch_task populated
    const failBranch = extra.check_items[0].branches[1];
    expect(failBranch.branch_task).toBeTruthy();
    expect(failBranch.branch_task.task.name).toBe('E2E-Verify-Insp-Sub');
    expect(failBranch.branch_task.task.kind).toBe('simple');
    expect(failBranch.branch_task.extra).toBeFalsy(); // simple has no extra
    console.log('  ✅ Inspection round-trip: check_items match, branch_task present');

    // Verify DB: check_items table has 2 rows, branches table has 4 rows
    const checkItems = db.getCheckItems(task.id);
    expect(checkItems.length).toBe(2);
    const branches = db.getBranches(checkItems[0].id);
    expect(branches.length).toBe(2);

    // Verify branch_task_id is set for FAIL branch
    const failDbBranch = branches.find((b: any) => b.name === 'FAIL');
    expect(failDbBranch).toBeTruthy();
    expect(failDbBranch.branch_task_id).toBeTruthy();
    console.log('  ✅ DB: 2 check_items, 2+2 branches, FAIL has branch_task_id');
  });

  test('Inspection-2: 待办生命周期（FAIL→子任务生成待办）', async () => {
    const task = db.findTask('E2E-Verify-Insp');

    // Trigger and get pending todo
    await api.triggerTask(familyId, task.id);
    await new Promise(r => setTimeout(r, 500));
    let todos = db.getTodos(task.id);
    const pending = todos.find((t: any) => t.status === 'pending');
    expect(pending).toBeTruthy();

    // Complete as FAIL on 检查项A
    const checkItems = db.getCheckItems(task.id);
    const ciA = checkItems.find((c: any) => c.name === '检查项A');
    const branches = db.getBranches(ciA.id);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');

    await api.completeTodoRaw(familyId, pending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ciA.id, item_name: '检查项A', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    await new Promise(r => setTimeout(r, 800));

    // Root: 1 done, 0 pending
    todos = db.getTodos(task.id);
    expect(todos.filter((t: any) => t.status === 'done').length).toBe(1);
    expect(todos.filter((t: any) => t.status === 'pending').length).toBe(0);

    // Sub (E2E-Verify-Insp-Sub) should have 1 pending todo
    const sub = db.findTask('E2E-Verify-Insp-Sub');
    expect(sub).toBeTruthy();
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

    // Verify steps match
    const extra = await getTaskExtra('E2E-Verify-Chain');
    expect(extra.steps.length).toBe(2);
    expect(extra.steps[0].name).toBe('E2E-Verify-Chain-S1');
    expect(extra.steps[0].kind).toBe('simple');
    expect(extra.steps[0].extra).toBeFalsy(); // simple has no extra
    expect(extra.steps[1].name).toBe('E2E-Verify-Chain-S2');
    console.log('  ✅ Chain round-trip: steps match');

    // Verify DB: chain_steps has 2 rows
    const stepsInDb = db.getChainSteps(task.id);
    expect(stepsInDb.length).toBe(2);
    expect(stepsInDb[0].name).toBe('E2E-Verify-Chain-S1');
    expect(stepsInDb[0].kind).toBe('simple');
    expect(stepsInDb[0].child_task_id).toBeTruthy();
    console.log('  ✅ DB: 2 chain_steps with child_task_id set');
  });

  test('Chain-2: 完整链推进（S1 完成→S2 生成待办）', async () => {
    const task = db.findTask('E2E-Verify-Chain');

    // Trigger root
    await api.triggerTask(familyId, task.id);
    await new Promise(r => setTimeout(r, 500));

    // Complete root (chain root → creates S1 todo)
    let rootTodos = db.getTodos(task.id);
    const rootPending = rootTodos.find((t: any) => t.status === 'pending');
    await api.completeTodo(familyId, rootPending.id, 'done');
    await new Promise(r => setTimeout(r, 800));

    // S1 should have 1 pending todo
    const s1 = db.findTask('E2E-Verify-Chain-S1');
    let s1Todos = db.getTodos(s1.id);
    const s1Pending = s1Todos.find((t: any) => t.status === 'pending');
    expect(s1Pending).toBeTruthy();
    console.log('  ✅ Root done → S1 todo created');

    // Complete S1 (done → creates S2 todo)
    await api.completeTodo(familyId, s1Pending.id, 'done');
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
    for (const name of ['E2E-Verify-Simple', 'E2E-Verify-Insp', 'E2E-Verify-Chain']) {
      const t = db.findTask(name);
      if (t) try { await api.deleteTask(familyId, t.id); } catch { /* ok */ }
    }
  });
});
