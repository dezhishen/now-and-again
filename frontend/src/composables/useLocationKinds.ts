import type { Component } from 'vue'

export interface LocationKindDef {
  /** Human-readable label for dropdown */
  label: string
  /** Icon for display */
  icon: string
  /** Optional kind-specific card component (replaces default display) */
  card?: Component
  /** Optional kind-specific form fields */
  formComponent?: Component
}

const kinds: Record<string, LocationKindDef> = {}

export function registerLocationKind(kind: string, def: LocationKindDef) {
  kinds[kind] = def
}

export function getLocationKind(kind: string): LocationKindDef | undefined {
  return kinds[kind]
}

/** Returns all registered kinds for dropdown population. */
export function getLocationKinds(): { kind: string; label: string; icon: string }[] {
  return Object.entries(kinds).map(([kind, def]) => ({
    kind,
    label: def.label,
    icon: def.icon,
  }))
}

export function getLocationKindLabel(kind: string): string {
  return kinds[kind]?.label || kind
}

export function getLocationKindIcon(kind: string): string {
  return kinds[kind]?.icon || ''
}

export function getLocationCard(kind: string): Component | null {
  return kinds[kind]?.card || null
}
