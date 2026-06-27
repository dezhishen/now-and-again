import { registerLocationKind } from '@/composables/useLocationKinds'

export function initLocationKinds() {
  registerLocationKind('indoor', {
    label: '室内',
    icon: '🏠',
  })
}
