/**
 * Low-level API helpers for backend interaction.
 * Used by fixtures for setup/teardown, and by specs for assertions.
 *
 * All family-scoped endpoints require X-Family-Id header.
 */
const API_BASE = 'http://localhost:8080/api';

let authToken = '';
let currentFamilyId = '';

export function setToken(token: string) { authToken = token; }
export function getToken() { return authToken; }
export function setFamilyId(id: string) { currentFamilyId = id; }
export function getFamilyId() { return currentFamilyId; }

async function request<T = any>(
  method: string,
  path: string,
  body?: any,
): Promise<{ status: number; data: T }> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (authToken) headers['Authorization'] = `Bearer ${authToken}`;
  if (currentFamilyId && !path.startsWith('/auth') && !path.startsWith('/users')) {
    headers['X-Family-Id'] = currentFamilyId;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  let data: any = null;
  const text = await res.text();
  try { data = JSON.parse(text); } catch { data = text; }

  return { status: res.status, data };
}

// ─── Auth ─────────────────────────────────────────────────────────
export async function login(username: string, password: string) {
  const res = await request<any>('POST', '/auth/login', { username, password });
  const token = res.data?.data?.access_token || res.data?.access_token || '';
  if (token) setToken(token);
  return { status: res.status, data: res.data, token };
}

// ─── Families ─────────────────────────────────────────────────────
export async function listMyFamilies() {
  return request('GET', '/users/me/families');
}

export async function createFamily(name: string) {
  const res = await request<any>('POST', '/families', { name });
  const id = res.data?.data?.id || res.data?.id || '';
  if (id) setFamilyId(id);
  return { status: res.status, data: res.data, familyId: id };
}

// ─── Tasks (family-scoped) ────────────────────────────────────────
export async function createTask(familyId: string, task: {
  name: string;
  kind?: string;
  scheduleType?: string;
  scheduleData?: any;
  groupId?: string;
  locationId?: string;
  displaySummary?: string;
  extra?: any;
}) {
  setFamilyId(familyId);
  return request('POST', '/tasks', {
    task: {
      name: task.name,
      kind: task.kind || 'simple',
      schedule_type: task.scheduleType || 'daily',
      schedule_data: task.scheduleData || { time: '09:00' },
      group_id: task.groupId || '',
      location_id: task.locationId || '',
      display_summary: task.displaySummary || '',
    },
    extra: task.extra || null,
  });
}

export async function listTasks(familyId: string) {
  setFamilyId(familyId);
  return request('GET', '/tasks');
}

export async function triggerTask(familyId: string, taskId: string) {
  setFamilyId(familyId);
  return request('POST', `/tasks/${taskId}/trigger`);
}

export async function deleteTask(familyId: string, taskId: string) {
  setFamilyId(familyId);
  return request('DELETE', `/tasks/${taskId}`);
}

// ─── Todos (family-scoped) ────────────────────────────────────────
export async function listTodos(familyId: string) {
  setFamilyId(familyId);
  return request('GET', '/todos');
}

export async function completeTodo(familyId: string, todoId: string, status: string = 'done', remark?: string) {
  setFamilyId(familyId);
  return request('PUT', `/todos/${todoId}`, {
    todo: { status, remark: remark || '' },
    extra: null,
  });
}
