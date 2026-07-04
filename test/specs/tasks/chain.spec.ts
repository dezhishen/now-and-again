/**
 * Chain 任务 — 简单测试（仅嵌套 simple）
 *
 * 场景：chain → simple → simple → simple
 */
import { test, expect } from '@playwright/test';
import { loginViaApi } from '../../fixtures/auth';
import * as api from '../../utils/api';
import { db } from '../../utils/db';

let familyId: string;
const ROOT = 'E2E-链-简单';

test.describe('Chain (简单: 全 simple 步骤)', () => {

  test.beforeAll(async () => {
    await loginViaApi('admin', '12345678');
    const res = await api.listMyFamilies();
    const families = res.data?.data || res.data;
    if (families.length === 0) {
      const cr = await api.createFamily('E2E测试家庭');
      familyId = cr.familyId;
    } else { familyId = families[0].id; api.setFamilyId(familyId); }
  });

  test('STEP-1: 创建 chain（3步全 simple）', async () => {
    const res = await api.createTask(familyId, {
      name: ROOT, kind: 'chain',
      extra: {
        steps: [
          { name: ROOT + '-S1', kind: 'simple' },
          { name: ROOT + '-S2', kind: 'simple' },
          { name: ROOT + '-S3', kind: 'simple' },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);
    await new Promise(r => setTimeout(r, 1000));
  });

  test('STEP-2: 验证 root + 3 步骤 Kind/CreatedByKind + chain_steps DB + 叶子节点数据', async () => {
    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === ROOT);
    expect(root).toBeTruthy();
    expect(root.kind).toBe('chain');
    expect(root.created_by_kind).toBe('chain');
    expect(root.is_root).toBe(1);

    const expectedSteps = [ROOT + '-S1', ROOT + '-S2', ROOT + '-S3'];
    for (const name of expectedSteps) {
      const t = tasks.find((x: any) => x.name === name);
      expect(t).toBeTruthy();
      expect(t.kind).toBe('simple');
      expect(t.created_by_kind).toBe('chain');
      expect(t.is_root).toBe(0);
      expect(t.schedule_type).toBeTruthy();
    }

    // Verify chain_steps table: every row has name, sort_order, kind, child_task_id
    const steps = db.getChainSteps(root.id);
    expect(steps.length).toBe(3);
    for (let i = 0; i < 3; i++) {
      expect(steps[i].name).toBe(expectedSteps[i]);
      expect(steps[i].sort_order).toBe(i);
      expect(steps[i].kind).toBe('simple');
      expect(steps[i].child_task_id).toBeTruthy();
    }
    console.log('  ✅ Root + 3 steps: kind/created_by_kind verified + chain_steps DB + leaf is_root/schedule_type');
  });

  test('STEP-3: 链推进 — 完成 S1 → S2 生成待办', async () => {
    const tasks = db.getTasks();
    const s1 = tasks.find((t: any) => t.name === ROOT + '-S1');
    expect(s1).toBeTruthy();

    // Trigger todo for S1
    const trigRes = await api.triggerTask(familyId, s1.id);
    expect([200, 201]).toContain(trigRes.status);
    await new Promise(r => setTimeout(r, 500));
    const todos = db.getTodos(s1.id);
    const s1Todo = todos.find((t: any) => t.status === 'pending');
    expect(s1Todo).toBeTruthy();

    // Complete S1 → chain progression creates S2 todo
    const compRes = await api.completeTodo(familyId, s1Todo.id, 'done');
    expect([200, 201]).toContain(compRes.status);
    await new Promise(r => setTimeout(r, 800));

    // S1 todo should be done
    const s1TodosAfter = db.getTodos(s1.id);
    expect(s1TodosAfter.filter((t: any) => t.status === 'done').length).toBe(1);

    const s2 = tasks.find((t: any) => t.name === ROOT + '-S2');
    const s2Todos = db.getTodos(s2.id);
    expect(s2Todos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ Chain: S1 done → S2 todo created');
  });

  test('STEP-3b: 链推进 — 完成 S2 → S3 生成待办', async () => {
    const s2 = db.findTask(ROOT + '-S2');
    expect(s2).toBeTruthy();

    // Complete S2 pending
    const s2Todos = db.getTodos(s2.id);
    const s2Pending = s2Todos.find((t: any) => t.status === 'pending');
    expect(s2Pending).toBeTruthy();
    const compRes = await api.completeTodo(familyId, s2Pending.id, 'done');
    expect([200, 201]).toContain(compRes.status);
    await new Promise(r => setTimeout(r, 800));

    // S2 todo should be done
    const s2TodosAfter = db.getTodos(s2.id);
    expect(s2TodosAfter.filter((t: any) => t.status === 'done').length).toBe(1);

    // S3 should have a pending todo
    const s3 = db.findTask(ROOT + '-S3');
    const s3Todos = db.getTodos(s3.id);
    expect(s3Todos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ Chain: S2 done → S3 todo created');
  });

  test('STEP-4: 验证 GetExtra 返回 steps（每个步骤验证到叶子节点）', async () => {
    const token = api.getToken();
    const root = db.findTask(ROOT);
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra ?? body?.extra ?? null;
    expect(extra.steps.length).toBe(3);

    // 验证每个 step 的完整叶子数据
    const expectedNames = [ROOT + '-S1', ROOT + '-S2', ROOT + '-S3'];
    for (let i = 0; i < 3; i++) {
      const s = extra.steps[i];
      expect(s.name).toBe(expectedNames[i]);
      expect(s.kind).toBe('simple');
      expect(s.sort_order).toBe(i);
      expect(s.child_task_id).toBeTruthy();
      expect(s.extra ?? null).toBeNull(); // simple step has no nested extra
    }
    console.log('  ✅ All 3 steps: name/kind/sort_order/child_task_id/extra verified to leaves');
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
  });
});
