/**
 * 复杂任务链测试 — 4 步链，含 inspection 嵌套
 *
 * 场景：simple → inspection(含子任务) → simple → inspection
 *
 * 验证：
 * - 每步 created_by_kind = "chain"
 * - 每步 Kind = 真实类型
 * - 链式推进跨越 4 步
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-复杂链';

test.describe('复杂任务链 (4步: simple→inspection→simple→inspection)', () => {

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

  test('STEP-1: 创建 4 步复杂链', async () => {
    const res = await api.createTask(familyId, {
      name: ROOT,
      kind: 'chain',
      extra: {
        steps: [
          { name: ROOT + '-S1-简单', kind: 'simple' },
          {
            name: ROOT + '-S2-巡检',
            kind: 'inspection',
            extra: {
              check_items: [{
                name: '质量检查',
                branches: [
                  { name: '合格', result: 'pass', create_todo: false },
                  { name: '不合格', result: 'fail', create_todo: true, branch_task: { task: { name: ROOT + '-S2-返工', kind: 'simple' } } },
                ],
              }],
            },
          },
          { name: ROOT + '-S3-确认', kind: 'simple' },
          {
            name: ROOT + '-S4-终检',
            kind: 'inspection',
            extra: {
              check_items: [{
                name: '终检项',
                branches: [
                  { name: '通过', result: 'pass', create_todo: false },
                  { name: '不通过', result: 'fail', create_todo: false },
                ],
              }],
            },
          },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 1200));
  });

  test('STEP-2: 验证 4 步 kind 和 created_by_kind', async () => {
    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === ROOT);
    expect(root).toBeTruthy();

    // Root chain: created_by_kind should be updated to "chain" by SaveExtra
    expect(root.created_by_kind).toBe('chain');

    const steps = [
      { name: ROOT + '-S1-简单', kind: 'simple' },
      { name: ROOT + '-S2-巡检', kind: 'inspection' },
      { name: ROOT + '-S3-确认', kind: 'simple' },
      { name: ROOT + '-S4-终检', kind: 'inspection' },
    ];

    for (const s of steps) {
      const t = tasks.find((x: any) => x.name === s.name);
      expect(t).toBeTruthy();
      expect(t.kind).toBe(s.kind);
      expect(t.created_by_kind).toBe('chain');
      console.log(`  ✅ ${s.name}: kind=${t.kind} created_by=${t.created_by_kind}`);
    }
  });

  test('STEP-3: 验证 S2 巡检子任务存在', async () => {
    const tasks = db.getTasks();
    const sub = tasks.find((t: any) => t.name === ROOT + '-S2-返工');
    expect(sub).toBeTruthy();
    expect(sub.kind).toBe('simple');
    expect(sub.created_by_kind).toBe('inspection');
    console.log(`  ✅ S2 返工子任务: kind=${sub.kind} created_by=${sub.created_by_kind}`);
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) {
      try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
    }
  });
});
