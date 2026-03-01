<script setup>
import { computed } from 'vue'

const props = defineProps({ points: { type: Array, default: () => [] } })

const chartData = computed(() => {
  const pts = [...props.points].reverse().filter((p) => p.temperature != null)
  if (pts.length < 2) return null

  const temps = pts.map((p) => p.temperature)
  const min = Math.min(...temps)
  const max = Math.max(...temps)
  const span = Math.max(1, max - min)
  const padded = span * 0.1
  const yMin = min - padded
  const yMax = max + padded
  const ySpan = yMax - yMin

  const coords = pts.map((p, i) => ({
    x: (i / (pts.length - 1)) * 100,
    y: 100 - ((p.temperature - yMin) / ySpan) * 100
  }))

  const path = coords.map((c) => `${c.x},${c.y}`).join(' ')

  // Area fill path
  const area = `0,100 ${path} 100,100`

  // Y-axis labels
  const yLabels = []
  const step = span <= 3 ? 1 : Math.ceil(span / 4)
  for (let v = Math.floor(min); v <= Math.ceil(max); v += step) {
    yLabels.push({
      value: v,
      y: 100 - ((v - yMin) / ySpan) * 100
    })
  }

  return { path, area, yLabels, min, max, latest: temps[temps.length - 1] }
})
</script>

<template>
  <div class="rounded-xl border border-edge bg-panel p-5">
    <div class="flex items-center justify-between mb-4">
      <p class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)]">Temperature History</p>
      <p v-if="chartData" class="mono text-xs text-[var(--text-secondary)]">
        {{ chartData.min }}° — {{ chartData.max }}°
      </p>
    </div>

    <div class="relative">
      <svg v-if="chartData" viewBox="-8 -4 116 108" class="h-44 w-full overflow-visible" preserveAspectRatio="none">
        <!-- Horizontal grid lines -->
        <line
          v-for="label in chartData.yLabels" :key="label.value"
          x1="0" x2="100" :y1="label.y" :y2="label.y"
          stroke="var(--edge-subtle)" stroke-width="0.3"
        />

        <!-- Y-axis labels -->
        <text
          v-for="label in chartData.yLabels" :key="'t' + label.value"
          x="-4" :y="label.y + 1.5"
          text-anchor="end" fill="var(--text-tertiary)" font-size="3.5"
          font-family="'IBM Plex Mono', monospace"
        >{{ label.value }}°</text>

        <!-- Area fill -->
        <polygon
          :points="chartData.area"
          fill="url(#areaGrad)"
        />

        <!-- Line -->
        <polyline
          :points="chartData.path"
          fill="none"
          stroke="#c8ff3e"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          vector-effect="non-scaling-stroke"
        />

        <defs>
          <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#c8ff3e" stop-opacity="0.12" />
            <stop offset="100%" stop-color="#c8ff3e" stop-opacity="0" />
          </linearGradient>
        </defs>
      </svg>

      <div v-else class="flex h-44 items-center justify-center">
        <p class="mono text-xs text-[var(--text-tertiary)]">No temperature data available</p>
      </div>
    </div>
  </div>
</template>
