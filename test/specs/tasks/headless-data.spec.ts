import { test, expect } from '@playwright/test';
import { db } from '../../utils/db';
import * as api from '../../utils/api';

function uniqueName(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
}

async function ensureAdminAccount() {
  try {
    await fetch('http://localhost:8080/api/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        display_name: '管理员',
        username: 'admin',
        email: 'admin@local.test',
        password: '12345678',
      }),
    });
  } catch {
    // Ignore setup failure; login below will surface real issue if any.
  }
}

function extractFamilies(payload: any): any[] {
  if (Array.isArray(payload)) return payload;
  if (Array.isArray(payload?.data)) return payload.data;
  if (Array.isArray(payload?.data?.families)) return payload.data.families;
  if (Array.isArray(payload?.families)) return payload.families;
  return [];
}

async function ensureAuthAndFamily(): Promise<string> {
  await ensureAdminAccount();

  const loginRes = await api.login('admin', '12345678');
  expect(loginRes.status).toBe(200);
  expect(loginRes.token).toBeTruthy();

  const listRes = await api.listMyFamilies();
  const families = extractFamilies(listRes.data);
  if (families.length > 0) {
    const id = families[0].id;
    api.setFamilyId(id);
    return id;
  }

  const created = await api.createFamily(uniqueName('E2E-Data-Family'));
  expect([200, 201]).toContain(created.status);
  expect(created.familyId).toBeTruthy();
  return created.familyId;
}

async function waitFor<T>(fn: () => T | null | undefined, timeoutMs: number = 8000): Promise<T> {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = fn();
    if (value) return value;
    await new Promise((r) => setTimeout(r, 200));
  }
  throw new Error('waitFor timeout');
}

test.describe.configure({ mode: 'serial' });

test.describe('任务与待办（无头数据覆盖）', () => {
  let familyId = '';

  test.beforeEach(async () => {
    familyId = await ensureAuthAndFamily();
  });

  test('simple: 创建+触发+完成，校验 owner_kind 与 todo 状态', async () => {
    const taskName = uniqueName('HL-Simple');
    const res = await api.createTask(familyId, { name: taskName, kind: 'simple' });
    expect([200, 201]).toContain(res.status);

    const task = await waitFor(() => db.findTask(taskName));
    expect(task.kind).toBe('simple');
    expect(task.owner_kind).toBe('simple');

    const trigger = await api.triggerTask(familyId, task.id);
    expect([200, 201]).toContain(trigger.status);

    const pending = await waitFor(() => db.getTodos(task.id).find((t: any) => t.status === 'pending'));
    expect(pending).toBeTruthy();

    const doneRes = await api.completeTodo(familyId, pending.id, 'done', 'headless-simple');
    expect([200, 201]).toContain(doneRes.status);

    const done = await waitFor(() => db.getTodos(task.id).find((t: any) => t.status === 'done'));
    expect(done).toBeTruthy();
  });

  test('inspection: 创建+触发，校验 owner_kind 与 pending todo', async () => {
    const taskName = uniqueName('HL-Inspect');
    const res = await api.createTask(familyId, {
      name: taskName,
      kind: 'inspection',
      extra: {
        check_items: [
          {
            name: '巡检项-HL-A',
            branches: [
              { name: '正常', create_todo: false },
              { name: '异常', create_todo: false },
            ],
          },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);

    const task = await waitFor(() => db.findTask(taskName));
    expect(task.kind).toBe('inspection');
    expect(['simple', 'inspection']).toContain(task.owner_kind);

    const trigger = await api.triggerTask(familyId, task.id);
    expect([200, 201]).toContain(trigger.status);

    const pending = await waitFor(() => db.getTodos(task.id).find((t: any) => t.status === 'pending'));
    expect(pending).toBeTruthy();
  });

  test('inspection: 完成后写入 inspection_results，todo 置 done', async () => {
    const taskName = uniqueName('HL-Inspect-Deep');
    const res = await api.createTask(familyId, {
      name: taskName,
      kind: 'inspection',
      extra: {
        check_items: [
          {
            name: '巡检项-HL-Deep',
            branches: [
              { name: '正常', create_todo: false },
              { name: '异常', create_todo: false },
            ],
          },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);

    const task = await waitFor(() => db.findTask(taskName));
    const checkItem = await waitFor(() => db.getCheckItems(task.id).find((i: any) => i.name === '巡检项-HL-Deep'));
    const branch = await waitFor(() => db.getBranches(checkItem.id).find((b: any) => b.name === '正常'));

    const trigger = await api.triggerTask(familyId, task.id);
    expect([200, 201]).toContain(trigger.status);

    const pending = await waitFor(() => db.getTodos(task.id).find((t: any) => t.status === 'pending'));

    const complete = await api.completeTodoRaw(familyId, pending.id, {
      todo: { status: 'done', remark: 'headless-inspection' },
      extra: {
        selections: [
          {
            item_id: checkItem.id,
            branch_id: branch.id,
            item_name: checkItem.name,
            branch_name: branch.name,
          },
        ],
        display_summary: `${checkItem.name}:${branch.name}`,
      },
    });
    expect([200, 201]).toContain(complete.status);

    const results = await waitFor(() => db.getInspectionResults(task.id));
    expect(results.some((r: any) => r.item_name === checkItem.name && r.branch_name === '正常')).toBeTruthy();
    expect(db.getTodos(task.id).some((t: any) => t.status === 'done')).toBeTruthy();
  });

  test('chain: 创建+触发，校验 root owner_kind 与 pending todo', async () => {
    const taskName = uniqueName('HL-Chain');
    const res = await api.createTask(familyId, {
      name: taskName,
      kind: 'chain',
      extra: {
        steps: [
          { name: 'HL-链路-步骤1', kind: 'simple' },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);

    const root = await waitFor(() => db.findTask(taskName));
    expect(root.kind).toBe('chain');
    expect(root.owner_kind).toBe('chain');

    const trigger = await api.triggerTask(familyId, root.id);
    expect([200, 201]).toContain(trigger.status);

    const pending = await waitFor(() => db.getTodos(root.id).find((t: any) => t.status === 'pending'));
    expect(pending).toBeTruthy();
  });

  test('chain: 子步骤任务应继承 owner_kind=chain', async () => {
    const taskName = uniqueName('HL-Chain-Owner');
    const res = await api.createTask(familyId, {
      name: taskName,
      kind: 'chain',
      extra: {
        steps: [
          { name: 'HL-所有权-步骤1', kind: 'simple' },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);

    const root = await waitFor(() => db.findTask(taskName));
    expect(root.owner_kind).toBe('chain');

    const steps = await waitFor(() => {
      const arr = db.getChainSteps(root.id);
      return arr.length > 0 ? arr : null;
    });

    const firstStepTask = db.findTaskById(steps[0].child_task_id);
    expect(firstStepTask).toBeTruthy();
    expect(firstStepTask.kind).toBe('simple');
    expect(firstStepTask.owner_kind).toBe('chain');
  });

  test('chain: 完成 root 与 step1 后应生成 step2 pending todo', async () => {
    const taskName = uniqueName('HL-Chain-Next');
    const res = await api.createTask(familyId, {
      name: taskName,
      kind: 'chain',
      extra: {
        steps: [
          { name: 'HL-链路-step1', kind: 'simple' },
          { name: 'HL-链路-step2', kind: 'simple' },
        ],
      },
    });
    expect([200, 201]).toContain(res.status);

    const root = await waitFor(() => db.findTask(taskName));
    const steps = await waitFor(() => {
      const arr = db.getChainSteps(root.id);
      return arr.length === 2 ? arr : null;
    });

    const step1 = db.findTaskById(steps[0].child_task_id);
    const step2 = db.findTaskById(steps[1].child_task_id);
    expect(step1).toBeTruthy();
    expect(step2).toBeTruthy();

    const trigger = await api.triggerTask(familyId, root.id);
    expect([200, 201]).toContain(trigger.status);

    const rootPending = await waitFor(() => db.getTodos(root.id).find((t: any) => t.status === 'pending'));
    const rootDone = await api.completeTodo(familyId, rootPending.id, 'done', 'root done');
    expect([200, 201]).toContain(rootDone.status);

    const step1Pending = await waitFor(() => db.getTodos(step1.id).find((t: any) => t.status === 'pending'));
    const step1Done = await api.completeTodo(familyId, step1Pending.id, 'done', 'step1 done');
    expect([200, 201]).toContain(step1Done.status);

    const step2Pending = await waitFor(() => db.getTodos(step2.id).find((t: any) => t.status === 'pending'));
    expect(step2Pending).toBeTruthy();
  });

  test('chain: 更新 inspection 步骤后应持久化最新 check_items 数据', async () => {
    const taskName = uniqueName('HL-Chain-Inspect-Update');
    const create = await api.createTask(familyId, {
      name: taskName,
      kind: 'chain',
      extra: {
        steps: [
          {
            name: 'HL-Inspect-Step',
            kind: 'inspection',
            extra: {
              check_items: [
                {
                  name: '检查项-A',
                  branches: [
                    { name: '正常', create_todo: false },
                    { name: '异常', create_todo: false },
                  ],
                },
              ],
            },
          },
        ],
      },
    });
    expect([200, 201]).toContain(create.status);

    const root = await waitFor(() => db.findTask(taskName));
    const oldSteps = await waitFor(() => {
      const arr = db.getChainSteps(root.id);
      return arr.length > 0 ? arr : null;
    });
    const oldStepTaskId = oldSteps[0].child_task_id;

    const oldItems = db.getCheckItems(oldStepTaskId);
    expect(oldItems.some((i: any) => i.name === '检查项-A')).toBeTruthy();

    const update = await api.updateTask(familyId, root.id, {
      name: taskName,
      kind: 'chain',
      extra: {
        steps: [
          {
            name: 'HL-Inspect-Step',
            kind: 'inspection',
            extra: {
              check_items: [
                {
                  name: '检查项-B',
                  branches: [
                    { name: '正常', create_todo: false },
                    { name: '异常', create_todo: false },
                  ],
                },
              ],
            },
          },
        ],
      },
    });
    expect([200, 201]).toContain(update.status);

    const newSteps = await waitFor(() => {
      const arr = db.getChainSteps(root.id);
      return arr.length > 0 ? arr : null;
    });
    const newStepTaskId = newSteps[0].child_task_id;

    const newItems = await waitFor(() => db.getCheckItems(newStepTaskId));
    expect(newItems.some((i: any) => i.name === '检查项-B')).toBeTruthy();
  });

  test('inspection 分支触发 chain: 应按顺序推进采购链步骤', async () => {
    const taskName = uniqueName('HL-Inspect-Branch-Chain');
    const res = await api.createTask(familyId, {
      name: taskName,
      kind: 'inspection',
      extra: {
        check_items: [
          {
            name: '家庭消耗品',
            branches: [
              { name: '正常', create_todo: false },
              {
                name: '异常',
                create_todo: true,
                branch_task: {
                  task: {
                    name: uniqueName('采购任务链'),
                    kind: 'chain',
                    schedule_type: 'once',
                    schedule_data: { time: '09:00' },
                  },
                  extra: {
                    steps: [
                      { name: '购买', kind: 'simple' },
                      { name: '获取取件码', kind: 'simple' },
                      { name: '取件', kind: 'simple' },
                      { name: '收纳', kind: 'simple' },
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

    const task = await waitFor(() => db.findTask(taskName));
    const checkItem = await waitFor(() => db.getCheckItems(task.id).find((i: any) => i.name === '家庭消耗品'));
    const abnormal = await waitFor(() => db.getBranches(checkItem.id).find((b: any) => b.name === '异常'));
    expect(abnormal.branch_task_id).toBeTruthy();

    const trigger = await api.triggerTask(familyId, task.id);
    expect([200, 201]).toContain(trigger.status);
    const inspectPending = await waitFor(() => db.getTodos(task.id).find((t: any) => t.status === 'pending'));

    const completeInspection = await api.completeTodoRaw(familyId, inspectPending.id, {
      todo: { status: 'done', remark: '触发采购链' },
      extra: {
        selections: [
          {
            item_id: checkItem.id,
            branch_id: abnormal.id,
            item_name: checkItem.name,
            branch_name: abnormal.name,
          },
        ],
      },
    });
    expect([200, 201]).toContain(completeInspection.status);

    const branchRootTask = await waitFor(() => db.findTaskById(abnormal.branch_task_id));
    expect(branchRootTask.kind).toBe('chain');
    expect(branchRootTask.owner_kind).toBe('chain');

    const chainSteps = await waitFor(() => {
      const arr = db.getChainSteps(branchRootTask.id);
      return arr.length === 4 ? arr : null;
    });
    const s1 = db.findTaskById(chainSteps[0].child_task_id);
    const s2 = db.findTaskById(chainSteps[1].child_task_id);
    expect(s1?.name).toBe('购买');
    expect(s2?.name).toBe('获取取件码');

    // Branch chain starts with a root pending todo.
    const rootPending = await waitFor(() => db.getTodos(branchRootTask.id).find((t: any) => t.status === 'pending'));
    expect(rootPending).toBeTruthy();

    // Complete root -> should generate step1 pending.
    const rootDone = await api.completeTodo(familyId, rootPending.id, 'done', '开始采购流程');
    expect([200, 201]).toContain(rootDone.status);
    const s1Pending = await waitFor(() => db.getTodos(s1.id).find((t: any) => t.status === 'pending'));
    expect(s1Pending).toBeTruthy();

    // Complete step1 -> should generate step2 pending.
    const s1Done = await api.completeTodo(familyId, s1Pending.id, 'done', '已购买');
    expect([200, 201]).toContain(s1Done.status);
    const s2Pending = await waitFor(() => db.getTodos(s2.id).find((t: any) => t.status === 'pending'));
    expect(s2Pending).toBeTruthy();
  });

  test('template(chain): 渲染模板后创建任务，应生成链步骤并可触发', async () => {
    const tplCode = uniqueName('hl_chain_tpl').toLowerCase().replace(/[^a-z0-9_\-]/g, '_');
    const taskName = uniqueName('HL-Template-Chain-Task');

    const createTpl = await api.createFamilyTemplate(familyId, {
      template_code: tplCode,
      name: uniqueName('HL-模板任务链'),
      description: 'headless 模板链路校验',
      kind: 'chain',
      icon: '🧪',
      sort_order: 1,
      enabled: true,
      parameters: [],
      task_defaults: {
        name: taskName,
        kind: 'chain',
        schedule_type: 'daily',
        schedule_data: { time: '09:00' },
        enabled: true,
      },
      extra_schema: {
        steps: [
          { name: '模板步骤-1', kind: 'simple' },
          { name: '模板步骤-2', kind: 'simple' },
        ],
      },
    });
    expect([200, 201]).toContain(createTpl.status);

    const render = await api.renderTaskTemplate(familyId, tplCode, {});
    expect([200, 201]).toContain(render.status);
    const rendered = (render.data as any)?.data || render.data;
    expect(rendered?.task_defaults?.name).toBe(taskName);
    expect(Array.isArray(rendered?.extra_schema?.steps)).toBeTruthy();
    expect(rendered.extra_schema.steps.length).toBe(2);

    const createTaskRes = await api.createTask(familyId, {
      name: rendered.task_defaults.name,
      kind: rendered.task_defaults.kind || 'chain',
      scheduleType: rendered.task_defaults.schedule_type || 'daily',
      scheduleData: rendered.task_defaults.schedule_data || { time: '09:00' },
      groupId: rendered.task_defaults.group_id || '',
      locationId: rendered.task_defaults.location_id || '',
      extra: rendered.extra_schema,
    });
    expect([200, 201]).toContain(createTaskRes.status);

    const root = await waitFor(() => db.findTask(taskName));
    expect(root.kind).toBe('chain');
    expect(root.owner_kind).toBe('chain');

    const steps = await waitFor(() => {
      const arr = db.getChainSteps(root.id);
      return arr.length === 2 ? arr : null;
    });
    expect(steps[0].name).toBe('模板步骤-1');
    expect(steps[1].name).toBe('模板步骤-2');

    const s1 = db.findTaskById(steps[0].child_task_id);
    const s2 = db.findTaskById(steps[1].child_task_id);
    expect(s1?.owner_kind).toBe('chain');
    expect(s2?.owner_kind).toBe('chain');

    const trigger = await api.triggerTask(familyId, root.id);
    expect([200, 201]).toContain(trigger.status);

    const rootPending = await waitFor(() => db.getTodos(root.id).find((t: any) => t.status === 'pending'));
    expect(rootPending).toBeTruthy();
  });
});
