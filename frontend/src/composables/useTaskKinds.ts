import type { Component } from 'vue'
import type { CheckItem } from '@/types'

export interface TaskKindDef {
  card: Component
  inspectComponent?: Component
  todoInfo?: Component
  todoActions: Component
  formComponent?: Component
  todoBadgeKey?: string
  labelKey: string
  createLabelKey: string
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

export function getDefaultCheckItems(kind: string): CheckItem[] | undefined {
  return kinds[kind]?.defaultCheckItems
}

export function getTaskKinds(): { kind: string; labelKey: string }[] {
  return Object.entries(kinds).map(([kind, def]) => ({ kind, labelKey: def.labelKey }))
}
