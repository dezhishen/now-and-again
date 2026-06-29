<script setup lang="ts">
/**
 * Shared base fields for all task forms.
 * Used internally by SimpleTaskForm, InspectionTaskForm, and future task kind forms.
 *
 * v-model shape: { name, schedule_type, schedule_data, group_id, location_id, enabled }
 */
import { useI18n } from '@/i18n'
import type { I18nKey } from '@/i18n'
import type { FamilyGroup } from '@/types'

export interface TaskBaseFields {
  name: string
  schedule_type: string
  schedule_data: {
    time?: string
    date?: string
    days?: number[]
  }
  group_id?: string
  location_id?: string
  enabled: boolean
}

const { t } = useI18n()

const model = defineModel<TaskBaseFields>({ required: true })

defineProps<{
  groups: FamilyGroup[]
  locations: { id: string; name: string }[]
}>()

const SCHEDULE_TYPES: { value: string; labelKey: I18nKey }[] = [
  { value: 'once', labelKey: 'schedule.once' },
  { value: 'daily', labelKey: 'schedule.daily' },
  { value: 'weekly', labelKey: 'schedule.weekly' },
  { value: 'monthly', labelKey: 'schedule.monthly' },
  { value: 'interval', labelKey: 'schedule.interval' },
]

const WEEKDAYS = ['一', '二', '三', '四', '五', '六', '日']
const MONTH_DAYS = Array.from({ length: 31 }, (_, i) => i + 1)

function ensureSD(task: TaskBaseFields) {
  if (!task.schedule_data || typeof task.schedule_data !== 'object') {
    task.schedule_data = {}
  }
}

function toggleDay(task: TaskBaseFields, d: number) {
  ensureSD(task)
  if (!Array.isArray(task.schedule_data.days)) {
    task.schedule_data.days = []
  }
  const idx = task.schedule_data.days.indexOf(d)
  if (idx >= 0) task.schedule_data.days.splice(idx, 1)
  else task.schedule_data.days.push(d)
}
</script>

<template>
  <div class="space-y-3">
    <!-- Name -->
    <div>
      <label class="text-xs text-gray-400 block mb-1">任务名称</label>
      <input v-model="model.name" class="input" placeholder="输入任务名称" />
    </div>

    <!-- Schedule type -->
    <div>
      <label class="text-xs text-gray-400 block mb-1">调度方式</label>
      <select v-model="model.schedule_type" class="input">
        <option v-for="s in SCHEDULE_TYPES" :key="s.value" :value="s.value">{{ t(s.labelKey) }}</option>
      </select>
    </div>

    <!-- Time -->
    <div>
      <label class="text-xs text-gray-400 block mb-1">触发时间</label>
      <input v-model="model.schedule_data.time" type="time" class="input" />
    </div>

    <!-- Once: date -->
    <div v-if="model.schedule_type === 'once'">
      <label class="text-xs text-gray-400 block mb-1">执行日期</label>
      <input v-model="model.schedule_data.date" type="date" class="input" />
      <p class="text-xs text-gray-400 mt-1">选择过去的日期会立即生成待办</p>
    </div>

    <!-- Weekly / Monthly / Interval day picker -->
    <div v-if="model.schedule_type !== 'daily' && model.schedule_type !== 'once'">
      <label class="text-xs text-gray-400 block mb-1">
        {{ model.schedule_type === 'weekly' ? '选择星期' : model.schedule_type === 'monthly' ? '选择日期' : '间隔天数' }}
      </label>
      <div class="flex flex-wrap gap-1">
        <template v-if="model.schedule_type === 'weekly'">
          <button v-for="(name, i) in WEEKDAYS" :key="i"
            class="text-xs px-2 py-1 rounded border transition-colors"
            :class="(model.schedule_data.days || []).includes(i + 1) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
            @click="toggleDay(model, i + 1)">{{ name }}</button>
        </template>
        <template v-else-if="model.schedule_type === 'monthly'">
          <button v-for="d in MONTH_DAYS" :key="d"
            class="text-xs w-7 h-7 rounded border transition-colors flex items-center justify-center"
            :class="(model.schedule_data.days || []).includes(d) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
            @click="toggleDay(model, d)">{{ d }}</button>
        </template>
        <template v-else>
          <input type="number" :model-value="model.schedule_data.days?.[0] || 1"
            @input="ensureSD(model); model.schedule_data.days = [parseInt(($event.target as HTMLInputElement).value) || 1]"
            class="input w-20" placeholder="天数" min="1" />
        </template>
      </div>
    </div>

    <!-- Location -->
    <div>
      <label class="text-xs text-gray-400 block mb-1">关联地点（可选）</label>
      <div class="flex gap-2 items-center">
        <select v-model="model.location_id" class="input flex-1">
          <option value="">不关联</option>
          <option v-for="loc in $props.locations" :key="loc.id" :value="loc.id">{{ loc.name }}</option>
        </select>
        <button v-if="model.location_id" class="text-xs text-gray-400 hover:text-danger flex-shrink-0" @click="model.location_id = ''">清除</button>
      </div>
    </div>

    <!-- Group -->
    <div>
      <label class="text-xs text-gray-400 block mb-1">分配给小组（可选）</label>
      <div class="flex gap-2 items-center">
        <select v-model="model.group_id" class="input flex-1">
          <option value="">全部成员</option>
          <option v-for="g in $props.groups" :key="g.id" :value="g.id">{{ g.name }}</option>
        </select>
        <button v-if="model.group_id" class="text-xs text-gray-400 hover:text-danger flex-shrink-0" @click="model.group_id = ''">清除</button>
      </div>
    </div>
  </div>
</template>
