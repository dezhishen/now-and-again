import { describe, it, expect } from 'vitest'
import zhCN from '../locales/zh-CN'
import en from '../locales/en'

type DeepRecord = Record<string, unknown>

/** Recursively collect all leaf keys from a nested object, dot-joined. */
function collectKeys(obj: DeepRecord, prefix = ''): string[] {
  const keys: string[] = []
  for (const [key, val] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (val !== null && typeof val === 'object' && !Array.isArray(val)) {
      keys.push(...collectKeys(val as DeepRecord, fullKey))
    } else {
      keys.push(fullKey)
    }
  }
  return keys
}

/** Recursively check that every key in `source` exists in `target`. */
function keysExist(source: DeepRecord, target: DeepRecord, prefix = ''): string[] {
  const missing: string[] = []
  for (const [key, val] of Object.entries(source)) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (!(key in target)) {
      missing.push(fullKey)
    } else if (val !== null && typeof val === 'object' && !Array.isArray(val)) {
      if (target[key] === null || typeof target[key] !== 'object') {
        missing.push(fullKey)
      } else {
        missing.push(...keysExist(val as DeepRecord, (target[key] as DeepRecord), fullKey))
      }
    }
  }
  return missing
}

describe('i18n locale completeness', () => {
  it('zh-CN should have all required keys', () => {
    const keys = collectKeys(zhCN as unknown as DeepRecord)
    expect(keys.length).toBeGreaterThan(0)
    // At minimum, common task kind keys should exist
    expect(keys).toContain('taskKind.simple')
    expect(keys).toContain('taskKind.inspection')
    expect(keys).toContain('taskKind.chain')
  })

  it('en should have all keys that zh-CN has', () => {
    const missing = keysExist(zhCN as unknown as DeepRecord, en as unknown as DeepRecord)
    if (missing.length > 0) {
      console.log('Missing English translations:', missing)
    }
    expect(missing).toEqual([])
  })

  it('zh-CN should have task kind labels', () => {
    const zh = zhCN as unknown as DeepRecord
    const taskKind = zh.taskKind as DeepRecord | undefined
    expect(taskKind).toBeDefined()
    expect(typeof taskKind?.simple).toBe('string')
    expect(typeof taskKind?.inspection).toBe('string')
    expect(typeof taskKind?.chain).toBe('string')
  })

  it('en should have task kind labels', () => {
    const enObj = en as unknown as DeepRecord
    const taskKind = enObj.taskKind as DeepRecord | undefined
    expect(taskKind).toBeDefined()
    expect(typeof taskKind?.simple).toBe('string')
    expect(typeof taskKind?.inspection).toBe('string')
    expect(typeof taskKind?.chain).toBe('string')
  })

  it('should have todo action translations in both locales', () => {
    const zh = zhCN as unknown as DeepRecord
    const enObj = en as unknown as DeepRecord
    const zhTodo = zh.todo as DeepRecord | undefined
    const enTodo = enObj.todo as DeepRecord | undefined
    expect(zhTodo).toBeDefined()
    expect(enTodo).toBeDefined()

    // Check existing todo action keys
    for (const key of ['quickDone', 'remark', 'skip', 'inspect']) {
      expect(typeof (zhTodo as any)?.[key]).toBe('string')
      expect(typeof (enTodo as any)?.[key]).toBe('string')
    }
  })
})

describe('i18n key count parity', () => {
  it('should have same number of top-level keys', () => {
    const zhKeys = Object.keys(zhCN as unknown as DeepRecord).sort()
    const enKeys = Object.keys(en as unknown as DeepRecord).sort()
    expect(zhKeys).toEqual(enKeys)
  })

  it('should have same number of total leaf keys', () => {
    const zhKeys = collectKeys(zhCN as unknown as DeepRecord)
    const enKeys = collectKeys(en as unknown as DeepRecord)
    expect(zhKeys.length).toBe(enKeys.length)
  })
})
