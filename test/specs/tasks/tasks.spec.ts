/**
 * 任务模块 — simple / inspection 类型任务测试
 *
 * 验证：
 * - 根任务 created_by_kind = "simple"（用户直建）
 * - inspection 子任务 created_by_kind = "inspection"（handler 创建）
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const SIMPLE_TASK = 'E2E-简单任务';
const INSPECTION_TASK = 'E2E-巡检任务';

test.describe('任务模块', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');

    // Get or use existing family
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length > 0) {
      familyId = families[0].id;
      api.setFamilyId(familyId);
    } else {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    }
  });

  // ─── Simple Task ────────────────────────────────────────────────

  test('STEP-1: 通过 API 创建 simple 任务', async () => {
    const res = await api.createTask(familyId, { name: SIMPLE_TASK, kind: 'simple' });
    expect(res.status === 200 || res.status === 201).toBe(true);

    // Verify in DB
    await new Promise(r => setTimeout(r, 500));
    const task = db.findTask(SIMPLE_TASK);
    expect(task).toBeTruthy();
    expect(task.kind).toBe('simple');
  });

  test('STEP-2: 验证根任务 created_by_kind = "simple"', async () => {
    expect(db.assertCreatedByKind(SIMPLE_TASK, 'simple')).toBe(true);
  });

  // ─── Inspection Task ────────────────────────────────────────────

  test('STEP-3: 通过 API 创建 inspection 任务（含子任务）', async () => {
    const res = await api.createTask(familyId, {
      name: INSPECTION_TASK,
      kind: 'inspection',
      extra: {
        check_items: [{
          name: '区域A',
          branches: [
            { name: '合格', result: 'pass', create_todo: false },
            { name: '不合格', result: 'fail', create_todo: true, branch_task: { task: { name: INSPECTION_TASK + '-不合格处理', kind: 'simple' } } },
          ],
        }],
      },
    });
    expect(res.status === 200 || res.status === 201).toBe(true);

    await new Promise(r => setTimeout(r, 500));
    const task = db.findTask(INSPECTION_TASK);
    expect(task).toBeTruthy();
    expect(task.kind).toBe('inspection');
  });

  test('STEP-4: 验证 inspection 根任务 created_by_kind = "simple"', async () => {
    expect(db.assertCreatedByKind(INSPECTION_TASK, 'simple')).toBe(true);
  });

  test('STEP-5: 验证 inspection 子任务 created_by_kind = "inspection"', async () => {
    await new Promise(r => setTimeout(r, 500));

    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === INSPECTION_TASK);
    expect(root).toBeTruthy();

    const children = tasks.filter((t: any) =>
      t.parent_task_id === root.id && t.is_root === 0
    );

    expect(children.length).toBeGreaterThan(0);
    for (const child of children) {
      expect(child.created_by_kind).toBe('inspection');
      console.log('  ✅ Sub-task "' + child.name + '" created_by_kind="' + child.created_by_kind + '"');
    }
  });

  // ─── Cleanup ────────────────────────────────────────────────────

  test.afterAll(async () => {
    // Clean up test tasks via API
    const tasks = db.getTasks();
    for (const t of tasks) {
      if (t.name && (t.name as string).startsWith('E2E-')) {
        try { await api.deleteTask(familyId, t.id as string); } catch { /* ok */ }
      }
    }
  });
});
