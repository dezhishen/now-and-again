import { describe, it, expect, beforeEach } from 'vitest'
import { defineComponent } from 'vue'
import {
  registerTaskKind,
  getTaskKind,
  getTaskCard,
  getTodoActions,
  getTodoInfo,
  getTodoBadgeKey,
  getFormComponent,
  getInspectComponent,
  getCreateLabelKey,
  getDefaultCheckItems,
  buildDisplaySummary,
  serializeExtra,
  parseExtra,
} from '../useTaskKinds'

const DummyComponent = defineComponent({ template: '<div />' })

beforeEach(() => {
  // Reset the registry by re-registering only the test kind
  // Use a dynamic import approach - actually just re-register
  registerTaskKind('test-kind', {
    card: DummyComponent,
    todoActions: DummyComponent,
    labelKey: 'test.label' as any,
    createLabelKey: 'test.create' as any,
    todoBadgeKey: 'test.badge' as any,
    formComponent: DummyComponent,
    inspectComponent: DummyComponent,
    todoInfo: DummyComponent,
    defaultCheckItems: [{ name: 'item1' }],
    buildDisplaySummary: ({ task }) => `summary: ${task.name}`,
    serializeExtra: (data: any[]) => data.map((d: any) => ({ ...d, serialized: true })),
    parseExtra: (extra: any) => Array.isArray(extra) ? extra : [],
  })
})

describe('registerTaskKind', () => {
  it('should register a task kind', () => {
    const def = getTaskKind('test-kind')
    expect(def).toBeDefined()
    expect(def?.labelKey).toBe('test.label')
  })

  it('should return undefined for unregistered kind', () => {
    const def = getTaskKind('nonexistent')
    expect(def).toBeUndefined()
  })
})

describe('getTaskCard', () => {
  it('should return the card component for registered kind', () => {
    const card = getTaskCard('test-kind')
    expect(card).toBe(DummyComponent)
  })

  it('should return null for unregistered kind', () => {
    expect(getTaskCard('nonexistent')).toBeNull()
  })
})

describe('getTodoActions', () => {
  it('should return the todoActions component', () => {
    expect(getTodoActions('test-kind')).toBe(DummyComponent)
  })

  it('should return null for unregistered kind', () => {
    expect(getTodoActions('nonexistent')).toBeNull()
  })
})

describe('getTodoInfo', () => {
  it('should return the todoInfo component when set', () => {
    expect(getTodoInfo('test-kind')).toBe(DummyComponent)
  })

  it('should return null when not set', () => {
    registerTaskKind('no-info-kind', {
      card: DummyComponent,
      todoActions: DummyComponent,
      labelKey: 'no.info' as any,
      createLabelKey: '' as any,
    })
    expect(getTodoInfo('no-info-kind')).toBeNull()
  })
})

describe('getTodoBadgeKey', () => {
  it('should return the badge key when set', () => {
    expect(getTodoBadgeKey('test-kind')).toBe('test.badge')
  })

  it('should return empty string when not set', () => {
    registerTaskKind('no-badge-kind', {
      card: DummyComponent,
      todoActions: DummyComponent,
      labelKey: 'no.badge' as any,
      createLabelKey: '' as any,
    })
    expect(getTodoBadgeKey('no-badge-kind')).toBe('')
  })
})

describe('getFormComponent', () => {
  it('should return the form component when set', () => {
    expect(getFormComponent('test-kind')).toBe(DummyComponent)
  })

  it('should return null when not set', () => {
    expect(getFormComponent('nonexistent')).toBeNull()
  })
})

describe('getInspectComponent', () => {
  it('should return the inspect component when set', () => {
    expect(getInspectComponent('test-kind')).toBe(DummyComponent)
  })
})

describe('getCreateLabelKey', () => {
  it('should return the create label key when set', () => {
    expect(getCreateLabelKey('test-kind')).toBe('test.create')
  })

  it('should return the default when not set', () => {
    registerTaskKind('no-create-kind', {
      card: DummyComponent,
      todoActions: DummyComponent,
      labelKey: 'no.create' as any,
      createLabelKey: '' as any,
    })
    expect(getCreateLabelKey('no-create-kind')).toBe('taskKind.create')
  })
})

describe('getDefaultCheckItems', () => {
  it('should return default check items when set', () => {
    const items = getDefaultCheckItems('test-kind')
    expect(items).toEqual([{ name: 'item1' }])
  })

  it('should return undefined when not set', () => {
    expect(getDefaultCheckItems('nonexistent')).toBeUndefined()
  })
})

describe('buildDisplaySummary', () => {
  it('should call the registered function', () => {
    const result = buildDisplaySummary('test-kind', { task: { name: '我的任务' }, extra: null })
    expect(result).toBe('summary: 我的任务')
  })

  it('should return empty string when no builder registered', () => {
    const result = buildDisplaySummary('nonexistent', { task: { name: 'x' }, extra: null })
    expect(result).toBe('')
  })
})

describe('serializeExtra', () => {
  it('should call the registered serializer', () => {
    const result = serializeExtra('test-kind', [{ id: 1 }])
    expect(result).toEqual([{ id: 1, serialized: true }])
  })

  it('should return undefined when no serializer registered', () => {
    const data = [{ id: 1 }]
    const result = serializeExtra('nonexistent', data)
    expect(result).toBeUndefined()
  })
})

describe('parseExtra', () => {
  it('should call the registered parser', () => {
    registerTaskKind('parse-kind', {
      card: DummyComponent,
      todoActions: DummyComponent,
      labelKey: 'parse' as any,
      createLabelKey: '' as any,
      parseExtra: (extra: any) => (extra ? [extra] : []),
    })
    const result = parseExtra('parse-kind', { item: 1 })
    expect(result).toEqual([{ item: 1 }])
  })

  it('should return [] when no parser registered', () => {
    const result = parseExtra('nonexistent', null)
    expect(result).toEqual([])
  })
})
