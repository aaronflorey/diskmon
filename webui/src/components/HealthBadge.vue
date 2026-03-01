<script setup>
import { computed } from 'vue'
import { healthColor, healthLabel, healthIcon } from '../stores/format'
const props = defineProps({ status: { type: String, default: 'UNKNOWN' }, compact: Boolean })
const klass = computed(() => healthColor(props.status))
const label = computed(() => healthLabel(props.status))
const icon = computed(() => healthIcon(props.status))

const bgClass = computed(() => {
  if (props.status === 'RED') return 'bg-danger/10'
  if (props.status === 'YELLOW') return 'bg-warm/10'
  if (props.status === 'GREEN') return 'bg-ok/8'
  return 'bg-white/5'
})

const dotClass = computed(() => {
  if (props.status === 'RED') return 'bg-danger'
  if (props.status === 'YELLOW') return 'bg-warm'
  if (props.status === 'GREEN') return 'bg-ok'
  return 'bg-[var(--text-tertiary)]'
})
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-xs font-medium"
    :class="[klass, bgClass]"
  >
    <span class="h-1.5 w-1.5 rounded-full" :class="dotClass"
      :style="status === 'RED' ? 'animation: pulse-dot 1.5s ease-in-out infinite' : ''"
    ></span>
    <span v-if="!compact">{{ label }}</span>
    <span v-else class="mono">{{ icon }}</span>
  </span>
</template>
