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

  test('STEP-2: 验证 root + 3 步骤 Kind/CreatedByKind', async () => {
    const tasks = db.getTasks();
    const root = tasks.find((t: any) => t.name === ROOT);
    expect(root).toBeTruthy();
    expect(root.kind).toBe('chain');
    expect(root.created_by_kind).toBe('chain');
    console.log('  ✅ Root: kind=chain, created_by=chain');

    for (const name of [ROOT + '-S1', ROOT + '-S2', ROOT + '-S3']) {
      const t = tasks.find((x: any) => x.name === name);
      expect(t).toBeTruthy();
      expect(t.kind).toBe('simple');
      expect(t.created_by_kind).toBe('chain');
      console.log(`  ✅ ${name}: kind=simple, created_by=chain`);
    }
  });

  test('STEP-3: 链推进 — 完成 S1 → S2 生成待办', async () => {
    const tasks = db.getTasks();
    const s1 = tasks.find((t: any) => t.name === ROOT + '-S1');
    expect(s1).toBeTruthy();

    // Trigger todo for S1
    await api.triggerTask(familyId, s1.id);
    await new Promise(r => setTimeout(r, 500));
    const todos = db.getTodos(s1.id);
    const s1Todo = todos.find((t: any) => t.status === 'pending');
    expect(s1Todo).toBeTruthy();

    // Complete S1 → chain progression creates S2 todo
    await api.completeTodo(familyId, s1Todo.id, 'done');
    await new Promise(r => setTimeout(r, 800));

    const s2 = tasks.find((t: any) => t.name === ROOT + '-S2');
    const s2Todos = db.getTodos(s2.id);
    expect(s2Todos.some((t: any) => t.status === 'pending')).toBe(true);
    console.log('  ✅ Chain: S1 done → S2 todo created');
  });

  test('STEP-4: 验证 GetExtra 返回 steps（无嵌套 extra）', async () => {
    const token = api.getToken();
    const root = db.findTask(ROOT);
    const res = await fetch(`http://localhost:8080/api/tasks/${root.id}?with_extra=true`, {
      headers: { Authorization: `Bearer ${token}`, 'X-Family-Id': familyId },
    });
    const body = await res.json();
    const extra = body?.data?.extra || body?.extra;
    expect(extra.steps.length).toBe(3);
    for (const s of extra.steps) {
      expect(s.kind).toBe('simple');
      expect(s.extra).toBeFalsy(); // simple has no extra
    }
    console.log('  ✅ All 3 steps: kind=simple, extra=undefined');
  });

  test.afterAll(async () => {
    const root = db.findTask(ROOT);
    if (root) try { await api.deleteTask(familyId, root.id as string); } catch { /* ok */ }
  });
});
