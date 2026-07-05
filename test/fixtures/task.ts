/**
 * Task fixture — create/delete tasks for test setup/teardown.
 *
 * Usage:
 *   import { createSimpleTask, createInspectionTask, cleanupTasks } from '@fixtures/task';
 */
import * as api from '../utils/api';
import { db } from '../utils/db';

let createdTaskIds: string[] = [];

/** Track a task for cleanup. */
function track(id: string) {
  createdTaskIds.push(id);
  return id;
}

/**
 * Create a simple task via API.
 * Returns the task object from the API response.
 */
export async function createSimpleTaskViaApi(
  familyId: string,
  name: string,
): Promise<any> {
  const res = await api.createTask(familyId, { name, kind: 'simple' });
  if (res.status === 200 || res.status === 201) {
    track(res.data?.id || res.data?.task?.id);
    return res.data?.task || res.data;
  }
  throw new Error(`Create simple task failed: ${res.status} ${JSON.stringify(res.data)}`);
}

/**
 * Create an inspection task via API.
 */
export async function createInspectionTaskViaApi(
  familyId: string,
  name: string,
  extra?: any,
): Promise<any> {
  const res = await api.createTask(familyId, {
    name,
    kind: 'inspection',
    extra: extra || {
      check_items: [
        {
          name: '区域A',
          branches: [
            { name: '合格', result: 'pass', create_todo: false },
            { name: '不合格', result: 'fail', create_todo: true, task: { name: name + '-不合格处理', kind: 'simple' } },
          ],
        },
      ],
    },
  });
  if (res.status === 200 || res.status === 201) {
    track(res.data?.id || res.data?.task?.id);
    return res.data?.task || res.data;
  }
  throw new Error(`Create inspection task failed: ${res.status} ${JSON.stringify(res.data)}`);
}

/**
 * Verify that a task's owner_kind matches the expected value.
 */
export function assertOwnerKind(taskName: string, expectedKind: string): boolean {
  return db.assertOwnerKind(taskName, expectedKind);
}

/**
 * Clean up all tasks created during this test run.
 */
export async function cleanupTasks(familyId?: string): Promise<void> {
  for (const id of createdTaskIds.reverse()) {
    try { await api.deleteTask(familyId || api.getFamilyId(), id); } catch { /* ignore */ }
  }
  createdTaskIds = [];
}
