<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api/client'
import HealthBadge from '../components/HealthBadge.vue'
import TemperatureBadge from '../components/TemperatureBadge.vue'
import AttributeTable from '../components/AttributeTable.vue'
import HistoryChart from '../components/HistoryChart.vue'
import { formatPowerHours, driveType, healthBorderAccent } from '../stores/format'

const route = useRoute()
const loading = ref(true)
const error = ref('')
const detail = ref(null)
const attributes = ref([])
const history = ref([])

const type = computed(() => detail.value ? driveType(detail.value.device) : '')

onMounted(async () => {
  try {
    const id = route.params.id
    const [d, a, h] = await Promise.all([api.drive(id), api.attributes(id), api.history(id)])
    detail.value = d
    attributes.value = a
    history.value = h
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section>
    <router-link
      to="/"
      class="rise mono inline-flex items-center gap-2 text-xs uppercase tracking-[0.15em] text-[var(--text-tertiary)] transition-colors hover:text-[var(--text-secondary)]"
    >
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
        <path d="M9 3L5 7L9 11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      Dashboard
    </router-link>

    <div v-if="loading" class="mt-6 rounded-xl border border-edge bg-panel p-6">
      <div class="flex items-center gap-3">
        <div class="h-4 w-4 rounded-full border-2 border-accent/40 border-t-accent animate-spin"></div>
        <span class="mono text-sm text-[var(--text-secondary)]">Loading drive data...</span>
      </div>
    </div>

    <div v-else-if="error" class="mt-6 rounded-xl border border-danger/40 bg-danger/5 p-6">
      <p class="mono text-sm text-danger">{{ error }}</p>
    </div>

    <div v-else class="mt-6 space-y-5">
      <!-- Drive header -->
      <div class="rise rounded-xl border bg-panel p-6" :class="healthBorderAccent(detail.health)">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <div class="flex items-center gap-2 mb-2">
              <span class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)]">{{ detail.device }}</span>
              <span
                class="mono rounded px-1 py-0.5 text-2xs uppercase"
                :class="type === 'nvme' ? 'bg-accent/10 text-accent/60' : 'bg-white/5 text-[var(--text-tertiary)]'"
              >{{ type }}</span>
            </div>
            <h2 class="text-2xl font-bold tracking-tight">{{ detail.model || detail.device }}</h2>
            <p class="mono mt-1 text-xs text-[var(--text-tertiary)]">{{ detail.serial || 'n/a' }}</p>
          </div>
          <HealthBadge :status="detail.health" />
        </div>
      </div>

      <!-- Metric cards -->
      <div class="rise grid gap-3 sm:grid-cols-2 lg:grid-cols-4" style="animation-delay: 80ms;">
        <article class="rounded-xl border border-edge bg-panel p-4">
          <p class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)]">Health Score</p>
          <p class="mt-2 text-2xl font-bold tabular-nums">{{ detail.health_score ?? '--' }}<span class="text-sm font-normal text-[var(--text-tertiary)]">/100</span></p>
        </article>

        <article class="rounded-xl border border-edge bg-panel p-4">
          <p class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)] mb-2">Temperature</p>
          <TemperatureBadge :value="detail.temperature" />
        </article>

        <article class="rounded-xl border border-edge bg-panel p-4">
          <p class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)]">Power On</p>
          <p class="mono mt-2 text-sm font-medium">{{ formatPowerHours(detail.power_on_hours) }}</p>
        </article>

        <article class="rounded-xl border border-edge bg-panel p-4">
          <p class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)]">Last Seen</p>
          <p class="mono mt-2 text-sm text-[var(--text-secondary)]">{{ detail.last_seen ? new Date(detail.last_seen).toLocaleString() : 'n/a' }}</p>
        </article>
      </div>

      <!-- Health reasons if any -->
      <div
        v-if="detail.health_reasons"
        class="rise rounded-xl border border-danger/20 bg-danger/5 p-4"
        style="animation-delay: 120ms;"
      >
        <p class="mono text-2xs uppercase tracking-wider text-danger/70 mb-2">Health Issues</p>
        <p class="text-sm text-danger/90">{{ detail.health_reasons }}</p>
      </div>

      <HistoryChart :points="history" class="rise" style="animation-delay: 160ms;" />
      <AttributeTable :rows="attributes" class="rise" style="animation-delay: 200ms;" />
    </div>
  </section>
</template>
