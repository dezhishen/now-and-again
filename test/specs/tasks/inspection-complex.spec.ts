/**
 * Inspection 任务 — 复杂测试
 *
 * 覆盖：
 *   inspection → inspection → simple         （自嵌套 + 引 simple）
 *   inspection → chain → simple → simple     （引 chain 复杂类型）
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-巡检-复杂';

test.describe('Inspection (复杂: 自嵌套+引chain+引simple)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 inspection（含 inspection→inspection→simple + inspection→chain→simple）', async () => {
    const res = await api.createTask(familyId, {
      name: ROOT, kind: 'inspection',
      extra: {
        check_items: [
          // ── 检查项A: 自嵌套 inspection → inspection → simple ──
          {
            name: '自嵌套',
            branches: [
              { name: '合格', result: 'pass', create_todo: false },
              {
                name: '不合格', result: 'fail', create_todo: true,
                branch_task: {
                  task: { name: ROOT + '-L2', kind: 'inspection', schedule_type: 'once', schedule_data: { time: '09:00' } },
                  extra: {
                    check_items: [{
                      name: '二级',
                      branches: [
                        { name: '合格', result: 'pass', create_todo: false },
                        { name: '不合格', result: 'fail', create_todo: true, branch_task: { task: { name: ROOT + '-L3', kind: 'simple', schedule_type: 'once', schedule_data: { time: '09:00' } } } },
                      ],
                    }],
                  },
                },
              },
            ],
          },
          // ── 检查项B: 引用 chain ──
          {
            name: '引Chain',
            branches: [
              { name: '合格', result: 'pass', create_todo: false },
              {
                name: '不合格', result: 'fail', create_todo: true,
                branch_task: {
                  task: { name: ROOT + '-Chain', kind: 'chain', schedule_type: 'once', schedule_data: { time: '09:00' } },
                  extra: {
                    steps: [
                      { name: ROOT + '-Chain-S1', kind: 'simple' },
                      { name: ROOT + '-Chain-S2', kind: 'simple' },
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
    await new Promise(r => setTimeout(r, 1500));
  });

  test('STEP-2: 验证任务树 Kind/CreatedByKind', async () => {
    const tasks = db.getTasks();

    // Root
    const root = tasks.find((t: any) => t.name === ROOT);
    expect(root).toBeTruthy();
    expect(root.kind).toBe('inspection');
    expect(root.created_by_kind).toBe('simple');
    console.log('  ✅ Root: kind=inspection, created_by=simple');

    // 自嵌套分支
    const l2 = tasks.find((t: any) => t.name === ROOT + '-L2');
    expect(l2).toBeTruthy();
    expect(l2.kind).toBe('inspection');
    expect(l2.created_by_kind).toBe('inspection');
    console.log('  ✅ L2: kind=inspection, created_by=inspection  ← 自嵌套');

    const l3 = tasks.find((t: any) => t.name === ROOT + '-L3');
    expect(l3).toBeTruthy();
    expect(l3.kind).toBe('simple');
    expect(l3.created_by_kind).toBe('inspection');
    console.log('  ✅ L3: kind=simple, created_by=inspection   ← 引 simple');

    // 引 Chain 分支
    const chain = tasks.find((t: any) => t.name === ROOT + '-Chain');
    expect(chain).toBeTruthy();
    expect(chain.kind).toBe('chain');
    expect(chain.created_by_kind).toBe('inspection');   // ← inspection 引 chain
    console.log('  ✅ Chain: kind=chain, created_by=inspection  ← 引 chain 类型');

    const chainS1 = tasks.find((t: any) => t.name === ROOT + '-Chain-S1');
    expect(chainS1).toBeTruthy();
    expect(chainS1.kind).toBe('simple');
    expect(chainS1.created_by_kind).toBe('chain');
    console.log('  ✅ Chain-S1: kind=simple, created_by=chain');

    const chainS2 = tasks.find((t: any) => t.name === ROOT + '-Chain-S2');
    expect(chainS2).toBeTruthy();
    expect(chainS2.kind).toBe('simple');
    expect(chainS2.created_by_kind).toBe('chain');
    console.log('  ✅ Chain-S2: kind=simple, created_by=chain');

    const ours = tasks.filter((t: any) => t.name && (t.name as string).startsWith('E2E-巡检-复杂'));
    console.log(`  📊 Total: ${ours.length} (expected >= 6)`);
    expect(ours.length).toBeGreaterThanOrEqual(6);
  });

  test('STEP-3: 验证 GetExtra 递归到叶子节点（自嵌套 + 引chain 全链路）', async () => {
    const token = api.getToken();
    const root = db.findTask(ROOT);
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra ?? body?.extra ?? null;
    expect(extra.check_items.length).toBe(2);

    // ═══ 自嵌套分支: inspection → inspection → simple ═══
    const selfNest = extra.check_items.find((ci: any) => ci.name === '自嵌套');
    expect(selfNest.branches.length).toBe(2);

    // 合格分支
    const snOk = selfNest.branches.find((b: any) => b.name === '合格');
    expect(snOk.create_todo).toBe(false);

    // 不合格分支 → L2(inspection)
    const snFail = selfNest.branches.find((b: any) => b.name === '不合格');
    expect(snFail.create_todo).toBe(true);
    expect(snFail.branch_task.task.name).toBe(ROOT + '-L2');
    expect(snFail.branch_task.task.kind).toBe('inspection');

    // L2 的 extra: check_items → 二级 → 不合格 → L3(simple 叶子)
    const l2Extra = snFail.branch_task.extra;
    expect(l2Extra.check_items.length).toBe(1);
    expect(l2Extra.check_items[0].name).toBe('二级');
    expect(l2Extra.check_items[0].branches.length).toBe(2);

    const l2Ok = l2Extra.check_items[0].branches.find((b: any) => b.name === '合格');
    expect(l2Ok.create_todo).toBe(false);

    const l2Fail = l2Extra.check_items[0].branches.find((b: any) => b.name === '不合格');
    expect(l2Fail.create_todo).toBe(true);
    expect(l2Fail.branch_task.task.name).toBe(ROOT + '-L3');
    expect(l2Fail.branch_task.task.kind).toBe('simple');
    expect(l2Fail.branch_task.extra ?? null).toBeNull(); // simple 叶子
    console.log('  ✅ 自嵌套: inspection→inspection→simple 全链路 — 4层递归到叶子');

    // ═══ 引Chain 分支: inspection → chain → simple → simple ═══
    const chainRef = extra.check_items.find((ci: any) => ci.name === '引Chain');
    const crOk = chainRef.branches.find((b: any) => b.name === '合格');
    expect(crOk.create_todo).toBe(false);

    const crFail = chainRef.branches.find((b: any) => b.name === '不合格');
    expect(crFail.create_todo).toBe(true);
    expect(crFail.branch_task.task.name).toBe(ROOT + '-Chain');
    expect(crFail.branch_task.task.kind).toBe('chain');

    // Chain 的 steps
    const chainSteps = crFail.branch_task.extra.steps;
    expect(chainSteps.length).toBe(2);
    const chainExpected = [ROOT + '-Chain-S1', ROOT + '-Chain-S2'];
    for (let i = 0; i < 2; i++) {
      const cs = chainSteps[i];
      expect(cs.name).toBe(chainExpected[i]);
      expect(cs.kind).toBe('simple');
      expect(cs.child_task_id).toBeTruthy();
      expect(cs.extra ?? null).toBeNull();
    }
    console.log('  ✅ 引Chain: chain→simple×2 全链路 — 每个step.name/kind/child_task_id 验证');
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
  });
});
