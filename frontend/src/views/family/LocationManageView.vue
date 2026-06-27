<script setup lang="ts">
import { ref, onMounted, inject, watch, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { getLocationKinds, getLocationKindIcon } from '@/composables/useLocationKinds'
import { initLocationKinds } from '@/components/locations/init'
import type { Location, FloorPlan } from '@/types'

initLocationKinds()

const { t } = useI18n()
const toast = useToast()
const route = useRoute()
const familyId = route.params.familyId as string

const refreshKey = inject<Ref<string>>('refreshKey', ref(''))
watch(refreshKey, (newVal) => { if (newVal === 'locations') { loadLocations(); loadPlans() } })

const locations = ref<Location[]>([])
const floorPlans = ref<FloorPlan[]>([])
const loading = ref(true)

// ─── Edit modal ──────────────────────────────────────────────────
const editing = ref(false)
const editLoc = ref<Location | null>(null)
const formName = ref('')
const formColor = ref('#3b82f6')
const formPlanId = ref('')
const formKind = ref('indoor')
const locationKinds = getLocationKinds()

const PRESET_COLORS = ['#3b82f6','#ef4444','#22c55e','#f59e0b','#8b5cf6','#ec4899','#06b6d4','#78716c']

onMounted(async () => {
  loading.value = true
  await Promise.all([loadLocations(), loadPlans()])
  loading.value = false
})

async function loadLocations() {
  try {
    locations.value = await api.get<Location[]>('/families/' + familyId + '/locations')
  } catch { locations.value = [] }
}

async function loadPlans() {
  try {
    floorPlans.value = await api.get<FloorPlan[]>('/families/' + familyId + '/floor-plans')
  } catch { floorPlans.value = [] }
}

function openCreate() {
  editLoc.value = null
  formName.value = ''
  formColor.value = '#3b82f6'
  formPlanId.value = ''
  formKind.value = 'indoor'
  editing.value = true
}

function openEdit(loc: Location) {
  editLoc.value = loc
  formName.value = loc.name
  formColor.value = loc.color
  formPlanId.value = loc.floor_plan_id || ''
  formKind.value = loc.kind || 'indoor'
  editing.value = true
}

async function saveLocation() {
  if (!formName.value.trim()) return
  const body: any = { name: formName.value.trim(), kind: formKind.value, color: formColor.value }
  if (formPlanId.value) {
    body.floor_plan_id = formPlanId.value
  }

  try {
    if (editLoc.value) {
      const updated = await api.put<Location>('/locations/' + editLoc.value.id, {
        name: formName.value.trim(),
        kind: formKind.value,
        color: formColor.value,
        floor_plan_id: formPlanId.value || null,
      })
      const idx = locations.value.findIndex(l => l.id === editLoc.value!.id)
      if (idx >= 0) locations.value[idx] = updated
      toast.success(t('locations.updated'))
    } else {
      const created = await api.post<Location>('/families/' + familyId + '/locations', body)
      locations.value.push(created)
      toast.success(t('locations.created'))
    }
    editing.value = false
  } catch (e: any) { toast.error(e.message) }
}

async function deleteLocation(loc: Location) {
  if (!await useConfirm(t('locations.deleteConfirm').replace('{name}', loc.name))) return
  try {
    await api.delete('/locations/' + loc.id)
    locations.value = locations.value.filter(l => l.id !== loc.id)
    toast.success(t('locations.deleted'))
  } catch (e: any) { toast.error(e.message) }
}

async function unlinkPlan(loc: Location) {
  if (!await useConfirm(t('locations.unlinkConfirm').replace('{name}', loc.name))) return
  api.put<Location>('/locations/' + loc.id, { floor_plan_id: '' }).then(updated => {
    const idx = locations.value.findIndex(l => l.id === loc.id)
    if (idx >= 0) locations.value[idx] = updated
    toast.success(t('locations.unlinked'))
  }).catch((e: any) => toast.error(e.message))
}

function getPlanLabel(planId: string) {
  return floorPlans.value.find(p => p.id === planId)?.label || ''
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <button class="btn-primary" @click="openCreate">+ {{ t('locations.add') }}</button>
    </div>

    <!-- List -->
    <div v-if="locations.length === 0" class="text-center text-gray-400 py-8">
      {{ t('locations.empty') }}
    </div>
    <div v-else class="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-3">
      <div v-for="loc in locations" :key="loc.id"
        class="card flex items-center justify-between gap-3 hover:shadow-md transition-shadow cursor-pointer"
        @click="openEdit(loc)"
      >
        <div class="flex items-center gap-3 min-w-0">
          <span class="w-5 h-5 rounded-full flex-shrink-0" :style="{ background: loc.color }"></span>
          <div class="min-w-0">
            <p class="font-medium dark:text-gray-200 text-sm truncate">
              {{ getLocationKindIcon(loc.kind) }} {{ loc.name }}
            </p>
            <p v-if="loc.floor_plan_id" class="text-xs text-gray-400">
              🏠 {{ getPlanLabel(loc.floor_plan_id) }}
            </p>
            <p v-else class="text-xs text-gray-300 italic">{{ t('locations.notOnPlan') }}</p>
          </div>
        </div>
        <div class="flex gap-1 flex-shrink-0">
          <button v-if="loc.floor_plan_id" class="text-xs text-gray-400 hover:text-red-500 px-1" @click.stop="unlinkPlan(loc)" title="取消标记">✕</button>
          <button class="text-xs text-gray-400 hover:text-red-500 px-1" @click.stop="deleteLocation(loc)" title="删除">🗑</button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <Teleport to="body">
      <div v-if="editing" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @mousedown.self="editing = false">
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto">
          <h3 class="text-lg font-semibold mb-4 dark:text-gray-200">
            {{ editLoc ? t('locations.editLocation') : t('locations.addLocation') }}
          </h3>

          <!-- Name -->
          <label class="block text-sm text-gray-500 dark:text-gray-400 mb-1">{{ t('locations.name') }}</label>
          <input v-model="formName" class="input mb-4" :placeholder="t('locations.namePlaceholder')" />

          <!-- Color -->
          <label class="block text-sm text-gray-500 dark:text-gray-400 mb-1">{{ t('locations.color') }}</label>
          <div class="flex gap-2 mb-4 flex-wrap">
            <button v-for="c in PRESET_COLORS" :key="c" class="w-8 h-8 rounded-full border-2 transition-transform" :class="formColor === c ? 'border-gray-800 dark:border-white scale-110' : 'border-transparent'" :style="{ background: c }" @click="formColor = c" />
          </div>

          <!-- Kind -->
          <label class="block text-sm text-gray-500 dark:text-gray-400 mb-1">{{ t('locations.kind') }}</label>
          <div class="flex gap-2 mb-4 flex-wrap">
            <button v-for="k in locationKinds" :key="k.kind"
              class="px-3 py-1.5 text-sm rounded-lg border transition-colors"
              :class="formKind === k.kind ? 'border-primary bg-primary/10 text-primary' : 'border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-gray-400'"
              @click="formKind = k.kind"
            >{{ k.icon }} {{ k.label }}</button>
          </div>

          <!-- Floor Plan linking -->
          <label class="block text-sm text-gray-500 dark:text-gray-400 mb-1">{{ t('locations.linkPlan') }}</label>
          <select v-model="formPlanId" class="input mb-4">
            <option value="">{{ t('locations.noPlan') }}</option>
            <option v-for="fp in floorPlans" :key="fp.id" :value="fp.id">{{ fp.label }}</option>
          </select>

          <!-- Actions -->
          <div class="flex gap-2 justify-end">
            <button class="btn-secondary" @click="editing = false">{{ t('locations.cancel') }}</button>
            <button class="btn-primary" @click="saveLocation">{{ t('locations.save') }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
