<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'
import type { Task, TaskType } from '@/types'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

const tasks = ref<Task[]>([])
const taskTypes = ref<TaskType[]>([])
const showCreate = ref(false)
const newTask = ref({ title: '', task_type_id: '', priority: 'medium', description: '' })
const error = ref('')

onMounted(async () => {
  await loadTasks()
  try { taskTypes.value = await api.get<TaskType[]>('/task-types') } catch { /* */ }
})

async function loadTasks() {
  try {
    const data = await api.get<{ data: Task[] }>('/families/' + familyId + '/tasks?page_size=50')
    tasks.value = (data as any).data || data as any || []
  } catch { tasks.value = [] }
}

async function createTask() {
  error.value = ''
  try {
    await api.post('/families/' + familyId + '/tasks', {
      ...newTask.value,
      family_id: familyId,
    })
    newTask.value = { title: '', task_type_id: '', priority: 'medium', description: '' }
    showCreate.value = false
    await loadTasks()
  } catch (e: any) { error.value = e.message }
}

async function completeTask(taskId: string) {
  await api.patch('/tasks/' + taskId, { status: 'done' })
  await loadTasks()
}

async function startTask(taskId: string) {
  await api.patch('/tasks/' + taskId, { status: 'in_progress' })
  await loadTasks()
}

function statusColor(s: string) {
  return s === 'todo' ? 'text-yellow-600' : s === 'in_progress' ? 'text-blue-600' : 'text-green-600'
}
function statusLabel(s: string) {
  return s === 'todo' ? '待办' : s === 'in_progress' ? '进行中' : '已完成'
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-bold">{{ t('tasks.heading') }}</h2>
      <button class="btn-primary text-sm" @click="showCreate = !showCreate">{{ showCreate ? '取消' : '+ 新建任务' }}</button>
    </div>

    <div v-if="showCreate" class="card mb-4 flex flex-col gap-2">
      <input v-model="newTask.title" class="input" placeholder="任务标题" @keyup.enter="createTask" />
      <div class="flex gap-2">
        <select v-model="newTask.task_type_id" class="input flex-1">
          <option value="">选择类型</option>
          <option v-for="tt in taskTypes" :key="tt.id" :value="tt.id">{{ tt.icon }} {{ tt.name }}</option>
        </select>
        <select v-model="newTask.priority" class="input w-28">
          <option value="low">低</option>
          <option value="medium">中</option>
          <option value="high">高</option>
        </select>
        <button class="btn-primary" @click="createTask">创建</button>
      </div>
      <p v-if="error" class="text-danger text-sm">{{ error }}</p>
    </div>

    <div v-if="tasks.length === 0" class="text-center text-muted py-8">暂无任务</div>

    <div v-for="task in tasks" :key="task.id" class="card mb-2 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-lg">{{ task.task_type?.icon || '📋' }}</span>
        <div>
          <div class="font-medium" :class="{ 'line-through text-muted': task.status === 'done' }">{{ task.title }}</div>
          <div class="text-xs text-muted">{{ task.task_type?.name }} · <span :class="statusColor(task.status)">{{ statusLabel(task.status) }}</span></div>
        </div>
      </div>
      <div class="flex gap-1">
        <button v-if="task.status === 'todo'" class="text-sm px-3 py-1 rounded bg-blue-100 text-blue-700 hover:bg-blue-200" @click="startTask(task.id)">开始</button>
        <button v-if="task.status !== 'done'" class="text-sm px-3 py-1 rounded bg-green-100 text-green-700 hover:bg-green-200" @click="completeTask(task.id)">完成</button>
      </div>
    </div>
  </div>
</template>
