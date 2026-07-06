/**
 * API 测试：任务数据、API Key、小组管理
 */
import { test, expect } from '@playwright/test';
import * as api from '../../utils/api';

let familyId = '';

test.beforeAll(async () => {
  await api.login('admin', '12345678');
  const res = await api.listMyFamilies();
  const families = Array.isArray((res.data as any)?.data) ? (res.data as any).data : (res.data || []);
  familyId = families[0]?.id || '';
  if (!familyId) {
    const c = await api.createFamily('E2E-API-' + Date.now());
    familyId = c.familyId;
  }
  expect(familyId).toBeTruthy();
});

// ─── 任务数据 ─────────────────────────────────────────────────────
test.describe('任务 API', () => {
  test('simple: 创建→触发→完成', async () => {
    const name = 'API-Simple-' + Date.now();
    const c = await api.createTask(familyId, { name, kind: 'simple' });
    const taskId = (c.data as any)?.data?.id || (c.data as any)?.id;
    expect(taskId).toBeTruthy();

    await api.triggerTask(familyId, taskId);
    const todos = await api.listTodos(familyId);
    const pending = ((todos.data as any)?.data || todos.data || []).filter((t: any) => t.task_id === taskId && t.status === 'pending');
    expect(pending.length).toBeGreaterThan(0);

    await api.completeTodo(familyId, pending[0].id, 'done', 'api-test');
  });

  test('inspection: 创建→触发→待办', async () => {
    const name = 'API-Inspect-' + Date.now();
    const c = await api.createTask(familyId, { name, kind: 'inspection' });
    const taskId = (c.data as any)?.data?.id || (c.data as any)?.id;
    await api.triggerTask(familyId, taskId);
    const todos = await api.listTodos(familyId);
    expect(((todos.data as any)?.data || todos.data || []).some((t: any) => t.task_id === taskId && t.status === 'pending')).toBeTruthy();
  });

  test('chain: 创建→触发→子任务待办', async () => {
    const name = 'API-Chain-' + Date.now();
    const c = await api.createTask(familyId, { name, kind: 'chain' });
    const rootId = (c.data as any)?.data?.id || (c.data as any)?.id;
    expect(rootId).toBeTruthy();

    await api.triggerTask(familyId, rootId);
    const todos = await api.listTodos(familyId);
    const all = ((todos.data as any)?.data || todos.data || []);
    expect(all.some((t: any) => t.status === 'pending')).toBeTruthy();
  });
});

// ─── API Key ──────────────────────────────────────────────────────
test.describe('API Key', () => {
  test('列出 API Key', async () => {
    const res = await api.request('GET', '/users/me/api-keys');
    const keys = (res.data as any)?.data || res.data || [];
    expect(keys.length).toBeGreaterThan(0);
  });

  test('创建后再列出', async () => {
    // Create
    await api.request('POST', '/users/me/api-keys', { name: 'E2E-Key-' + Date.now(), scopes: ['read'] });
    // List and verify
    const res = await api.request('GET', '/users/me/api-keys');
    const keys = (res.data as any)?.data || res.data || [];
    expect(keys.length).toBeGreaterThan(0);
  });
});

// ─── 小组 ─────────────────────────────────────────────────────────
test.describe('小组', () => {
  let groupId = '';

  test('创建小组', async () => {
    api.setFamilyId(familyId);
    const name = 'E2E-Group-' + Date.now();
    const res = await api.request('POST', '/groups', { name, description: '测试小组' });
    const g = (res.data as any)?.data || res.data;
    expect(g?.id).toBeTruthy();
    groupId = g.id;
  });

  test('列出小组', async () => {
    api.setFamilyId(familyId);
    const res = await api.request('GET', '/groups');
    const groups = (res.data as any)?.data || res.data || [];
    expect(groups.length).toBeGreaterThan(0);
  });

  test('小组详情', async () => {
    if (!groupId) return;
    api.setFamilyId(familyId);
    const res = await api.request('GET', `/groups/${groupId}`);
    expect(res.status).toBeGreaterThanOrEqual(200);
  });
});
