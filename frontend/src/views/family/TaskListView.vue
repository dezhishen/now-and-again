<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { api } from '@/api/client'

const { t } = useI18n()
const route = useRoute()
const familyId = route.params.familyId as string

interface TaskItem { id: string; title: string; status: string; task_code: string }
interface SchedType { code: string; name: string; icon: string }

const tasks = ref<TaskItem[]>([])
const schedTypes = ref<SchedType[]>([])
const showCreate = ref(false)
const newTask = ref({ title: '', task_code: '', priority: 'medium' })
const error = ref('')

onMounted(async () => {
  await loadTasks()
  try { schedTypes.value = await api.get<SchedType[]>('/schedule-types') } catch { /* */ }
})

async function loadTasks() {
  try { tasks.value = await api.get<TaskItem[]>('/families/' + familyId + '/tasks?page_size=50') } catch { tasks.value = [] }
}

async function createTask() {
  error.value = ''
  if (!newTask.value.task_code) { error.value = '请选择类型'; return }
  try {
    await api.post('/families/' + familyId + '/tasks', { ...newTask.value, family_id: familyId })
    newTask.value = { title: '', task_code: '', priority: 'medium' }
    showCreate.value = false
    await loadTasks()
  } catch (e: any) { error.value = e.message }
}

async function updateStatus(taskId: string, status: string) {
  await api.patch('/tasks/' + taskId, { status })
  await loadTasks()
}

function stLabel(s: string) { return s === 'todo' ? '待办' : s === 'in_progress' ? '进行中' : '已完成' }
function stColor(s: string) { return s === 'todo' ? 'text-yellow-600' : s === 'in_progress' ? 'text-blue-600' : 'text-green-600' }
function typeName(code: string) { return schedTypes.value.find(t => t.code === code)?.name || code }
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl md:text-2xl font-bold dark:text-gray-200">{{ t('tasks.heading') }}</h2>
      <button class="btn-primary text-sm" @click="showCreate = !showCreate">{{ showCreate ? '取消' : '+ 新建' }}</button>
    </div>

    <div v-if="showCreate" class="card mb-4 flex flex-col gap-2">
      <input v-model="newTask.title" class="input" placeholder="任务标题" @keyup.enter="createTask" />
      <div class="flex flex-col sm:flex-row gap-2">
        <select v-model="newTask.task_code" class="input flex-1">
          <option value="">选择类型</option>
          <option v-for="st in schedTypes" :key="st.code" :value="st.code">{{ st.icon }} {{ st.name }}</option>
        </select>
        <select v-model="newTask.priority" class="input w-full sm:w-28">
          <option value="low">低</option><option value="medium">中</option><option value="high">高</option>
        </select>
        <button class="btn-primary" @click="createTask">创建</button>
      </div>
      <p v-if="error" class="text-danger text-sm">{{ error }}</p>
    </div>

    <div v-if="tasks.length === 0" class="text-center text-gray-400 py-8">暂无任务</div>

    <div v-for="task in tasks" :key="task.id" class="card mb-2 flex items-center justify-between">
      <div>
        <div class="font-medium dark:text-gray-200" :class="{ 'line-through text-gray-400': task.status === 'done' }">{{ task.title }}</div>
        <div class="text-xs text-gray-400">{{ typeName(task.task_code) }} · <span :class="stColor(task.status)">{{ stLabel(task.status) }}</span></div>
      </div>
      <div class="flex gap-1 flex-shrink-0">
        <button v-if="task.status === 'todo'" class="text-xs px-2 py-1 rounded bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300" @click="updateStatus(task.id, 'in_progress')">开始</button>
        <button v-if="task.status !== 'done'" class="text-xs px-2 py-1 rounded bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300" @click="updateStatus(task.id, 'done')">完成</button>
      </div>
    </div>
  </div>
</template>
