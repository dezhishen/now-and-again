/**
 * Database verification helpers.
 * Uses python3 + sqlite3 for cross-platform DB access.
 */
import { execSync } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const DB_PATH = path.resolve(__dirname, '../../data/now-and-again.db');

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
    return sql('SELECT id, name, kind, created_by_kind, parent_task_id, is_root FROM tasks');
  },

  /** Find task by exact name */
  findTask(name: string): Record<string, any> | null {
    const safe = name.replace(/'/g, "''");
    const rows = sql("SELECT * FROM tasks WHERE name = '" + safe + "'");
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
    return sql("SELECT id, name, create_todo, check_item_id FROM check_item_branches WHERE check_item_id = '" + checkItemId + "'");
  },

  /**
   * Verify a task's created_by_kind field.
   * Logs result and returns true if matches, false otherwise.
   */
  assertCreatedByKind(taskName: string, expectedKind: string): boolean {
    const task = this.findTask(taskName);
    if (!task) {
      console.error('  \u274c Task "' + taskName + '" not found in DB');
      return false;
    }
    if (task.created_by_kind !== expectedKind) {
      console.error(
        '  \u274c Task "' + taskName + '": created_by_kind expected="' +
        expectedKind + '", got="' + task.created_by_kind + '"'
      );
      return false;
    }
    console.log('  \u2705 Task "' + taskName + '" created_by_kind="' + task.created_by_kind + '"');
    return true;
  },
};
