import { registerTaskKind } from '@/composables/useTaskKinds'
import SimpleTaskBody from '@/components/tasks/kinds/simple/SimpleTaskBody.vue'
import SimpleTodoActions from '@/components/tasks/kinds/simple/SimpleTodoActions.vue'
import InspectionTaskBody from '@/components/tasks/kinds/inspection/InspectionTaskBody.vue'
import InspectionInspect from '@/components/tasks/kinds/inspection/InspectionInspect.vue'
import InspectionTodoActions from '@/components/tasks/kinds/inspection/InspectionTodoActions.vue'
import InspectionTodoInfo from '@/components/tasks/kinds/inspection/InspectionTodoInfo.vue'
import TaskFormCheckItems from '@/components/tasks/TaskFormCheckItems.vue'

export function initTaskKinds() {
  registerTaskKind('simple', {
    card: SimpleTaskBody,
    todoActions: SimpleTodoActions,
    label: '任务',
    createLabel: '创建任务',
  })

  registerTaskKind('inspection', {
    card: InspectionTaskBody,
    inspectComponent: InspectionInspect,
    todoActions: InspectionTodoActions,
    todoInfo: InspectionTodoInfo,
    formComponent: TaskFormCheckItems,
    todoBadge: '巡检',
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
