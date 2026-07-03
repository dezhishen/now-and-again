<script setup lang="ts">
import { useI18n } from '@/i18n'
import { getScheduleTypes } from './registry'

const { t } = useI18n()

const SCHEDULE_TYPES = getScheduleTypes()

const scheduleType = defineModel<string>('scheduleType', { default: 'daily' })
const scheduleTime = defineModel<string>('scheduleTime', { default: '09:00' })
const scheduleDate = defineModel<string>('scheduleDate', { default: '' })
const scheduleDays = defineModel<number[]>('scheduleDays', { default: () => [] })
const scheduleYearDay = defineModel<number>('scheduleYearDay', { default: 1 })

const WEEKDAYS = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']

function toggleDay(d: number) {
  const idx = scheduleDays.value.indexOf(d)
  if (idx >= 0) scheduleDays.value.splice(idx, 1)
  else scheduleDays.value.push(d)
}
</script>

<template>
  <div class="space-y-3">
    <div>
      <label class="text-xs text-gray-400 block mb-1">调度方式</label>
      <select v-model="scheduleType" class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm">
        <option v-for="s in SCHEDULE_TYPES" :key="s.value" :value="s.value">{{ t(s.labelKey) }}</option>
      </select>
    </div>
    <div>
      <label class="text-xs text-gray-400 block mb-1">触发时间</label>
      <input v-model="scheduleTime" type="time" class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
    </div>
    <div v-if="scheduleType === 'once'">
      <label class="text-xs text-gray-400 block mb-1">执行日期</label>
      <input v-model="scheduleDate" type="date" class="w-full rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" />
    </div>
    <div v-if="scheduleType !== 'daily' && scheduleType !== 'once'">
      <label class="text-xs text-gray-400 block mb-1">
        {{ scheduleType === 'weekly' ? '选择星期' : scheduleType === 'monthly' ? '选择日期' : scheduleType === 'yearly' ? '选择月份' : '间隔天数' }}
      </label>
      <div class="flex flex-wrap gap-1">
        <template v-if="scheduleType === 'weekly'">
          <button v-for="(name, i) in WEEKDAYS" :key="i"
            class="text-xs px-2 py-1 rounded border transition-colors"
            :class="scheduleDays.includes(i+1) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
            @click="toggleDay(i+1)">{{ name }}</button>
        </template>
        <template v-else-if="scheduleType === 'monthly'">
          <button v-for="d in 31" :key="d"
            class="text-xs w-7 h-7 rounded border transition-colors flex items-center justify-center"
            :class="scheduleDays.includes(d) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
            @click="toggleDay(d)">{{ d }}</button>
        </template>
        <template v-else-if="scheduleType === 'yearly'">
          <div class="flex flex-wrap gap-1 mb-2">
            <button v-for="m in 12" :key="m"
              class="text-xs px-2 py-1 rounded border transition-colors"
              :class="scheduleDays.includes(m) ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
              @click="toggleDay(m)">{{ m }}月</button>
          </div>
          <label class="text-xs text-gray-400 block mb-1">选择日期</label>
          <div class="flex flex-wrap gap-1">
            <button v-for="d in 31" :key="d"
              class="text-xs w-7 h-7 rounded border transition-colors flex items-center justify-center"
              :class="scheduleYearDay === d ? 'bg-primary text-white border-primary' : 'border-gray-200 dark:border-gray-600 dark:text-gray-400'"
              @click="scheduleYearDay = d">{{ d }}</button>
          </div>
        </template>
        <template v-else>
          <input type="number" :model-value="scheduleDays[0] || 1" @input="scheduleDays = [$event.target ? Number(($event.target as HTMLInputElement).value) : 1]" class="w-20 rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-3 py-2 text-sm" placeholder="天数" min="1" />
        </template>
      </div>
    </div>
  </div>
</template>
