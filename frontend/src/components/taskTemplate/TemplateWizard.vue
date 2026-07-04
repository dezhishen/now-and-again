<script setup lang="ts">
import { ref, watch, computed, markRaw, type Component } from 'vue'
import { useI18n } from '@/i18n'
import { useToast } from '@/composables/useToast'
import { useErrorHandler } from '@/composables/useErrorHandler'
import ErrorDisplay from '@/components/ErrorDisplay.vue'
import { createFamilyTemplate, updateFamilyTemplate } from '@/api/task-templates'
import { getTaskKind } from '@/composables/useTaskKinds'
import type { TaskTemplate, TemplateParameter } from '@/types'

import BasicInfoStep from './steps/BasicInfoStep.vue'
import RawEditorStep from './steps/RawEditorStep.vue'
import ParamsStep from './steps/ParamsStep.vue'
import PreviewStep from './steps/PreviewStep.vue'

const { t } = useI18n()
const toast = useToast()
const { error, setError, clearError } = useErrorHandler()

const props = defineProps<{
  editing?: TaskTemplate | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

// ── Wizard state ──────────────────────────────────────────────────

const currentStep = ref(0)
const saving = ref(false)

// Step 1: basic info
const templateCode = ref('')
const name = ref('')
const description = ref('')
const kind = ref('simple')
const icon = ref('')
const sortOrder = ref(0)
const enabled = ref(true)
const codeManuallyEdited = ref(false)

// Step 2: extra data (owned by plugin or fallback YAML)
const extraData = ref<any>(undefined)     // plugin-owned state
const extraSchemaRaw = ref<any>({})        // fallback YAML state
const taskDefaultsRaw = ref<any>({})       // fallback YAML state

// Step 3: parameters
const parameters = ref<TemplateParameter[]>([])

// ── Step definitions (zero kind-specific branching) ──────────────

interface StepDef {
  id: string
  label: string
  component: Component
}

const steps = computed<StepDef[]>(() => {
  const list: StepDef[] = [
    { id: 'basic', label: '基本信息', component: markRaw(BasicInfoStep) },
    { id: 'params', label: '参数配置', component: markRaw(ParamsStep) },
  ]

  const plugin = getTaskKind(kind.value)
  if (plugin?.templateWizardStep) {
    list.push({ id: 'extra', label: plugin.wizardStepLabel || '类型配置', component: markRaw(plugin.templateWizardStep) })
  } else {
    list.push({ id: 'extra', label: '高级配置', component: markRaw(RawEditorStep) })
  }

  list.push(
    { id: 'preview', label: '预览确认', component: markRaw(PreviewStep) },
  )
  return list
})

const totalSteps = computed(() => steps.value.length)
const isFirstStep = computed(() => currentStep.value === 0)
const isLastStep = computed(() => currentStep.value === totalSteps.value - 1)

// ── Initialize from editing template ──────────────────────────────

watch(() => props.editing, (tmpl) => {
  if (!tmpl) return
  templateCode.value = tmpl.template_code
  name.value = tmpl.name
  description.value = tmpl.description || ''
  kind.value = tmpl.kind
  icon.value = tmpl.icon || ''
  sortOrder.value = tmpl.sort_order
  enabled.value = tmpl.enabled
  parameters.value = tmpl.parameters ? JSON.parse(JSON.stringify(tmpl.parameters)) : []

  const es = tmpl.extra_schema ? JSON.parse(JSON.stringify(tmpl.extra_schema)) : undefined
  const td = tmpl.task_defaults ? JSON.parse(JSON.stringify(tmpl.task_defaults)) : {}

  const plugin = getTaskKind(tmpl.kind)
  if (plugin?.templateWizardStep && plugin.parseExtra && es) {
    extraData.value = plugin.parseExtra(es)
  } else {
    extraSchemaRaw.value = es || {}
    taskDefaultsRaw.value = td
  }
}, { immediate: true })

// Reset extra state when kind changes (for new templates)
watch(kind, (newKind) => {
  if (props.editing) return // don't reset when editing
  const plugin = getTaskKind(newKind)
  if (plugin?.templateWizardStep) {
    extraData.value = plugin.parseExtra?.({}) ?? plugin.defaultCheckItems ?? []
    taskDefaultsRaw.value = {}
  } else {
    extraSchemaRaw.value = {}
    taskDefaultsRaw.value = {}
  }
})

// ── Navigation ────────────────────────────────────────────────────

function next() {
  if (currentStep.value < totalSteps.value - 1) {
    currentStep.value++
  }
}

function prev() {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

// ── Save ──────────────────────────────────────────────────────────

async function handleSave() {
  saving.value = true
  clearError()
  try {
    const plugin = getTaskKind(kind.value)

    let task_defaults: any
    let extra_schema: any

    if (plugin?.templateWizardStep && plugin.serializeExtra) {
      // Plugin visual editor → serialize
      extra_schema = plugin.serializeExtra(extraData.value ?? [])
      task_defaults = taskDefaultsRaw.value
    } else {
      // Fallback: raw YAML → already in reactive state
      extra_schema = extraSchemaRaw.value
      task_defaults = taskDefaultsRaw.value
    }

    if (props.editing) {
      await updateFamilyTemplate(props.editing.template_code, {
        name: name.value,
        description: description.value,
        kind: kind.value,
        icon: icon.value,
        sort_order: sortOrder.value,
        enabled: enabled.value,
        parameters: parameters.value,
        task_defaults,
        extra_schema,
      })
      toast.success(t('taskTemplate.updated'))
    } else {
      await createFamilyTemplate({
        template_code: templateCode.value,
        name: name.value,
        description: description.value,
        kind: kind.value,
        icon: icon.value,
        sort_order: sortOrder.value,
        enabled: enabled.value,
        parameters: parameters.value,
        task_defaults,
        extra_schema,
      })
      toast.success(t('taskTemplate.created'))
    }
    emit('saved')
  } catch (e: any) {
    setError(e)
  } finally {
    saving.value = false
  }
}

// ── Step component binding (zero kind-specific logic) ────────────

const currentStepDef = computed(() => steps.value[currentStep.value])

/** Returns the correct v-model binding based on step id */
const stepModel = computed<Record<string, any>>(() => {
  const id = currentStepDef.value?.id
  switch (id) {
    case 'basic':
      return {
        templateCode: templateCode.value, name: name.value, description: description.value,
        kind: kind.value, icon: icon.value, sortOrder: sortOrder.value, enabled: enabled.value,
        codeManuallyEdited: codeManuallyEdited.value,
        editing: props.editing,
      }
    case 'extra':
      // Check if plugin has a component → bind extraData; otherwise bind raw YAML
      if (getTaskKind(kind.value)?.templateWizardStep) {
        return { modelValue: extraData }
      }
      return { taskDefaults: taskDefaultsRaw, extraSchema: extraSchemaRaw }
    case 'params':
      return { parameters }
    case 'preview':
      return {
        templateCode: templateCode.value,
        name: name.value,
        description: description.value,
        kind: kind.value,
        icon: icon.value,
        sortOrder: sortOrder.value,
        enabled: enabled.value,
        parameters: parameters.value,
        taskDefaults: getTaskKind(kind.value)?.templateWizardStep ? taskDefaultsRaw.value : taskDefaultsRaw.value,
        extraSchema: getTaskKind(kind.value)?.templateWizardStep
          ? (getTaskKind(kind.value)?.serializeExtra?.(extraData.value ?? []) ?? {})
          : extraSchemaRaw.value,
      }
    default:
      return {}
  }
})

// Fun fact: for "extra" step with plugin component, we use :model-value / @update:model-value
// For fallback, we pass taskDefaults + extraSchema as props with .sync
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" v-esc="() => emit('close')">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-hidden flex flex-col">
      <!-- Header -->
      <div class="px-6 pt-5 pb-0">
        <ErrorDisplay :error="error" @close="clearError" />
        <h4 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">
          {{ editing ? t('taskTemplate.editFamily') : t('taskTemplate.createFamily') }}
        </h4>
        <!-- Step indicator -->
        <div class="flex items-center gap-2 mb-3">
          <template v-for="(s, i) in steps" :key="s.id">
            <span
              class="text-xs px-2 py-0.5 rounded-full transition-colors"
              :class="i === currentStep ? 'bg-primary text-white' : i < currentStep ? 'bg-primary/10 text-primary' : 'bg-gray-100 dark:bg-gray-700 text-gray-400'"
            >{{ i + 1 }}. {{ s.label }}</span>
            <span v-if="i < steps.length - 1" class="text-gray-300 dark:text-gray-600">→</span>
          </template>
        </div>
      </div>

      <!-- Step content -->
      <div class="flex-1 overflow-y-auto px-6 py-4">
        <!-- Step 1: BasicInfo -->
        <BasicInfoStep
          v-if="currentStepDef?.id === 'basic'"
          v-model:templateCode="templateCode"
          v-model:name="name"
          v-model:description="description"
          v-model:kind="kind"
          v-model:icon="icon"
          v-model:sortOrder="sortOrder"
          v-model:enabled="enabled"
          v-model:codeManuallyEdited="codeManuallyEdited"
          :editing="editing"
        />

        <!-- Step 2: Params -->
        <ParamsStep
          v-else-if="currentStepDef?.id === 'params'"
          v-model:parameters="parameters"
        />

        <!-- Step 3: Plugin visual editor OR fallback YAML -->
        <template v-else-if="currentStepDef?.id === 'extra'">
          <component
            v-if="getTaskKind(kind)?.templateWizardStep"
            :is="currentStepDef.component"
            :model-value="extraData"
            @update:model-value="extraData = $event"
          />
          <component
            v-else
            :is="currentStepDef.component"
            v-model:taskDefaults="taskDefaultsRaw"
            v-model:extraSchema="extraSchemaRaw"
            :parameters="parameters"
          />
        </template>

        <!-- Step 4: Preview -->
        <PreviewStep
          v-else-if="currentStepDef?.id === 'preview'"
          v-bind="(stepModel as any)"
        />
      </div>

      <!-- Footer -->
      <div class="flex justify-between px-6 py-4 border-t dark:border-gray-700">
        <button
          class="btn-secondary text-sm"
          @click="emit('close')"
        >{{ t('confirm.cancel') }}</button>
        <div class="flex gap-2">
          <button
            v-if="!isFirstStep"
            class="btn-secondary text-sm"
            @click="prev"
          >← 上一步</button>
          <button
            v-if="!isLastStep"
            class="btn-primary text-sm"
            @click="next"
          >下一步 →</button>
          <button
            v-if="isLastStep"
            class="btn-primary text-sm"
            :disabled="saving"
            @click="handleSave"
          >{{ saving ? '...' : t('confirm.ok') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
