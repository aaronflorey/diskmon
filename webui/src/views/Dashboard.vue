<script setup>
import { onMounted, ref, computed } from 'vue'
import { api } from '../api/client'
import DriveCard from '../components/DriveCard.vue'
import { driveType } from '../stores/format'
import { useEventStream } from '../composables/useEventStream'

const loading = ref(true)
const error = ref('')
const drives = ref([])

const grouped = computed(() => {
  const groups = {}
  drives.value.forEach((d) => {
    const type = driveType(d.device)
    if (!groups[type]) groups[type] = []
    groups[type].push(d)
  })
  // Sort: NVMe first, then HDD, then anything else
  const order = ['nvme', 'hdd', 'unknown']
  return order
    .filter((k) => groups[k]?.length)
    .map((k) => ({ type: k, drives: groups[k] }))
})

const stats = computed(() => {
  const all = drives.value
  return {
    total: all.length,
    healthy: all.filter((d) => d.health === 'GREEN').length,
    warning: all.filter((d) => d.health === 'YELLOW').length,
    critical: all.filter((d) => d.health === 'RED').length
  }
})

const labels = {
  nvme: 'NVMe Drives',
  hdd: 'Hard Drives',
  unknown: 'Other Devices'
}

async function reload(showLoading = false) {
  if (showLoading) loading.value = true
  try {
    drives.value = await api.drives()
    error.value = ''
  } catch (err) {
    error.value = err.message
  } finally {
    if (showLoading) loading.value = false
  }
}

const { connect } = useEventStream(
  ['sample.inserted', 'test.updated'],
  () => reload(false),
  { debounceMs: 300 }
)

onMounted(async () => {
  await reload(true)
  connect()
})
</script>

<template>
  <section>
    <div v-if="loading" class="rounded-xl border border-edge bg-panel p-6">
      <div class="flex items-center gap-3">
        <div class="h-4 w-4 rounded-full border-2 border-accent/40 border-t-accent animate-spin"></div>
        <span class="mono text-sm text-[var(--text-secondary)]">Scanning drives...</span>
      </div>
    </div>

    <div v-else-if="error" class="rounded-xl border border-danger/40 bg-danger/5 p-6">
      <p class="mono text-sm text-danger">{{ error }}</p>
    </div>

    <div v-else>
      <!-- Fleet overview bar -->
      <div class="rise mb-8 flex flex-wrap items-center gap-6 rounded-xl border border-edge bg-panel px-5 py-3">
        <div class="flex items-center gap-2">
          <span class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)]">Fleet</span>
          <span class="mono text-sm font-medium">{{ stats.total }} drives</span>
        </div>
        <div class="h-4 w-px bg-edge"></div>
        <div class="flex items-center gap-4">
          <div v-if="stats.healthy" class="flex items-center gap-1.5">
            <span class="h-1.5 w-1.5 rounded-full bg-ok"></span>
            <span class="mono text-xs text-ok/80">{{ stats.healthy }}</span>
          </div>
          <div v-if="stats.warning" class="flex items-center gap-1.5">
            <span class="h-1.5 w-1.5 rounded-full bg-warm"></span>
            <span class="mono text-xs text-warm/80">{{ stats.warning }}</span>
          </div>
          <div v-if="stats.critical" class="flex items-center gap-1.5">
            <span class="h-1.5 w-1.5 rounded-full bg-danger" style="animation: pulse-dot 1.5s ease-in-out infinite"></span>
            <span class="mono text-xs text-danger/80">{{ stats.critical }}</span>
          </div>
        </div>
      </div>

      <!-- Grouped drive sections -->
      <div v-for="(group, gi) in grouped" :key="group.type" :class="gi > 0 ? 'mt-10' : ''">
        <div class="rise mb-4 flex items-center gap-3" :style="{ animationDelay: `${gi * 100}ms` }">
          <h2 class="mono text-xs font-medium uppercase tracking-[0.2em] text-[var(--text-tertiary)]">
            {{ labels[group.type] || group.type }}
          </h2>
          <div class="h-px flex-1 bg-edge/50"></div>
          <span class="mono text-2xs text-[var(--text-tertiary)]">{{ group.drives.length }}</span>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          <DriveCard
            v-for="(drive, di) in group.drives"
            :key="drive.id"
            :drive="drive"
            :index="gi * 10 + di"
          />
        </div>
      </div>
    </div>
  </section>
</template>
