<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api/client'
import type { User } from '@/types'

const users = ref<User[]>([])

onMounted(async () => {
  users.value = await api.get<User[]>('/admin/users')
})
</script>

<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">用户管理</h2>
    <div class="card overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b text-left text-muted">
            <th class="py-2 px-3">用户名</th>
            <th class="py-2 px-3">显示名称</th>
            <th class="py-2 px-3">邮箱</th>
            <th class="py-2 px-3">角色</th>
            <th class="py-2 px-3">注册时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id" class="border-b hover:bg-gray-50">
            <td class="py-2 px-3 font-medium">{{ u.username }}</td>
            <td class="py-2 px-3">{{ u.display_name }}</td>
            <td class="py-2 px-3">{{ u.email }}</td>
            <td class="py-2 px-3">
              <span v-if="u.is_admin" class="text-primary font-medium">管理员</span>
              <span v-else class="text-muted">成员</span>
            </td>
            <td class="py-2 px-3 text-muted">{{ u.created_at?.split('T')[0] }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
