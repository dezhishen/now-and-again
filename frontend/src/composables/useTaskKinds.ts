import type { Component } from 'vue'

// ── Plugin kinds MUST NOT leak internal types to the main flow. ──
// All plugin-specific data is typed as `any` / `any[]` here.
// Concrete types (e.g. CheckItem) belong inside the plugin's own files.

export interface TaskKindDef {
  card: Component
  inspectComponent?: Component
  todoInfo?: Component
  todoActions: Component
  formComponent?: Component
  todoBadgeKey?: string
  labelKey: string
  createLabelKey: string
  defaultCheckItems?: any[]
  /** 从 { task, extra } 生成 display_summary。插件自行解析 extra。 */
  buildDisplaySummary?: (taskWithExtra: { task: any; extra: any }) => string
  serializeExtra?: (formData: any[]) => any
  parseExtra?: (extra: any) => any[]
}

const kinds: Record<string, TaskKindDef> = {}

export function registerTaskKind(kind: string, def: TaskKindDef) {
  kinds[kind] = def
}

export function getTaskKind(kind: string): TaskKindDef | undefined {
  return kinds[kind]
}

export function getTaskCard(kind: string): Component | null {
  return kinds[kind]?.card || null
}

export function getTodoActions(kind: string): Component | null {
  return kinds[kind]?.todoActions || null
}

export function getTodoInfo(kind: string): Component | null {
  return kinds[kind]?.todoInfo || null
}

export function getTodoBadgeKey(kind: string): string {
  return kinds[kind]?.todoBadgeKey || ''
}

export function getFormComponent(kind: string): Component | null {
  return kinds[kind]?.formComponent || null
}

export function getInspectComponent(kind: string): Component | null {
  return kinds[kind]?.inspectComponent || null
}

export function getCreateLabelKey(kind: string): string {
  return kinds[kind]?.createLabelKey || 'taskKind.create'
}

export function getDefaultCheckItems(kind: string): any[] | undefined {
  return kinds[kind]?.defaultCheckItems
}

export function buildDisplaySummary(kind: string, taskWithExtra: { task: any; extra: any }): string {
  return kinds[kind]?.buildDisplaySummary?.(taskWithExtra) || ''
}

export function serializeExtra(kind: string, formData: any[]): any {
  return kinds[kind]?.serializeExtra?.(formData)
}

export function parseExtra(kind: string, extra: any): any[] {
  return kinds[kind]?.parseExtra?.(extra) || []
}

export function getTaskKinds(): { kind: string; labelKey: string }[] {
  return Object.entries(kinds).map(([kind, def]) => ({ kind, labelKey: def.labelKey }))
}
