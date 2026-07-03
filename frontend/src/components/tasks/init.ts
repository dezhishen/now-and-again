import { registerTaskKind } from '@/composables/useTaskKinds'
import SimpleTaskBody from '@/components/tasks/kinds/simple/SimpleTaskBody.vue'
import SimpleTodoActions from '@/components/tasks/kinds/simple/SimpleTodoActions.vue'
import InspectionTaskBody from '@/components/tasks/kinds/inspection/InspectionTaskBody.vue'
import InspectionInspect from '@/components/tasks/kinds/inspection/InspectionInspect.vue'
import InspectionTodoActions from '@/components/tasks/kinds/inspection/InspectionTodoActions.vue'
import InspectionTodoInfo from '@/components/tasks/kinds/inspection/InspectionTodoInfo.vue'
import TaskFormCheckItems from '@/components/tasks/TaskFormCheckItems.vue'
import ChainTaskBody from '@/components/tasks/kinds/chain/ChainTaskBody.vue'
import ChainTodoActions from '@/components/tasks/kinds/chain/ChainTodoActions.vue'
import ChainForm from '@/components/tasks/kinds/chain/ChainForm.vue'

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
    templateWizardStep: TaskFormCheckItems,
    wizardStepLabel: '巡检项配置',
  })

  registerTaskKind('chain', {
    card: ChainTaskBody,
    todoActions: ChainTodoActions,
    formComponent: ChainForm,
    labelKey: 'taskKind.chain',
    createLabelKey: 'taskKind.createChain',
    todoBadgeKey: 'taskKind.chain',
    buildDisplaySummary({ extra }) {
      const steps: any[] = extra?.steps || []
      if (steps.length === 0) return ''
      const names = steps.map((s: any) => s.name).filter(Boolean)
      if (names.length <= 3) {
        return names.join(' → ')
      }
      return names.slice(0, 3).join(' → ') + ` → 等${steps.length}项`
    },
    serializeExtra(formData) {
      return {
        steps: (formData || []).map((s: any) => ({
          name: s.task?.task?.name || '',
          kind: s.task?.task?.kind || 'simple',
          group_id: s.task?.task?.group_id || '',
          location_id: s.task?.task?.location_id || '',
        }))
      }
    },
    parseExtra(extra) {
      return (extra?.steps || []).map((s: any) => ({
        name: s.name,
        task: {
          task: {
            name: s.name,
            kind: s.kind || 'simple',
            schedule_type: 'once',
            schedule_data: { time: '09:00' },
            group_id: s.group_id || '',
            location_id: s.location_id || '',
          },
        },
      }))
    },
    defaultCheckItems: [],
  })
}
