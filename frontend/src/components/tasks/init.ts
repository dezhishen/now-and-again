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
      if (items.length === 0) return ''
      const parts: string[] = []
      for (const ci of items) {
        if (!ci.name) continue
        const subBranches = (ci.branches || [])
          .filter((b: any) => b.create_todo && b.branch_task?.task?.name)
          .map((b: any) => b.branch_task.task.name)
        if (subBranches.length > 0) {
          parts.push(ci.name + '→' + subBranches.join(','))
        } else {
          parts.push(ci.name)
        }
      }
      return parts.length > 0 ? '巡检: ' + parts.join('; ') : ''
    },
    serializeExtra(items) {
      return { check_items: items }
    },
    parseExtra(extra) {
      return extra?.check_items || []
    },
    defaultCheckItems: [],
  })
}
