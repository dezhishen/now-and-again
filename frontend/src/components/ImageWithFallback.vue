<script setup lang="ts">
import { ref, computed } from 'vue'

const props = withDefaults(defineProps<{
  /** Image ID — will be fetched via /api/images/:id (server handles 302 redirect). */
  imageId?: string | null
  /** Fallback: direct src URL (used if imageId is not provided). */
  src?: string | null
  /** Alt text for the image. */
  alt?: string
  /** Text used for the fallback placeholder (e.g. first letter of name). */
  fallbackText?: string
  /** Container aspect ratio class, e.g. 'aspect-video', 'aspect-square'. */
  aspectRatio?: string
  /** Border radius class, e.g. 'rounded-lg', 'rounded-full'. */
  rounded?: string
  /** Extra classes applied to the container. */
  class?: string
}>(), {
  imageId: undefined,
  src: undefined,
  alt: '',
  fallbackText: '',
  aspectRatio: 'aspect-video',
  rounded: 'rounded-lg',
  class: '',
})

const loaded = ref(false)
const errored = ref(false)

/** Resolved src: prefer imageId (API endpoint) over raw src. */
const resolvedSrc = computed(() => {
  if (props.imageId) return `/api/images/${props.imageId}`
  return props.src || undefined
})

const showFallback = computed(() => !resolvedSrc.value || errored.value)

function onLoad() {
  loaded.value = true
  errored.value = false
}

function onError() {
  errored.value = true
}

const initial = computed(() => {
  if (props.fallbackText) return props.fallbackText.charAt(0).toUpperCase()
  return '?'
})
</script>

<template>
  <div
    :class="[aspectRatio, rounded, props.class]"
    class="overflow-hidden"
    :aria-label="alt || fallbackText || 'image'"
  >
    <!-- Image -->
    <img
      v-if="!showFallback"
      :src="resolvedSrc"
      :alt="alt"
      class="w-full h-full object-cover transition-opacity duration-300"
      :class="{ 'opacity-0': !loaded, 'opacity-100': loaded }"
      loading="lazy"
      @load="onLoad"
      @error="onError"
    />

    <!-- Fallback placeholder -->
    <div
      v-else
      class="w-full h-full bg-gradient-to-br from-primary/10 to-primary/5 dark:from-primary/20 dark:to-gray-800 flex items-center justify-center"
    >
      <span class="text-4xl opacity-30 select-none">{{ initial }}</span>
    </div>
  </div>
</template>
