<script setup lang="ts">
import { ref, onMounted, nextTick, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import type { FloorPlan, Location, Point } from '@/types'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

const floorPlans = ref<FloorPlan[]>([])
const loading = ref(true)
const error = ref('')
const uploading = ref(false)
const showUploadMenu = ref(false)

// ─── Edit modal ──────────────────────────────────────────────────
const editPlan = ref<FloorPlan | null>(null)
const marking = ref(false)
const newLocName = ref('')
const newLocColor = ref('#3b82f6')
const newLocPoint = ref<Point | null>(null)
const showLocInput = ref(false)
const editingLoc = ref<Location | null>(null)
const showLocations = ref(true)
const hoveredLoc = ref<string | null>(null)

const editImgRef = ref<HTMLImageElement | null>(null)
const imgNatural = ref({ w: 0, h: 0 })
const displaySize = ref({ w: 0, h: 0 })

onMounted(async () => { await loadPlans() })

async function loadPlans() {
  try { floorPlans.value = await api.get<FloorPlan[]>('/families/' + familyId + '/floor-plans') } catch { floorPlans.value = [] }
}

async function loadPlanDetail(planId: string) {
  try {
    const plan = await api.get<FloorPlan>('/floor-plans/' + planId)
    const idx = floorPlans.value.findIndex(p => p.id === planId)
    if (idx >= 0) floorPlans.value[idx] = plan
    return plan
  } catch { return null }
}

function openEdit(plan: FloorPlan) {
  editPlan.value = plan
  marking.value = false; newLocPoint.value = null
  showLocInput.value = false; imgNatural.value = { w: 0, h: 0 }
  loadPlanDetail(plan.id).then(p => { if (p) editPlan.value = p })
}

function closeEdit() { editPlan.value = null }

function onEditImgLoad() {
  if (!editImgRef.value) return
  imgNatural.value = { w: editImgRef.value.naturalWidth, h: editImgRef.value.naturalHeight }
  updateDisplaySize()
}

function updateDisplaySize() {
  if (!editImgRef.value) return
  displaySize.value = { w: editImgRef.value.clientWidth, h: editImgRef.value.clientHeight }
}

const resizeHandler = () => updateDisplaySize()
window.addEventListener('resize', resizeHandler)
onUnmounted(() => window.removeEventListener('resize', resizeHandler))

function scaleX(px: number) { return imgNatural.value.w ? (px / imgNatural.value.w) * displaySize.value.w : 0 }
function scaleY(py: number) { return imgNatural.value.h ? (py / imgNatural.value.h) * displaySize.value.h : 0 }
function unscaleX(sx: number) { return displaySize.value.w ? Math.round((sx / displaySize.value.w) * imgNatural.value.w) : 0 }
function unscaleY(sy: number) { return displaySize.value.h ? Math.round((sy / displaySize.value.h) * imgNatural.value.h) : 0 }
function svgOffset(e: MouseEvent): Point {
  const rect = (e.currentTarget as SVGSVGElement).getBoundingClientRect()
  return { x: e.clientX - rect.left, y: e.clientY - rect.top }
}

// ─── Upload ─────────────────────────────────────────────────────

async function uploadFile(e: Event) {
  const input = e.target as HTMLInputElement; const file = input.files?.[0]
  if (!file) return; showUploadMenu.value = false
  const label = floorPlans.value.length > 0 ? `${floorPlans.value.length + 1}F` : undefined
  await doUpload(file, label)
}

async function doUpload(file: File, label?: string, isCover?: boolean) {
  uploading.value = true; error.value = ''
  try {
    const form = new FormData(); form.append('file', file)
    if (label) form.append('label', label)
    if (isCover) form.append('is_cover', 'true')
    const token = api.getAccessToken()
    const res = await fetch('/api/families/' + familyId + '/floor-plans', {
      method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: form,
    })
    if (!res.ok) { const text = await res.text().catch(() => ''); throw new Error(text || `HTTP ${res.status}`) }
    const json = await res.json()
    if (!json.success) throw new Error(json.error || 'Unknown error')
    floorPlans.value.push(json.data); openEdit(json.data)
  } catch (e: any) { error.value = e.message } finally { uploading.value = false }
}

async function deletePlan(plan: FloorPlan) {
  if (!confirm(t('floorPlan.deleteConfirm'))) return
  try {
    await api.delete('/floor-plans/' + plan.id)
    floorPlans.value = floorPlans.value.filter(p => p.id !== plan.id)
    if (editPlan.value?.id === plan.id) editPlan.value = null
  } catch (e: any) { error.value = e.message }
}

async function setAsCover(plan: FloorPlan) {
  error.value = ''
  try {
    await api.put('/floor-plans/' + plan.id + '/cover')
    floorPlans.value.forEach(p => { p.is_cover = p.id === plan.id })
    if (editPlan.value?.id === plan.id) editPlan.value.is_cover = true
  } catch (e: any) { error.value = e.message }
}

// ─── Location marking ───────────────────────────────────────────

function onSvgClick(e: MouseEvent) {
  if (!marking.value) return
  const pos = svgOffset(e)
  newLocPoint.value = { x: unscaleX(pos.x), y: unscaleY(pos.y) }
  newLocName.value = ''
  newLocColor.value = unusedColor()
  editingLoc.value = null
  showLocInput.value = true
  marking.value = false
}

function startEditLoc(loc: Location) {
  editingLoc.value = loc
  newLocName.value = loc.name
  newLocColor.value = loc.color
  newLocPoint.value = null
  showLocInput.value = true
}

async function saveLocation() {
  if (!newLocName.value.trim() || !editPlan.value) return
  error.value = ''

  if (editingLoc.value) {
    try {
      const updated = await api.put<Location>('/locations/' + editingLoc.value.id, {
        name: newLocName.value.trim(),
        color: newLocColor.value,
      })
      const idx = editPlan.value.locations?.findIndex(l => l.id === editingLoc.value!.id)
      if (idx != null && idx >= 0 && editPlan.value.locations) editPlan.value.locations[idx] = updated
    } catch (e: any) { error.value = e.message; return }
  } else if (newLocPoint.value) {
    try {
      const loc = await api.post<Location>('/floor-plans/' + editPlan.value.id + '/locations', {
        name: newLocName.value.trim(),
        point: newLocPoint.value,
        color: newLocColor.value,
      })
      if (!editPlan.value.locations) editPlan.value.locations = []
      editPlan.value.locations.push(loc)
    } catch (e: any) { error.value = e.message; return }
  }

  showLocInput.value = false; newLocPoint.value = null; editingLoc.value = null
}

function cancelLoc() { showLocInput.value = false; newLocPoint.value = null; editingLoc.value = null }

async function deleteLocation(locId: string) {
  error.value = ''
  try {
    await api.delete('/locations/' + locId)
    if (editPlan.value?.locations) editPlan.value.locations = editPlan.value.locations.filter(l => l.id !== locId)
  } catch (e: any) { error.value = e.message }
}

const PRESET_COLORS = ['#3b82f6','#ef4444','#22c55e','#f59e0b','#8b5cf6','#ec4899','#06b6d4','#78716c']

function unusedColor(): string {
  const used = new Set((editPlan.value?.locations || []).map(l => l.color))
  if (editingLoc.value) used.delete(editingLoc.value.color)
  return PRESET_COLORS.find(c => !used.has(c)) || PRESET_COLORS[0]
}

// ─── Canvas Drawing ──────────────────────────────────────────────

const showDrawer = ref(false)
const drawCanvas = ref<HTMLCanvasElement | null>(null)
const drawCtx = ref<CanvasRenderingContext2D | null>(null)
const drawPoints = ref<Point[]>([])
const drawLines = ref<{ from: number; to: number }[]>([])
const lastDrawPos = ref<Point | null>(null)
const pendingStart = ref<Point | null>(null)
const snapToGrid = ref(true); const drawCover = ref(false); const GRID_SIZE = 20
function snap(val: number): number { return snapToGrid.value ? Math.round(val / GRID_SIZE) * GRID_SIZE : val }
function openDrawer() { showUploadMenu.value = false; showDrawer.value = true; drawPoints.value = []; drawLines.value = []; nextTick(() => initCanvas()) }
function closeDrawer() { showDrawer.value = false }
function initCanvas() { const c = drawCanvas.value; if (!c) return; const p = c.parentElement!; c.width = p.clientWidth; c.height = Math.min(p.clientWidth * 0.75, window.innerHeight * 0.6); const ctx = c.getContext('2d')!; ctx.fillStyle = '#fff'; ctx.fillRect(0, 0, c.width, c.height); drawCtx.value = ctx; redraw() }

function redraw() {
  const ctx = drawCtx.value; const c = drawCanvas.value; if (!ctx || !c) return
  ctx.fillStyle = '#fff'; ctx.fillRect(0, 0, c.width, c.height)
  ctx.strokeStyle = '#e5e7eb'; ctx.lineWidth = 0.5
  for (let x = GRID_SIZE; x < c.width; x += GRID_SIZE) { ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, c.height); ctx.stroke() }
  for (let y = GRID_SIZE; y < c.height; y += GRID_SIZE) { ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(c.width, y); ctx.stroke() }
  if (snapToGrid.value) { ctx.fillStyle = '#d1d5db'; for (let x = 0; x <= c.width; x += GRID_SIZE) { for (let y = 0; y <= c.height; y += GRID_SIZE) { ctx.beginPath(); ctx.arc(x, y, 1.5, 0, Math.PI * 2); ctx.fill() } } }
  ctx.strokeStyle = '#ef4444'; ctx.lineWidth = 2.5; ctx.lineCap = 'round'; ctx.lineJoin = 'round'
  for (const line of drawLines.value) { const a = drawPoints.value[line.from]; const b = drawPoints.value[line.to]; if (!a || !b) continue; ctx.beginPath(); ctx.moveTo(a.x, a.y); ctx.lineTo(b.x, b.y); ctx.stroke() }
  if (pendingStart.value && lastDrawPos.value) { ctx.strokeStyle = '#3b82f6'; ctx.lineWidth = 2; ctx.setLineDash([6, 3]); ctx.beginPath(); ctx.moveTo(pendingStart.value.x, pendingStart.value.y); ctx.lineTo(lastDrawPos.value.x, lastDrawPos.value.y); ctx.stroke(); ctx.setLineDash([]) }
  for (let i = 0; i < drawPoints.value.length; i++) { const p = drawPoints.value[i]; ctx.fillStyle = i % 2 === 0 ? '#22c55e' : '#ef4444'; ctx.beginPath(); ctx.arc(p.x, p.y, 5, 0, Math.PI * 2); ctx.fill(); ctx.strokeStyle = '#fff'; ctx.lineWidth = 1.5; ctx.beginPath(); ctx.arc(p.x, p.y, 5, 0, Math.PI * 2); ctx.stroke() }
  if (pendingStart.value) { ctx.fillStyle = '#22c55e'; ctx.beginPath(); ctx.arc(pendingStart.value.x, pendingStart.value.y, 5, 0, Math.PI * 2); ctx.fill(); ctx.strokeStyle = '#fff'; ctx.lineWidth = 1.5; ctx.beginPath(); ctx.arc(pendingStart.value.x, pendingStart.value.y, 5, 0, Math.PI * 2); ctx.stroke() }
}

function onCanvasClick(e: MouseEvent) {
  const c = drawCanvas.value; if (!c) return; const r = c.getBoundingClientRect(); const x = snap(e.clientX - r.left); const y = snap(e.clientY - r.top)
  if (!pendingStart.value) { pendingStart.value = { x, y }; lastDrawPos.value = null }
  else { const si = drawPoints.value.length; drawPoints.value.push(pendingStart.value, { x, y }); drawLines.value.push({ from: si, to: si + 1 }); pendingStart.value = null; lastDrawPos.value = null }
  redraw()
}
function onCanvasMove(e: MouseEvent) { if (!pendingStart.value) return; const c = drawCanvas.value; if (!c) return; const r = c.getBoundingClientRect(); lastDrawPos.value = { x: snap(e.clientX - r.left), y: snap(e.clientY - r.top) }; redraw() }
function undoLastPoint() { if (pendingStart.value) { pendingStart.value = null; lastDrawPos.value = null; redraw(); return } if (drawPoints.value.length === 0) return; drawPoints.value.pop(); drawPoints.value.pop(); if (drawLines.value.length > 0) drawLines.value.pop(); lastDrawPos.value = null; redraw() }
function clearCanvas() { drawPoints.value = []; drawLines.value = []; pendingStart.value = null; lastDrawPos.value = null; redraw() }

async function saveDrawing() {
  const c = drawCanvas.value; if (!c) return; const ctx = c.getContext('2d')!
  ctx.fillStyle = '#fff'; ctx.fillRect(0, 0, c.width, c.height)
  ctx.strokeStyle = '#374151'; ctx.lineWidth = 3; ctx.lineCap = 'round'; ctx.lineJoin = 'round'
  for (const line of drawLines.value) { const a = drawPoints.value[line.from]; const b = drawPoints.value[line.to]; if (!a || !b) continue; ctx.beginPath(); ctx.moveTo(a.x, a.y); ctx.lineTo(b.x, b.y); ctx.stroke() }
  const blob = await new Promise<Blob | null>(resolve => c.toBlob(resolve, 'image/png')); redraw()
  if (blob) { closeDrawer(); await doUpload(new File([blob], 'floor-plan.png', { type: 'image/png' }), `${floorPlans.value.length + 1}F`, drawCover.value); drawCover.value = false }
}

const onWindowClick = () => { showUploadMenu.value = false }
window.addEventListener('click', onWindowClick)
onUnmounted(() => window.removeEventListener('click', onWindowClick))
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl md:text-2xl font-bold dark:text-gray-200">{{ t('floorPlan.heading') }}</h2>
      <div class="relative" @click.stop>
        <button class="btn-primary text-sm" :class="{ 'opacity-50': uploading }" @click="showUploadMenu = !showUploadMenu">
          {{ uploading ? '...' : '+ ' + t('floorPlan.upload') }} ▾
        </button>
        <div v-if="showUploadMenu" class="absolute right-0 top-full mt-1 w-44 bg-white dark:bg-gray-800 rounded-lg shadow-lg border dark:border-gray-700 z-30 py-1">
          <label class="flex items-center gap-2 px-4 py-2.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer">
            🖼️ {{ t('floorPlan.uploadImage') }}
            <input type="file" accept="image/jpeg,image/png,image/webp" class="hidden" @change="uploadFile" />
          </label>
          <button class="flex items-center gap-2 px-4 py-2.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 w-full text-left" @click="openDrawer">✏️ {{ t('floorPlan.drawPlan') }}</button>
        </div>
      </div>
    </div>

    <p v-if="error" class="text-danger text-sm mb-3">{{ error }}</p>

    <LoadingSpinner v-if="loading" />
    <template v-else>

    <!-- Empty -->
    <div v-if="floorPlans.length === 0" class="card text-center py-12 text-gray-400">
      <p class="mb-4">{{ t('floorPlan.uploadHint') }}</p>
    </div>

    <!-- Card grid -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
      <div v-for="plan in floorPlans" :key="plan.id"
        class="card cursor-pointer hover:shadow-lg transition-shadow group overflow-hidden"
        @click="openEdit(plan)"
      >
        <div class="-mx-4 -mt-4 mb-3 aspect-video bg-gray-200 dark:bg-gray-700 overflow-hidden relative">
          <img v-if="plan.image_url" :src="plan.image_url" class="w-full h-full object-cover" />
          <div v-else class="w-full h-full flex items-center justify-center text-3xl opacity-30">{{ plan.label[0] }}</div>
          <span v-if="plan.is_cover" class="absolute top-2 left-2 px-2 py-0.5 rounded text-xs bg-yellow-400 text-yellow-900 font-medium">⭐ 封面</span>
        </div>
        <div class="flex items-center justify-between">
          <h3 class="font-medium dark:text-gray-200">{{ plan.label }}</h3>
          <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <button v-if="!plan.is_cover" class="text-xs px-1.5 py-0.5 rounded hover:bg-yellow-100 dark:hover:bg-yellow-900" @click.stop="setAsCover(plan)" title="设为封面">⭐</button>
            <button class="text-xs px-1.5 py-0.5 rounded text-danger hover:bg-red-50 dark:hover:bg-red-900/30" @click.stop="deletePlan(plan)" title="删除">🗑</button>
          </div>
        </div>
        <p class="text-xs text-gray-400 mt-1">{{ plan.locations?.length || 0 }} 个地点</p>
      </div>
    </div>

    <!-- ─── Edit Modal ────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="editPlan" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="closeEdit">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[95vw] max-w-5xl max-h-[95vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <div class="flex items-center gap-2">
              <h3 class="font-bold dark:text-gray-200">{{ editPlan.label }}</h3>
              <span v-if="editPlan.is_cover" class="text-yellow-500 text-sm">⭐ 封面</span>
            </div>
            <div class="flex items-center gap-2">
              <button v-if="!marking" class="btn-primary text-xs" @click="marking = true">{{ t('floorPlan.markLocation') }}</button>
              <button v-else class="text-xs px-2 py-1 rounded bg-gray-200 dark:bg-gray-700 dark:text-gray-300" @click="marking = false">取消标记</button>
              <button class="text-xs px-2 py-1 rounded" :class="showLocations ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' : 'bg-gray-100 dark:bg-gray-700 dark:text-gray-300'" @click="showLocations = !showLocations">{{ showLocations ? '📍 显示' : '📍 隐藏' }}</button>
              <button v-if="!editPlan.is_cover" class="text-xs px-2 py-1 rounded bg-yellow-100 dark:bg-yellow-900 text-yellow-700 dark:text-yellow-300" @click="setAsCover(editPlan)">⭐ 设为封面</button>
              <button class="text-xs px-2 py-1 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-lg leading-none" @click="closeEdit">✕</button>
            </div>
          </div>
          <div class="flex-1 overflow-auto p-4 flex flex-col lg:flex-row gap-4">
            <div class="flex-1 min-h-[300px] bg-gray-200 dark:bg-gray-700 rounded-lg overflow-auto relative select-none">
              <img ref="editImgRef" :src="editPlan.image_url" class="w-full h-auto block" :class="{ 'cursor-crosshair': marking }" @load="onEditImgLoad" draggable="false" />
              <svg v-if="imgNatural.w > 0 && showLocations" class="absolute inset-0 pointer-events-none" :width="displaySize.w" :height="displaySize.h" :viewBox="`0 0 ${displaySize.w} ${displaySize.h}`">
                <g v-for="loc in editPlan.locations || []" :key="loc.id" class="pointer-events-auto cursor-pointer" @click.stop @mouseenter="hoveredLoc = loc.id" @mouseleave="hoveredLoc = null">
                  <circle :cx="scaleX(loc.point.x)" :cy="scaleY(loc.point.y)" r="6" :fill="loc.color" :stroke="hoveredLoc === loc.id ? '#fff' : 'transparent'" stroke-width="2" class="transition-all" />
                  <text :x="scaleX(loc.point.x) + 10" :y="scaleY(loc.point.y) + 4" class="text-xs font-medium pointer-events-none" :fill="loc.color">{{ loc.name }}</text>
                </g>
              </svg>
              <!-- Click overlay for marking -->
              <div v-if="marking" class="absolute inset-0 z-10" @click="onSvgClick" />
            </div>
            <div class="w-full lg:w-56 flex-shrink-0 space-y-3">
              <div v-if="showLocInput" class="card space-y-2">
                <input v-model="newLocName" class="input" :placeholder="t('floorPlan.locNamePlaceholder')" @keyup.enter="saveLocation" />
                <div class="flex gap-1 flex-wrap">
                  <button v-for="c in PRESET_COLORS" :key="c" class="w-6 h-6 rounded-full border-2 transition-transform" :class="newLocColor === c ? 'border-gray-800 dark:border-white scale-125' : 'border-transparent'" :style="{ backgroundColor: c }" @click="newLocColor = c" />
                </div>
                <div class="flex gap-1">
                  <button class="btn-primary text-xs flex-1" @click="saveLocation">{{ t('floorPlan.saveRoom') }}</button>
                  <button class="text-xs px-2 py-1 rounded text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700" @click="cancelLoc">{{ t('floorPlan.cancel') }}</button>
                </div>
              </div>
              <div class="card">
                <h4 class="text-sm font-medium mb-2 dark:text-gray-200">地点列表</h4>
                <div v-if="!editPlan.locations?.length" class="text-xs text-gray-400 py-2">暂无地点</div>
                <div v-for="loc in editPlan.locations" :key="loc.id" class="flex items-center justify-between py-1.5 text-sm" @mouseenter="hoveredLoc = loc.id" @mouseleave="hoveredLoc = null">
                  <div class="flex items-center gap-2"><div class="w-2.5 h-2.5 rounded-full flex-shrink-0" :style="{ backgroundColor: loc.color }" /><span class="dark:text-gray-300 truncate">{{ loc.name }}</span></div>
                  <div class="flex gap-0.5 flex-shrink-0 ml-1">
                    <button class="text-xs px-1 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 dark:text-gray-400" @click="startEditLoc(loc)" title="编辑">✎</button>
                    <button class="text-xs text-danger hover:underline" @click="deleteLocation(loc.id)">删除</button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Drawing Modal ──────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDrawer" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="closeDrawer">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-[90vw] max-w-4xl max-h-[90vh] flex flex-col">
          <div class="flex items-center justify-between px-4 py-3 border-b dark:border-gray-700">
            <h3 class="font-bold dark:text-gray-200">{{ t('floorPlan.drawPlan') }}</h3>
            <div class="flex items-center gap-2">
              <button class="text-xs px-2 py-1 rounded transition-colors" :class="snapToGrid ? 'bg-primary text-white' : 'bg-gray-100 dark:bg-gray-700 dark:text-gray-300'" @click="snapToGrid = !snapToGrid">⊞ 吸附</button>
              <button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 dark:text-gray-300 hover:opacity-80" @click="undoLastPoint">↩ 撤销</button>
              <button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 dark:text-gray-300 hover:opacity-80" @click="clearCanvas">🗑 清除</button>
              <button class="text-xs px-2 py-1 rounded bg-primary text-white hover:opacity-90" :disabled="drawPoints.length < 2" @click="saveDrawing">保存</button>
              <button class="text-xs px-2 py-1 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-300" @click="closeDrawer">✕</button>
            </div>
          </div>
          <div class="flex-1 overflow-auto p-4 flex items-center justify-center bg-gray-100 dark:bg-gray-900">
            <canvas ref="drawCanvas" class="border border-gray-300 dark:border-gray-600 rounded shadow cursor-crosshair bg-white" @click="onCanvasClick" @mousemove="onCanvasMove" @contextmenu.prevent="undoLastPoint" />
          </div>
          <div class="px-4 py-2 border-t dark:border-gray-700 text-xs text-gray-400 flex gap-4 flex-wrap items-center">
            <span>🟢 点击设置起点</span><span>🔴 再次点击设置终点</span><span>🖱 右键撤销</span>
            <span v-if="snapToGrid" class="text-primary">⊞ 吸附网格 {{ GRID_SIZE }}px</span>
            <label class="flex items-center gap-1 ml-auto cursor-pointer text-gray-500 dark:text-gray-400 hover:text-primary">
              <input type="checkbox" v-model="drawCover" class="w-3.5 h-3.5" /> 设为封面
            </label>
          </div>
        </div>
      </div>
    </Teleport>
    </template>
  </div>
</template>
