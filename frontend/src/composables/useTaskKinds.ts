import type { Component } from 'vue'
import type { CheckItem } from '@/types'

interface TaskKindDef {
  card: Component
  label: string
  createLabel: string
  defaultCheckItems?: CheckItem[]
}

const kinds: Record<string, TaskKindDef> = {}

export function registerTaskKind(kind: string, def: TaskKindDef) {
  kinds[kind] = def
}

export function getTaskCard(kind: string): Component | null {
  return kinds[kind]?.card || null
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
