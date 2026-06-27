<script setup lang="ts">
import type { CheckItem, FamilyGroup } from '@/types'

const checkItems = defineModel<CheckItem[]>({ required: true })
defineProps<{ groups: FamilyGroup[] }>()

function addBranch(item: CheckItem) {
  item.branches.push({ name: '', create_todo: false })
}

function addItem() {
  checkItems.value.push({
    name: '',
    branches: [
      { name: '正常', create_todo: false },
      { name: '异常', create_todo: true, todo_name: '修复{name}' },
    ],
  })
}
</script>

<template>
  <div class="space-y-3 border-l-2 border-purple-400 pl-3">
    <div class="flex items-center justify-between">
      <p class="text-xs text-purple-600 dark:text-purple-400 font-medium">🔍 待检查项</p>
      <button class="text-xs text-primary hover:underline" @click="addItem">+ 添加检查项</button>
    </div>
    <div v-for="(item, i) in checkItems" :key="i" class="space-y-1 pb-2 border-b border-gray-100 dark:border-gray-700 last:border-0">
      <div class="flex gap-2 items-center">
        <input v-model="item.name" class="input flex-1 text-sm" placeholder="检查项名称，如 水槽" />
        <button class="text-xs text-danger hover:underline flex-shrink-0" @click="checkItems.splice(i, 1)">删除</button>
      </div>
      <!-- Branches within this item -->
      <div class="ml-2 space-y-1">
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-500">分支:</span>
          <button class="text-xs text-primary hover:underline" @click="addBranch(item)">+ 添加</button>
        </div>
        <div v-for="(b, j) in item.branches" :key="j" class="flex flex-wrap items-center gap-1 ml-2">
          <input v-model="b.name" class="input text-xs w-24" placeholder="分支名" />
          <label class="flex items-center gap-0.5 text-xs cursor-pointer">
            <input type="checkbox" v-model="b.create_todo" class="accent-purple-500" />
            <span class="text-gray-400">创建任务</span>
          </label>
          <template v-if="b.create_todo">
            <input v-model="b.todo_name" class="input text-xs flex-1 min-w-[120px]" placeholder="任务名" />
            <select v-model="b.group_id" class="input text-xs w-28">
              <option value="">小组</option>
              <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
            </select>
          </template>
          <button class="text-xs text-danger hover:underline" @click="item.branches.splice(j, 1)">×</button>
        </div>
      </div>
    </div>
  </div>
</template>
