<script setup>
import { computed } from 'vue'
const props = defineProps({ value: Number })

const state = computed(() => {
  if (props.value == null) return { color: 'text-[var(--text-secondary)]', bar: 'bg-[var(--edge)]', pct: 0 }
  if (props.value >= 55) return { color: 'text-danger', bar: 'bg-danger', pct: Math.min(100, (props.value / 70) * 100) }
  if (props.value >= 45) return { color: 'text-warm', bar: 'bg-warm', pct: (props.value / 70) * 100 }
  return { color: 'text-ok', bar: 'bg-ok', pct: (props.value / 70) * 100 }
})
</script>

<template>
  <div class="flex items-center gap-3">
    <span class="mono text-sm font-medium" :class="state.color">
      {{ value == null ? '--' : `${value}°C` }}
    </span>
    <div class="h-1.5 flex-1 rounded-full bg-white/[0.08] overflow-hidden" style="min-width: 40px; max-width: 120px;">
      <div
        class="h-full rounded-full transition-all duration-700"
        :class="state.bar"
        :style="{ width: state.pct + '%' }"
      ></div>
    </div>
  </div>
</template>
