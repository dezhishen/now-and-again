<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { User } from '@/types'

const users = ref<User[]>([])
onMounted(async () => { users.value = await api.get<User[]>('/admin/users') })
</script>

<template>
  <div>
    <h2 class="text-xl md:text-2xl font-bold mb-4 dark:text-gray-200">用户管理</h2>
    <div class="card overflow-x-auto">
      <table class="w-full text-sm min-w-[500px]">
        <thead>
          <tr class="border-b dark:border-gray-700 text-left text-gray-500 dark:text-gray-400">
            <th class="py-2 px-3">显示名称</th><th class="py-2 px-3">邮箱</th><th class="py-2 px-3">角色</th><th class="py-2 px-3 hidden sm:table-cell">注册时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
            <td class="py-2 px-3 font-medium dark:text-gray-200">{{ u.display_name }}</td>
            <td class="py-2 px-3 dark:text-gray-300">{{ u.email }}</td>
            <td class="py-2 px-3"><span v-if="u.roles.includes('admin')" class="text-primary font-medium">管理员</span><span v-else class="text-gray-400">成员</span></td>
            <td class="py-2 px-3 text-gray-400 hidden sm:table-cell">{{ u.created_at?.split('T')[0] }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
