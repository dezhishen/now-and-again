/**
 * 复杂嵌套巡检测试
 *
 * 场景：巡检任务含多个检查项，不合格分支创建子巡检任务，
 * 子巡检任务又含检查项，形成 3 层嵌套。
 *
 * 验证：
 * - 每层 created_by_kind 正确传播
 * - 每层子任务 Kind 正确
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-嵌套巡检-根';

test.describe('嵌套巡检 (inspection → inspection → inspection)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else {
      familyId = families[0].id;
      api.setFamilyId(familyId);
    }
  });

  test('STEP-1: 创建 3 层嵌套巡检任务', async () => {
    // Level 1: root inspection with a branch that creates Level 2 inspection
    const res = await api.createTask(familyId, {
      name: ROOT,
      kind: 'inspection',
      extra: {
        check_items: [
          {
            name: '一级检查',
            branches: [
              { name: '合格', result: 'pass', create_todo: false },
              {
                name: '不合格',
                result: 'fail',
                create_todo: true,
                branch_task: {
                  task: {
                    name: ROOT + '-L2',
                    kind: 'inspection',
                  },
                  extra: {
                    check_items: [
                      {
                        name: '二级检查',
                        branches: [
                          { name: '合格', result: 'pass', create_todo: false },
                          {
                            name: '不合格',
                            result: 'fail',
                            create_todo: true,
                            branch_task: {
                              task: {
                                name: ROOT + '-L3',
                                kind: 'simple',
                              },
                            },
                          },
                        ],
                      },
                    ],
                  },
                },
              },
            ],
          },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);

    await new Promise(r => setTimeout(r, 1000));
    const task = db.findTask(ROOT);
    expect(task).toBeTruthy();
    expect(task.kind).toBe('inspection');
  });

  test('STEP-2: 验证 3 层 created_by_kind', async () => {
    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === ROOT);
    expect(root).toBeTruthy();

    // Level 1 (root): user-created → created_by_kind = "simple"
    expect(root.created_by_kind).toBe('simple');
    console.log('  ✅ L1 root: created_by_kind=simple');

    // Level 2: created by inspection handler
    const l2 = tasks.find((t: any) => t.name === ROOT + '-L2');
    expect(l2).toBeTruthy();
    expect(l2.kind).toBe('inspection');
    expect(l2.created_by_kind).toBe('inspection');
    console.log('  ✅ L2: kind=inspection, created_by_kind=inspection');

    // Level 3: created by inspection handler (from L2's branch)
    const l3 = tasks.find((t: any) => t.name === ROOT + '-L3');
    expect(l3).toBeTruthy();
    expect(l3.kind).toBe('simple');
    expect(l3.created_by_kind).toBe('inspection');
    console.log('  ✅ L3: kind=simple, created_by_kind=inspection');
  });

  test('STEP-3: 验证子任务总数', async () => {
    const tasks = db.getTasks();
    // Should have: root + L2 + L3 = at least 3 tasks
    const ours = tasks.filter((t: any) =>
      t.name && (t.name as string).startsWith('E2E-嵌套巡检')
    );
    expect(ours.length).toBeGreaterThanOrEqual(3);
    console.log(`  ✅ Total nested tasks: ${ours.length}`);
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) {
      try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
    }
  });
});
