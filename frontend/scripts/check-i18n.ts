/**
 * Build-time i18n key validation script.
 *
 * Scans all .vue and .ts source files for `t('key')` calls and verifies
 * that every referenced key exists in BOTH locale files.
 *
 * Usage:
 *   npx tsx scripts/check-i18n.ts
 *   node --loader ts-node/esm scripts/check-i18n.ts
 *
 * Exit code 1 if any keys are missing → fails CI/build.
 */

import { readFileSync, readdirSync, statSync } from 'fs'
import { resolve, extname, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const SRC = resolve(__dirname, '..', 'src')

// ─── Load locale files ───────────────────────────────────────────────────────

// We load the TS locale files by reading them as text and extracting the
// default export object with a simple parser, since we can't easily import
// ESM TS files from a plain Node script without a transpiler.
// This is intentionally a lightweight static extractor.

function loadLocaleKeys(filePath: string): Map<string, true> {
  const raw = readFileSync(filePath, 'utf-8')
  const keys = new Map<string, true>()

  // Walk nested object structure: { a: { b: 'val', c: 'val' } } → 'a.b', 'a.c'
  // Uses a simple state-machine JSON-like parser for the export default { ... } block.
  const match = raw.match(/export\s+default\s+(\{[\s\S]*\})\s*$/m)
  if (!match) {
    console.error(`❌ Could not parse locale file: ${filePath}`)
    return keys
  }

  const objStr = match[1]
  // Extract all key paths from the object literal.
  // Strategy: recursively walk the structure by tracking brace depth.
  extractKeys(objStr, '', keys)
  return keys
}

function extractKeys(objStr: string, prefix: string, keys: Map<string, true>): void {
  // Remove the outermost braces
  let inner = objStr.trim()
  if (inner.startsWith('{')) inner = inner.slice(1)
  if (inner.endsWith('}')) inner = inner.slice(0, -1)
  inner = inner.trim()
  if (!inner) return

  // Split by top-level commas (not inside braces)
  const entries = splitTopLevel(inner)
  for (const entry of entries) {
    const colonIdx = findTopLevelColon(entry)
    if (colonIdx < 0) continue

    const rawKey = entry.slice(0, colonIdx).trim()
    // Unquote key
    const key = rawKey.replace(/^['"]|['"]$/g, '').replace(/^(\w+)$/, '$1')
    if (!key) continue

    let value = entry.slice(colonIdx + 1).trim()
    // Remove trailing comma
    if (value.endsWith(',')) value = value.slice(0, -1).trim()

    const fullKey = prefix ? `${prefix}.${key}` : key

    if (value.startsWith('{')) {
      // Nested object → recurse
      extractKeys(value, fullKey, keys)
    } else {
      // Leaf value (string, number, etc.)
      keys.set(fullKey, true)
    }
  }
}

function splitTopLevel(str: string): string[] {
  const parts: string[] = []
  let depth = 0
  let start = 0
  let inString = false
  let stringChar = ''

  for (let i = 0; i < str.length; i++) {
    const ch = str[i]
    if (inString) {
      if (ch === '\\') { i++; continue }
      if (ch === stringChar) inString = false
      continue
    }
    if (ch === "'" || ch === '"' || ch === '`') {
      inString = true
      stringChar = ch
      continue
    }
    if (ch === '{') depth++
    else if (ch === '}') depth--
    else if (ch === ',' && depth === 0) {
      parts.push(str.slice(start, i))
      start = i + 1
    }
  }
  // Last part
  const last = str.slice(start).trim()
  if (last) parts.push(last)

  return parts
}

function findTopLevelColon(str: string): number {
  let depth = 0
  let inString = false
  let stringChar = ''

  for (let i = 0; i < str.length; i++) {
    const ch = str[i]
    if (inString) {
      if (ch === '\\') { i++; continue }
      if (ch === stringChar) inString = false
      continue
    }
    if (ch === "'" || ch === '"' || ch === '`') {
      inString = true
      stringChar = ch
      continue
    }
    if (ch === '{') depth++
    else if (ch === '}') depth--
    else if (ch === ':' && depth === 0) return i
  }
  return -1
}

// ─── Scan source files for t('...') calls ────────────────────────────────────

interface KeyUsage {
  key: string
  file: string
  line: number
}

/** Remove JS/TS comments from source to avoid false positives in examples. */
function stripComments(code: string): string {
  // Remove block comments /* ... */
  code = code.replace(/\/\*[\s\S]*?\*\//g, '')
  // Remove line comments // ... (but preserve URLs like https://)
  code = code.replace(/(?<!https?:)\/\/.*$/gm, '')
  return code
}

function scanSourceFiles(dir: string): KeyUsage[] {
  const usages: KeyUsage[] = []
  walkDir(dir, (filePath) => {
    const ext = extname(filePath)
    if (ext !== '.vue' && ext !== '.ts' && ext !== '.tsx') return
    // Don't scan locale files themselves (they contain values, not key usages)
    if (filePath.includes('/i18n/locales/')) return

    const rawContent = readFileSync(filePath, 'utf-8')
    const content = stripComments(rawContent)
    const lines = rawContent.split('\n')

    // Match t('key'), t("key"), t(`key`) — literal dotted keys only
    const tCallRe = /\bt\s*\(\s*(['"`])((?:\\.|(?!\1).)*?)\1\s*\)/g
    let match: RegExpExecArray | null
    while ((match = tCallRe.exec(content)) !== null) {
      const key = match[2]
      // Must be a dotted i18n key: word.word or word.word.word
      if (!/^\w+\.\w+(\.\w+)*$/.test(key)) continue

      const pos = match.index
      const beforeMatch = rawContent.slice(0, pos)
      const line = beforeMatch.split('\n').length

      usages.push({ key, file: relativePath(filePath), line })
    }

    // Also scan for i18n key patterns in data structures:
    // labelKey: 'schedule.once', groupKey: 'apiKey.scope.family', etc.
    const dsKeyRe = /(?:labelKey|groupKey|badgeKey|todoBadgeKey|createLabelKey)\s*:\s*(['"])(\w+\.\w+(?:\.\w+)*)\1/g
    while ((match = dsKeyRe.exec(content)) !== null) {
      const key = match[2]

      const already = usages.some(u => u.key === key && u.file === relativePath(filePath))
      if (already) continue

      const pos = match.index
      const beforeMatch = rawContent.slice(0, pos)
      const line = beforeMatch.split('\n').length

      usages.push({ key, file: relativePath(filePath), line })
    }

    // Scan ROLE_LABELS-style records: { owner: 'members.role_owner', ... }
    const roleLabelRe = /\b(?:owner|admin|member)\s*:\s*(['"])(\w+\.\w+(?:\.\w+)*)\1/g
    while ((match = roleLabelRe.exec(content)) !== null) {
      const key = match[2]

      const already = usages.some(u => u.key === key && u.file === relativePath(filePath))
      if (already) continue

      const pos = match.index
      const beforeMatch = rawContent.slice(0, pos)
      const line = beforeMatch.split('\n').length

      usages.push({ key, file: relativePath(filePath), line })
    }
  })
  return usages
}

function walkDir(dir: string, cb: (filePath: string) => void): void {
  const entries = readdirSync(dir)
  for (const entry of entries) {
    const full = resolve(dir, entry)
    if (entry === 'node_modules' || entry === 'dist') continue
    try {
      const st = statSync(full)
      if (st.isDirectory()) {
        walkDir(full, cb)
      } else if (st.isFile()) {
        cb(full)
      }
    } catch { /* skip inaccessible files */ }
  }
}

function relativePath(abs: string): string {
  return abs.replace(SRC + '/', 'src/')
}

// ─── Main ─────────────────────────────────────────────────────────────────────

const zhCNPath = resolve(SRC, 'i18n/locales/zh-CN.ts')
const enPath = resolve(SRC, 'i18n/locales/en.ts')

console.log('🔍 Checking i18n keys...\n')

const zhKeys = loadLocaleKeys(zhCNPath)
const enKeys = loadLocaleKeys(enPath)
const allLocaleKeys = new Set([...zhKeys.keys(), ...enKeys.keys()])

console.log(`  zh-CN keys: ${zhKeys.size}`)
console.log(`  en keys:   ${enKeys.size}\n`)

// Check locale files are in sync
const zhOnly = [...zhKeys.keys()].filter(k => !enKeys.has(k))
const enOnly = [...enKeys.keys()].filter(k => !zhKeys.has(k))

if (zhOnly.length > 0) {
  console.log('⚠️  Keys in zh-CN but missing from en:')
  zhOnly.forEach(k => console.log(`     ${k}`))
}
if (enOnly.length > 0) {
  console.log('⚠️  Keys in en but missing from zh-CN:')
  enOnly.forEach(k => console.log(`     ${k}`))
}

// Scan source files
const usages = scanSourceFiles(SRC)

// Deduplicate by key
const uniqueKeys = new Map<string, KeyUsage[]>()
for (const u of usages) {
  if (!uniqueKeys.has(u.key)) uniqueKeys.set(u.key, [])
  uniqueKeys.get(u.key)!.push(u)
}

// Find keys used in source but missing from locale files
const missingKeys: { key: string; files: string[] }[] = []
for (const [key, refs] of uniqueKeys) {
  if (!allLocaleKeys.has(key)) {
    missingKeys.push({
      key,
      files: [...new Set(refs.map(r => `${r.file}:${r.line}`))],
    })
  }
}

if (missingKeys.length > 0) {
  console.log(`\n❌ MISSING i18n keys (${missingKeys.length}):`)
  console.log('   These keys are used in source code but do not exist in locale files:\n')
  for (const m of missingKeys) {
    console.log(`   ${m.key}`)
    for (const f of m.files) {
      console.log(`     → ${f}`)
    }
  }
  console.log('')
}

// Find keys in locale files that are never used
const usedKeys = new Set(uniqueKeys.keys())
const unusedKeys = [...allLocaleKeys].filter(k => !usedKeys.has(k))
if (unusedKeys.length > 0) {
  console.log(`ℹ️  Unused locale keys (${unusedKeys.length}):`)
  unusedKeys.sort().forEach(k => console.log(`     ${k}`))
  console.log('')
}

const hasErrors = missingKeys.length > 0 || zhOnly.length > 0 || enOnly.length > 0
const hasLocaleMismatch = zhOnly.length > 0 || enOnly.length > 0

if (hasErrors) {
  if (hasLocaleMismatch) {
    console.log('❌ Locale files are out of sync!')
  }
  if (missingKeys.length > 0) {
    console.log(`❌ ${missingKeys.length} missing translation key(s) found.`)
    console.log('   Add them to both src/i18n/locales/zh-CN.ts and src/i18n/locales/en.ts')
  }
  process.exit(1)
} else {
  console.log('✅ All i18n keys are valid!')
  process.exit(0)
}
