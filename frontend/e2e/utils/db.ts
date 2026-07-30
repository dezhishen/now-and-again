/**
 * Database verification helpers.
 * Uses python3 + sqlite3 for cross-platform DB access.
 */
import { execSync } from 'child_process';
import { existsSync } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CANDIDATE_DB_PATHS = [
  path.resolve(process.env.NA_DATA_DIR || '../../data', 'now-and-again.db'),
  path.resolve(__dirname, '../../data/now-and-again.db'),
  path.resolve(__dirname, '../../backend/data/now-and-again.db'),
];
const DB_PATH = CANDIDATE_DB_PATHS.find((p) => existsSync(p)) || CANDIDATE_DB_PATHS[0];

function shell(cmd: string): string {
  try {
    return execSync(cmd, { encoding: 'utf-8', stdio: ['pipe', 'pipe', 'pipe'] }).trim();
  } catch {
    return '';
  }
}

/**
 * Run a SQL query via python3 sqlite3 and return rows as objects.
 */
function sql(query: string): any[] {
  // Escape for Python triple-quoted string
  const escaped = query.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
  const py = [
    'import sqlite3, json',
    'conn = sqlite3.connect("' + DB_PATH + '")',
    'conn.row_factory = sqlite3.Row',
    'cursor = conn.execute("""' + escaped + '""")',
    'rows = [dict(r) for r in cursor.fetchall()]',
    'print(json.dumps(rows, default=str))',
    'conn.close()',
  ].join('\n');

  // Escape single quotes for shell -c '...'
  const safe = py.replace(/'/g, "'\\''");
  const output = shell("python3 -c '" + safe + "'");
  try { return output ? JSON.parse(output) : []; } catch { return []; }
}

export const db = {
  /** Get all tasks */
  getTasks() {
    return sql('SELECT id, name, kind, owner_kind, parent_task_id, is_root, schedule_type, enabled FROM tasks');
  },

  /** Find task by exact name */
  findTask(name: string): Record<string, any> | null {
    const safe = name.replace(/'/g, "''");
    const rows = sql("SELECT * FROM tasks WHERE name = '" + safe + "'");
    return rows[0] || null;
  },

  /** Find task by id */
  findTaskById(taskId: string): Record<string, any> | null {
    const safe = taskId.replace(/'/g, "''");
    const rows = sql("SELECT * FROM tasks WHERE id = '" + safe + "'");
    return rows[0] || null;
  },

  /** Get todos for a task */
  getTodos(taskId: string) {
    return sql("SELECT id, status, task_id FROM todos WHERE task_id = '" + taskId + "'");
  },

  /** Get check items for a task */
  getCheckItems(taskId: string) {
    return sql("SELECT id, name, task_id FROM check_items WHERE task_id = '" + taskId + "'");
  },

  /** Get branches for a check item */
  getBranches(checkItemId: string) {
    return sql("SELECT id, name, create_todo, branch_task_id, check_item_id FROM check_item_branches WHERE check_item_id = '" + checkItemId + "'");
  },

  /** Get inspection results for a task */
  getInspectionResults(taskId: string) {
    return sql("SELECT id, task_id, item_name, branch_name FROM inspection_results WHERE task_id = '" + taskId + "'");
  },

  /** Get chain steps for a root task */
  getChainSteps(rootTaskId: string) {
    return sql("SELECT id, task_id, sort_order, name, kind, child_task_id FROM chain_steps WHERE task_id = '" + rootTaskId + "' ORDER BY sort_order");
  },

  /**
   * Verify a task's owner_kind field.
   * Logs result and returns true if matches, false otherwise.
   */
  assertOwnerKind(taskName: string, expectedKind: string): boolean {
    const task = this.findTask(taskName);
    if (!task) {
      console.error('  \u274c Task "' + taskName + '" not found in DB');
      return false;
    }
    if (task.owner_kind !== expectedKind) {
      console.error(
        '  \u274c Task "' + taskName + '": owner_kind expected="' +
        expectedKind + '", got="' + task.owner_kind + '"'
      );
      return false;
    }
    console.log('  \u2705 Task "' + taskName + '" owner_kind="' + task.owner_kind + '"');
    return true;
  },
};
