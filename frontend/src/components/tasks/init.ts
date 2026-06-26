import { registerTaskKind } from '@/composables/useTaskKinds'
import SimpleTaskBody from '@/components/tasks/SimpleTaskBody.vue'
import InspectionTaskBody from '@/components/tasks/InspectionTaskBody.vue'

export function initTaskKinds() {
  registerTaskKind('simple', {
    card: SimpleTaskBody,
    label: '任务',
    createLabel: '创建任务',
  })

  registerTaskKind('inspection', {
    card: InspectionTaskBody,
    label: '巡检',
    createLabel: '创建巡检',
    defaultCheckItems: [
      {
        name: '检查项1',
        branches: [
          { name: '正常', create_todo: false },
          { name: '异常', create_todo: true, todo_name: '修复{name}' },
        ],
      },
    ],
  })
}
