/**
 * 复杂任务更新验证测试
 *
 * 验证：inspection / chain 任务的 update 功能
 * - 更新 check_items：旧子任务删除，新子任务创建
 * - 更新 chain steps：旧步骤+子任务删除，新步骤创建，CreatedByKind 正确
 * - 更新后 GET 一致性
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

test.describe('复杂任务更新', () => {

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
  // Inspection: 更新 check_items
  // ═══════════════════════════════════════════════════════════════

  test('Inspection-Update: 创建→更新 check_items→验证旧子任务删除+新子任务创建', async () => {
    // Create with 1 check item having 1 branch with create_todo
    await api.createTask(familyId, {
      name: 'E2E-Update-Insp', kind: 'inspection',
      extra: {
        check_items: [{
          name: '旧检查项',
          branches: [
            { name: 'OK', create_todo: false },
            { name: 'FAIL', create_todo: true, branch_task: { task: { name: 'E2E-Update-Insp-Old', kind: 'simple' } } },
          ],
        }],
      },
    });
    await new Promise(r => setTimeout(r, 800));
    let task = db.findTask('E2E-Update-Insp');
    expect(task).toBeTruthy();

    // Verify old child task exists
    let oldChild = db.findTask('E2E-Update-Insp-Old');
    expect(oldChild).toBeTruthy();
    expect(oldChild.created_by_kind).toBe('inspection');
    console.log('  ✅ Created with 1 check_item, 1 child task');

    // UPDATE: replace with 2 check items, no create_todo
    const updRes = await api.updateTask(familyId, task.id, {
      name: 'E2E-Update-Insp',
      extra: {
        check_items: [
          { name: '新检查项A', branches: [{ name: 'OK', create_todo: false }] },
          { name: '新检查项B', branches: [{ name: 'OK', create_todo: false }, { name: 'WARN', create_todo: false }] },
        ],
      },
    });
    console.log('  Update response status:', updRes.status);
    if (updRes.status !== 200) console.log('  Update response:', JSON.stringify(updRes.data));
    expect(updRes.status).toBe(200);
    await new Promise(r => setTimeout(r, 1000));

    // Verify old child task is DELETED
    oldChild = db.findTask('E2E-Update-Insp-Old');
    expect(oldChild).toBeNull();
    console.log('  ✅ Old child task deleted');

    // Verify new check_items
    const extra = await getTaskExtra('E2E-Update-Insp');
    expect(extra.check_items.length).toBe(2);
    expect(extra.check_items[0].name).toBe('新检查项A');
    expect(extra.check_items[1].name).toBe('新检查项B');
    expect(extra.check_items[1].branches.length).toBe(2);

    // Verify DB: 2 check_items, 3 branches total, 0 branch_task_id set
    const cis = db.getCheckItems(task.id);
    expect(cis.length).toBe(2);
    let totalBranches = 0;
    for (const ci of cis) {
      const branches = db.getBranches(ci.id);
      totalBranches += branches.length;
      for (const b of branches) {
        expect(b.branch_task_id).toBeFalsy(); // no create_todo
      }
    }
    expect(totalBranches).toBe(3);
    console.log('  ✅ Updated: 2 check_items, 3 branches, 0 child tasks');
  });

  // ═══════════════════════════════════════════════════════════════
  // Chain: 更新 steps
  // ═══════════════════════════════════════════════════════════════

  test('Chain-Update: 创建→更新 steps→验证旧步骤删除+新步骤 CreatedByKind', async () => {
    // Create chain with 2 simple steps
    await api.createTask(familyId, {
      name: 'E2E-Update-Chain', kind: 'chain',
      extra: { steps: [
        { name: 'E2E-Update-Chain-S1', kind: 'simple' },
        { name: 'E2E-Update-Chain-S2', kind: 'simple' },
      ]},
    });
    await new Promise(r => setTimeout(r, 800));
    let task = db.findTask('E2E-Update-Chain');
    expect(task).toBeTruthy();

    // Verify old children exist
    expect(db.findTask('E2E-Update-Chain-S1')).toBeTruthy();
    expect(db.findTask('E2E-Update-Chain-S2')).toBeTruthy();
    console.log('  ✅ Created with 2 steps');

    // UPDATE: replace with 3 steps, one inspection
    const updRes = await api.updateTask(familyId, task.id, {
      name: 'E2E-Update-Chain',
      extra: {
        steps: [
          { name: 'E2E-Update-Chain-New1', kind: 'simple' },
          { name: 'E2E-Update-Chain-New2', kind: 'inspection', extra: { check_items: [
            { name: '巡检', branches: [{ name: 'OK', create_todo: false }] },
          ]}},
          { name: 'E2E-Update-Chain-New3', kind: 'simple' },
        ],
      },
    });
    console.log('  Update response status:', updRes.status);
    if (updRes.status !== 200) console.log('  Update response:', JSON.stringify(updRes.data));
    expect(updRes.status).toBe(200);
    await new Promise(r => setTimeout(r, 1000));

    // Verify old children are DELETED
    expect(db.findTask('E2E-Update-Chain-S1')).toBeNull();
    expect(db.findTask('E2E-Update-Chain-S2')).toBeNull();
    console.log('  ✅ Old chain children deleted');

    // Verify new children exist with correct CreatedByKind
    const s1 = db.findTask('E2E-Update-Chain-New1');
    expect(s1).toBeTruthy();
    expect(s1.kind).toBe('simple');
    expect(s1.created_by_kind).toBe('chain');

    const s2 = db.findTask('E2E-Update-Chain-New2');
    expect(s2).toBeTruthy();
    expect(s2.kind).toBe('inspection');
    expect(s2.created_by_kind).toBe('chain');

    const s3 = db.findTask('E2E-Update-Chain-New3');
    expect(s3).toBeTruthy();
    expect(s3.kind).toBe('simple');
    expect(s3.created_by_kind).toBe('chain');

    console.log('  ✅ Updated: 3 new steps, all created_by=chain');

    // Verify chain_steps table
    const steps = db.getChainSteps(task.id);
    expect(steps.length).toBe(3);
    expect(steps[0].name).toBe('E2E-Update-Chain-New1');
    expect(steps[1].name).toBe('E2E-Update-Chain-New2');
    expect(steps[2].name).toBe('E2E-Update-Chain-New3');
    console.log('  ✅ chain_steps: 3 rows');

    // Verify GET extra
    const extra = await getTaskExtra('E2E-Update-Chain');
    expect(extra.steps.length).toBe(3);

    // New2 (inspection) should have nested extra with check_items
    const s2Extra = extra.steps[1];
    expect(s2Extra.kind).toBe('inspection');
    expect(s2Extra.extra).toBeTruthy();
    expect(s2Extra.extra.check_items.length).toBe(1);
    console.log('  ✅ GetExtra: S2 has nested check_items');
  });

  // ═══════════════════════════════════════════════════════════════
  // Inspection: 更新时新增 create_todo 分支（从无到有）
  // ═══════════════════════════════════════════════════════════════

  test('Inspection-Update-Add: 更新时新增 create_todo 分支 → 验证子任务创建', async () => {
    // Create with NO create_todo
    await api.createTask(familyId, {
      name: 'E2E-Update-Insp-Add', kind: 'inspection',
      extra: {
        check_items: [{
          name: '原检查项',
          branches: [
            { name: 'OK', create_todo: false },
          ],
        }],
      },
    });
    await new Promise(r => setTimeout(r, 800));
    let task = db.findTask('E2E-Update-Insp-Add');
    expect(task).toBeTruthy();

    // No child tasks initially
    expect(db.findTask('E2E-Update-Insp-Add-New')).toBeNull();
    console.log('  ✅ Initially: 1 check_item, 0 child tasks');

    // UPDATE: add a branch with create_todo
    const updRes = await api.updateTask(familyId, task.id, {
      name: 'E2E-Update-Insp-Add',
      extra: {
        check_items: [{
          name: '原检查项',
          branches: [
            { name: 'OK', create_todo: false },
            { name: 'FAIL', create_todo: true, branch_task: { task: { name: 'E2E-Update-Insp-Add-New', kind: 'simple' } } },
          ],
        }],
      },
    });
    expect(updRes.status).toBe(200);
    await new Promise(r => setTimeout(r, 1000));

    // New child task should now exist with complete leaf data
    const newChild = db.findTask('E2E-Update-Insp-Add-New');
    expect(newChild).toBeTruthy();
    expect(newChild.kind).toBe('simple');
    expect(newChild.created_by_kind).toBe('inspection');
    expect(newChild.is_root).toBe(0);
    expect(newChild.schedule_type).toBeTruthy();
    console.log('  ✅ New child task created after update');

    // Verify DB: 1 check_item, 2 branches, FAIL has branch_task_id, OK doesn't
    const cis = db.getCheckItems(task.id);
    expect(cis.length).toBe(1);
    expect(cis[0].name).toBe('原检查项');
    const branches = db.getBranches(cis[0].id);
    expect(branches.length).toBe(2);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');
    expect(failBranch).toBeTruthy();
    expect(failBranch.create_todo).toBe(1);
    expect(failBranch.branch_task_id).toBeTruthy();
    const okBranch = branches.find((b: any) => b.name === 'OK');
    expect(okBranch).toBeTruthy();
    expect(okBranch.create_todo).toBe(0);
    expect(okBranch.branch_task_id).toBeFalsy();
    console.log('  ✅ DB: 1 check_item, 2 branches, FAIL has branch_task_id, OK has none');

    // Verify GetExtra matches
    const extra = await getTaskExtra('E2E-Update-Insp-Add');
    expect(extra.check_items.length).toBe(1);
    expect(extra.check_items[0].name).toBe('原检查项');
    const extraFail = extra.check_items[0].branches.find((b: any) => b.name === 'FAIL');
    expect(extraFail.branch_task.task.name).toBe('E2E-Update-Insp-Add-New');
    expect(extraFail.branch_task.task.kind).toBe('simple');
    console.log('  ✅ GetExtra: FAIL branch_task verified to leaf');
  });

  // ═══════════════════════════════════════════════════════════════
  // Simple: 更新名称（不带 extra）— 验证仅字段更新，不触及 extra
  // ═══════════════════════════════════════════════════════════════

  test('Simple-Update-Name: 仅更新 simple 任务名称，不改变 extra', async () => {
    await api.createTask(familyId, {
      name: 'E2E-Update-Simple-Old', kind: 'simple',
    });
    await new Promise(r => setTimeout(r, 500));
    let task = db.findTask('E2E-Update-Simple-Old');
    expect(task).toBeTruthy();
    expect(task.kind).toBe('simple');

    // Update name only (no extra in body)
    const updRes = await api.updateTask(familyId, task.id, {
      name: 'E2E-Update-Simple-New',
    });
    expect(updRes.status).toBe(200);
    await new Promise(r => setTimeout(r, 500));

    // Old name should be gone, new name should exist
    expect(db.findTask('E2E-Update-Simple-Old')).toBeNull();
    const renamed = db.findTask('E2E-Update-Simple-New');
    expect(renamed).toBeTruthy();
    expect(renamed.kind).toBe('simple');
    expect(renamed.name).toBe('E2E-Update-Simple-New');
    console.log('  ✅ Simple renamed: old gone, new present');

    // Verify no extra created (simple hasn't any)
    const extra = await getTaskExtra('E2E-Update-Simple-New');
    expect(extra).toBeNull();
    console.log('  ✅ Extra still null after name-only update');
  });

  // ─── Cleanup ────────────────────────────────────────────────────

  test.afterAll(async () => {
    for (const name of ['E2E-Update-Insp', 'E2E-Update-Chain', 'E2E-Update-Insp-Add', 'E2E-Update-Simple-Old', 'E2E-Update-Simple-New']) {
      const t = db.findTask(name);
      if (t) try { await api.deleteTask(familyId, t.id); } catch { /* ok */ }
    }
  });
});
