<script setup>
const props = defineProps({ rows: { type: Array, default: () => [] } })

function statusClass(s) {
  if (s === 'RED') return 'text-danger'
  if (s === 'YELLOW') return 'text-warm'
  return 'text-ok/70'
}

function statusDot(s) {
  if (s === 'RED') return 'bg-danger'
  if (s === 'YELLOW') return 'bg-warm'
  return 'bg-ok/50'
}

function explanation(row) {
  if (row.status === 'RED') return 'Failed'
  if (row.status === 'YELLOW') return 'Near threshold'
  return 'OK'
}
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-edge bg-panel">
    <div class="px-5 py-3 border-b border-edge/50">
      <p class="mono text-2xs uppercase tracking-wider text-[var(--text-tertiary)]">SMART Attributes</p>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full border-collapse text-sm">
        <thead>
          <tr class="border-b border-edge/40">
            <th class="mono px-5 py-2.5 text-left text-2xs font-medium uppercase tracking-wider text-[var(--text-tertiary)]">Attribute</th>
            <th class="mono px-4 py-2.5 text-right text-2xs font-medium uppercase tracking-wider text-[var(--text-tertiary)]">Value</th>
            <th class="mono px-4 py-2.5 text-right text-2xs font-medium uppercase tracking-wider text-[var(--text-tertiary)]">Raw</th>
            <th class="mono px-4 py-2.5 text-right text-2xs font-medium uppercase tracking-wider text-[var(--text-tertiary)]">Thresh</th>
            <th class="mono px-4 py-2.5 text-center text-2xs font-medium uppercase tracking-wider text-[var(--text-tertiary)]">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in props.rows"
            :key="row.attribute_id"
            class="border-t border-edge/20 transition-colors hover:bg-white/[0.02]"
            :class="row.status === 'RED' ? 'bg-danger/[0.03]' : row.status === 'YELLOW' ? 'bg-warm/[0.02]' : ''"
          >
            <td class="px-5 py-2.5 text-[var(--text-primary)]">{{ row.name }}</td>
            <td class="mono px-4 py-2.5 text-right text-[var(--text-primary)]">{{ row.value }}</td>
            <td class="mono px-4 py-2.5 text-right text-[var(--text-secondary)]">{{ row.raw }}</td>
            <td class="mono px-4 py-2.5 text-right text-[var(--text-tertiary)]">{{ row.threshold || '—' }}</td>
            <td class="px-4 py-2.5">
              <div class="flex items-center justify-center gap-1.5">
                <span class="h-1.5 w-1.5 rounded-full" :class="statusDot(row.status)"></span>
                <span class="mono text-xs" :class="statusClass(row.status)">{{ explanation(row) }}</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="!props.rows.length" class="px-5 py-8 text-center">
      <p class="mono text-xs text-[var(--text-tertiary)]">No SMART attributes available</p>
    </div>
  </div>
</template>
