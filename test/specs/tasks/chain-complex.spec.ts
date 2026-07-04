/**
 * Chain 任务 — 复杂测试
 *
 * 覆盖：
 *   chain → inspection → inspection → simple     （引 inspection，inspection 自嵌套 + 引 simple）
 *   chain → chain → simple → simple               （自嵌套 chain）
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-链-复杂';

test.describe('Chain (复杂: 引inspection+自嵌套chain)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 chain（含 inspection 步骤 + chain 步骤）', async () => {
    const res = await api.createTask(familyId, {
      name: ROOT, kind: 'chain',
      extra: {
        steps: [
          { name: ROOT + '-S1', kind: 'simple' },
          // ── S2: 引 inspection（inspection 自嵌套 + 引 simple）──
          {
            name: ROOT + '-S2', kind: 'inspection',
            extra: {
              check_items: [{
                name: '巡检S2',
                branches: [
                  { name: '合格', result: 'pass', create_todo: false },
                  {
                    name: '不合格', result: 'fail', create_todo: true,
                    branch_task: {
                      task: { name: ROOT + '-S2-sub', kind: 'inspection', schedule_type: 'once', schedule_data: { time: '09:00' } },
                      extra: { check_items: [{ name: '子巡检', branches: [
                        { name: '合格', result: 'pass', create_todo: false },
                        { name: '不合格', result: 'fail', create_todo: true, branch_task: { task: { name: ROOT + '-S2-leaf', kind: 'simple', schedule_type: 'once', schedule_data: { time: '09:00' } } } },
                      ]}]},
                    },
                  },
                ],
              }],
            },
          },
          // ── S3: 自嵌套 chain ──
          {
            name: ROOT + '-S3', kind: 'chain',
            extra: {
              steps: [
                { name: ROOT + '-S3-A', kind: 'simple' },
                { name: ROOT + '-S3-B', kind: 'simple' },
              ],
            },
          },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 1500));
  });

  test('STEP-2: 验证全任务树 Kind/CreatedByKind', async () => {
    const tasks = db.getTasks();

    // Root
    const root = tasks.find((t: any) => t.name === ROOT);
    expect(root).toBeTruthy();
    expect(root.kind).toBe('chain');
    expect(root.created_by_kind).toBe('chain');
    console.log('  ✅ Root: kind=chain, created_by=chain');

    // S1: simple
    const s1 = tasks.find((t: any) => t.name === ROOT + '-S1');
    expect(s1).toBeTruthy();
    expect(s1.kind).toBe('simple');
    expect(s1.created_by_kind).toBe('chain');

    // S2: inspection → inspection → simple
    const s2 = tasks.find((t: any) => t.name === ROOT + '-S2');
    expect(s2).toBeTruthy();
    expect(s2.kind).toBe('inspection');
    expect(s2.created_by_kind).toBe('chain');
    console.log('  ✅ S2: kind=inspection, created_by=chain  ← 引 inspection');

    const s2sub = tasks.find((t: any) => t.name === ROOT + '-S2-sub');
    expect(s2sub).toBeTruthy();
    expect(s2sub.kind).toBe('inspection');
    expect(s2sub.created_by_kind).toBe('inspection');
    console.log('  ✅ S2-sub: kind=inspection, created_by=inspection  ← 自嵌套');

    const s2leaf = tasks.find((t: any) => t.name === ROOT + '-S2-leaf');
    expect(s2leaf).toBeTruthy();
    expect(s2leaf.kind).toBe('simple');
    expect(s2leaf.created_by_kind).toBe('inspection');
    console.log('  ✅ S2-leaf: kind=simple, created_by=inspection');

    // S3: chain → simple → simple
    const s3 = tasks.find((t: any) => t.name === ROOT + '-S3');
    expect(s3).toBeTruthy();
    expect(s3.kind).toBe('chain');
    expect(s3.created_by_kind).toBe('chain');
    console.log('  ✅ S3: kind=chain, created_by=chain     ← 自嵌套 chain');

    const s3a = tasks.find((t: any) => t.name === ROOT + '-S3-A');
    expect(s3a).toBeTruthy();
    expect(s3a.kind).toBe('simple');
    expect(s3a.created_by_kind).toBe('chain');
    console.log('  ✅ S3-A: kind=simple, created_by=chain');

    const s3b = tasks.find((t: any) => t.name === ROOT + '-S3-B');
    expect(s3b).toBeTruthy();
    expect(s3b.kind).toBe('simple');
    expect(s3b.created_by_kind).toBe('chain');
    console.log('  ✅ S3-B: kind=simple, created_by=chain');

    const ours = tasks.filter((t: any) => t.name && (t.name as string).startsWith('E2E-链-复杂'));
    console.log(`  📊 Total: ${ours.length} (expected >= 8)`);
    expect(ours.length).toBeGreaterThanOrEqual(8);
  });

  test('STEP-3: 验证 GetExtra 递归（含 S2 inspection 嵌套 + S3 chain steps）', async () => {
    const token = api.getToken();
    const root = db.findTask(ROOT);
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra || body?.extra;
    expect(extra.steps.length).toBe(3);

    // S1: simple → extra undefined
    expect(extra.steps[0].extra).toBeFalsy();
    console.log('  ✅ S1 extra=undefined (simple)');

    // S2: inspection → nested extra
    const s2Extra = extra.steps[1].extra;
    expect(s2Extra.check_items).toBeTruthy();
    const failB = s2Extra.check_items[0].branches.find((b: any) => b.name === '不合格');
    expect(failB.branch_task.extra).toBeTruthy();
    expect(failB.branch_task.extra.check_items.length).toBe(1);
    console.log('  ✅ S2 extra → S2-sub check_items present');

    // S3: chain → nested steps
    const s3Extra = extra.steps[2].extra;
    expect(s3Extra.steps).toBeTruthy();
    expect(s3Extra.steps.length).toBe(2);
    expect(s3Extra.steps[0].name).toBe(ROOT + '-S3-A');
    expect(s3Extra.steps[0].kind).toBe('simple');
    console.log('  ✅ S3 extra → nested chain steps[2] present');
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
  });
});
