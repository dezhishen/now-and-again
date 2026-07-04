/**
 * 回归测试: chain → inspection(branch=chain) → chain
 *
 * 验证: 链步骤为巡检、巡检分支为链时，检查项不丢失，待办可正常完成。
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-Chain-Insp-Chain';

test.describe('Chain → Inspection(branch=Chain)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 chain → inspection(branch=chain)', async () => {
    const res = await api.createTask(familyId, {
      name: ROOT, kind: 'chain',
      extra: {
        steps: [{
          name: ROOT + '-Insp', kind: 'inspection',
          extra: {
            check_items: [{
              name: '检查项',
              branches: [
                { name: 'OK', create_todo: false },
                {
                  name: 'FAIL', create_todo: true,
                  branch_task: {
                    task: { name: ROOT + '-BranchChain', kind: 'chain', schedule_type: 'once', schedule_data: { time: '09:00' } },
                    extra: { steps: [{ name: ROOT + '-BC-S1', kind: 'simple' }] },
                  },
                },
              ],
            }],
          },
        }],
      },
    });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 1500));

    const insp = db.findTask(ROOT + '-Insp');
    expect(insp).toBeTruthy();
    expect(db.getCheckItems(insp.id).length).toBeGreaterThanOrEqual(1);
    console.log('  ✅ Created with check_items');
  });

  test('STEP-2: 触发chain→完成root→巡检待办的extra包含check_items', async () => {
    const root = db.findTask(ROOT);
    const trigRes = await api.triggerTask(familyId, root.id);
    expect([200, 201]).toContain(trigRes.status);
    await new Promise(r => setTimeout(r, 500));

    const rootTodos = db.getTodos(root.id);
    const rootPending = rootTodos.find((t: any) => t.status === 'pending');
    expect(rootPending).toBeTruthy();
    const compRes = await api.completeTodo(familyId, rootPending.id, 'done');
    expect([200, 201]).toContain(compRes.status);
    await new Promise(r => setTimeout(r, 800));

    const insp = db.findTask(ROOT + '-Insp');
    expect(insp).toBeTruthy();
    expect(db.getCheckItems(insp.id).length).toBeGreaterThanOrEqual(1);

    // 巡检待办的 extra 应有 check_items（模拟前端打开巡检弹窗）
    const inspTodos = db.getTodos(insp.id);
    const inspPending = inspTodos.find((t: any) => t.status === 'pending');
    expect(inspPending).toBeTruthy();

    const token = api.getToken();
    const res = await fetch(`http://localhost:8080/api/todos/${inspPending.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extraData = body?.data?.extra || body?.extra;
    expect(extraData).toBeTruthy();
    expect(extraData.check_items).toBeTruthy();
    expect(extraData.check_items.length).toBeGreaterThanOrEqual(1);
    console.log('  ✅ Inspection todo extra has check_items');
  });

  test('STEP-3: 完成巡检(FAIL)→分支链生成待办→检查项不丢失', async () => {
    const insp = db.findTask(ROOT + '-Insp');
    const inspTodos = db.getTodos(insp.id);
    const inspPending = inspTodos.find((t: any) => t.status === 'pending');
    expect(inspPending).toBeTruthy();

    const cis = db.getCheckItems(insp.id);
    const ci = cis[0];
    const branches = db.getBranches(ci.id);
    const failBranch = branches.find((b: any) => b.name === 'FAIL');
    expect(failBranch).toBeTruthy();

    const compRes = await api.completeTodoRaw(familyId, inspPending.id, {
      todo: { status: 'done' },
      extra: { selections: [{ item_id: ci.id, item_name: '检查项', branch_id: failBranch.id, branch_name: 'FAIL' }] },
    });
    expect([200, 201]).toContain(compRes.status);
    await new Promise(r => setTimeout(r, 800));

    // 检查项仍存在
    expect(db.getCheckItems(insp.id).length).toBeGreaterThanOrEqual(1);

    // 分支链生成了待办
    const branchChain = db.findTask(ROOT + '-BranchChain');
    expect(branchChain).toBeTruthy();
    const bcTodos = db.getTodos(branchChain.id);
    expect(bcTodos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ Check items survived, branch chain todo generated');
  });

  test.afterAll(async () => {
    const t = db.findTask(ROOT);
    if (t) try { await api.deleteTask(familyId, t.id); } catch { /* ok */ }
  });
});
