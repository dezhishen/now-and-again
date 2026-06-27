import type { Component } from 'vue'
import type { CheckItem } from '@/types'

export interface TaskKindDef {
  /** Full card component for task list (name + status + location + group + body + buttons) */
  card: Component
  /** Inspect modal content for todo completion (only for kind=inspection) */
  inspectComponent?: Component
  /** Extra info below the todo name (e.g. check items count) */
  todoInfo?: Component
  /** Action buttons on todo card (e.g. 巡检 / 完成 / 跳过) */
  todoActions: Component
  /** Kind-specific form fields (e.g. check items editor) */
  formComponent?: Component
  /** Badge label displayed next to todo name */
  todoBadge?: string
  label: string
  createLabel: string
  defaultCheckItems?: CheckItem[]
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

export function getTodoBadge(kind: string): string {
  return kinds[kind]?.todoBadge || ''
}

export function getFormComponent(kind: string): Component | null {
  return kinds[kind]?.formComponent || null
}

export function getInspectComponent(kind: string): Component | null {
  return kinds[kind]?.inspectComponent || null
}

export function getCreateLabel(kind: string): string {
  return kinds[kind]?.createLabel || '创建任务'
}

export function getDefaultCheckItems(kind: string): CheckItem[] | undefined {
  return kinds[kind]?.defaultCheckItems
}

export function getTaskKinds(): { kind: string; label: string }[] {
  return Object.entries(kinds).map(([kind, def]) => ({ kind, label: def.label }))
}
