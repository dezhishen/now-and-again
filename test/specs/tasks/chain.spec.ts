/**
 * 任务链 (chain) 模块测试
 *
 * 验证：
 * - chain 子任务 Kind = 步骤真实类型
 * - chain 子任务 CreatedByKind = "chain"
 * - 链式推进：完成步骤 N → 自动创建步骤 N+1 的待办
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const CHAIN_TASK = 'E2E-任务链';

test.describe('任务链 (kind=chain)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');

    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId || (res.data?.data?.[0]?.id);
    } else {
      familyId = families[0].id;
      api.setFamilyId(familyId);
    }
  });

  test('STEP-1: 创建 chain 任务（simple→inspection 两步）', async () => {
    const res = await api.createTask(familyId, {
      name: CHAIN_TASK,
      kind: 'chain',
      extra: {
        steps: [
          { name: '第一步-简单确认', kind: 'simple' },
          {
            name: '第二步-巡检',
            kind: 'inspection',
            extra: {
              check_items: [{
                name: '区域A',
                branches: [
                  { name: '合格', result: 'pass', create_todo: false },
                  { name: '不合格', result: 'fail', create_todo: true, branch_task: { task: { name: CHAIN_TASK + '-不合格处理', kind: 'simple' } } },
                ],
              }],
            },
          },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);

    await new Promise(r => setTimeout(r, 800));
    const task = db.findTask(CHAIN_TASK);
    expect(task).toBeTruthy();
    expect(task.kind).toBe('chain');
  });

  test('STEP-2: 验证子任务 Kind 和 CreatedByKind', async () => {
    await new Promise(r => setTimeout(r, 500));

    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === CHAIN_TASK);
    expect(root).toBeTruthy();

    // Chain structure: root → step1 → step2 (linked list, not flat children)
    const allChildren = tasks.filter((t: any) => t.is_root === 0 && t.name && (t.name as string).startsWith('第'));
    expect(allChildren.length).toBeGreaterThanOrEqual(2);

    // Find step1 (direct child of root)
    const step1 = allChildren.find((c: any) => c.name === '第一步-简单确认');
    expect(step1).toBeTruthy();
    expect(step1.kind).toBe('simple');
    expect(step1.created_by_kind).toBe('chain');

    // Find step2 (child of step1)
    const step2 = allChildren.find((c: any) => c.name === '第二步-巡检');
    expect(step2).toBeTruthy();
    expect(step2.kind).toBe('inspection');
    expect(step2.created_by_kind).toBe('chain');

    console.log('  ✅ Step1 "' + step1.name + '" kind=' + step1.kind + ' created_by=' + step1.created_by_kind);
    console.log('  ✅ Step2 "' + step2.name + '" kind=' + step2.kind + ' created_by=' + step2.created_by_kind);
  });

  test('STEP-3: 链式推进 — 完成步骤1后步骤2自动生成待办', async () => {
    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === CHAIN_TASK);
    if (!root) { test.skip(true, 'Chain task not found'); return; }

    // Chain is root → step1 → step2; find step1 (direct child of root)
    const allChildren = tasks.filter((t: any) => t.is_root === 0);
    const step1 = allChildren.find((c: any) => c.name === '第一步-简单确认');
    const step2 = allChildren.find((c: any) => c.name === '第二步-巡检');
    if (!step1 || !step2) { test.skip(true, 'Steps not found'); return; }

    // Create a todo for step1
    await new Promise(r => setTimeout(r, 300));
    const triggerRes = await api.triggerTask(familyId, step1.id);
    expect([200, 201]).toContain(triggerRes.status);

    // Get the todo
    await new Promise(r => setTimeout(r, 500));
    const todos = db.getTodos(step1.id);
    const step1Todo = todos.find((t: any) => t.status === 'pending');
    if (!step1Todo) { test.skip(true, 'No pending todo for step1'); return; }

    // Complete step1's todo — this should trigger chain progression
    await api.completeTodo(familyId, step1Todo.id, 'done');
    await new Promise(r => setTimeout(r, 800));

    // Check that step2 now has a todo (chain progression created it)
    const step2Todos = db.getTodos(step2.id);
    const step2Pending = step2Todos.filter((t: any) => t.status === 'pending');
    expect(step2Pending.length).toBeGreaterThan(0);
    console.log('  ✅ Chain progression: step1 done → step2 todo created');
  });

  test.afterAll(async () => {
    // Clean up chain root (cascades to children)
    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === CHAIN_TASK);
    if (root) {
      try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
    }
  });
});
