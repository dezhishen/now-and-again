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
    labelKey: 'taskKind.simple',
    createLabelKey: 'taskKind.create',
  })

  registerTaskKind('inspection', {
    card: InspectionTaskBody,
    inspectComponent: InspectionInspect,
    todoActions: InspectionTodoActions,
    todoInfo: InspectionTodoInfo,
    formComponent: TaskFormCheckItems,
    todoBadgeKey: 'taskKind.inspect',
    labelKey: 'taskKind.inspect',
    createLabelKey: 'taskKind.createInspect',
    buildDisplaySummary({ extra }) {
      const items: any[] = extra?.check_items || []
      const names = items.map((ci: any) => ci.name).filter(Boolean)
      return names.length > 0 ? '巡检: ' + names.join(', ') : ''
    },
    serializeExtra(items) {
      return { check_items: items }
    },
    parseExtra(extra) {
      return extra?.check_items || []
    },
    defaultCheckItems: [
      {
        name: '检查项1',
        branches: [
          { name: '正常', create_todo: false },
          { name: '异常', create_todo: true, branch_task: { task: { name: '', kind: 'simple', schedule_type: 'once' } as any } },
        ],
      },
    ],
  })
}
