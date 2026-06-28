<script setup lang="ts">
/**
 * Receives { task: Task, extra: any } — never inspects task.extra.
 * Plugin-internal data lives in the top-level `extra`.
 */
const model = defineModel<{ task: any; extra: any } | null>('task', { required: true })
/** branch_id → { item_id, item_name, branch_name } */
const selections = defineModel<Record<string, { item_id: string; item_name: string; branch_name: string }>>('selections', { default: () => ({}) })

function selectBranch(branchId: string, itemId: string, itemName: string, branchName: string) {
  selections.value[branchId] = { item_id: itemId, item_name: itemName, branch_name: branchName }
}

function isSelected(branchId: string) {
  return branchId in selections.value
}
</script>

<template>
  <div class="flex-1 overflow-auto p-4 space-y-4">
    <div v-for="item in (model?.extra?.check_items || []) as any[]" :key="item.id || item.name" class="space-y-1">
      <p class="text-sm font-medium text-gray-600 dark:text-gray-300">{{ item.name }}</p>
      <div class="flex flex-wrap gap-1">
        <button
          v-for="b in item.branches" :key="b.id || b.name"
          class="text-xs px-2 py-1 rounded border transition-colors"
          :class="isSelected(b.id || b.name)
            ? (b.create_todo ? 'bg-red-500 text-white border-red-500' : 'bg-green-500 text-white border-green-500')
            : 'border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-primary'"
          @click="selectBranch(b.id || b.name, item.id || item.name, item.name, b.name)"
        >{{ b.name }}</button>
      </div>
    </div>
  </div>
</template>
