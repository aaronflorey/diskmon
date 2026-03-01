export function healthColor(status) {
  if (status === 'RED') return 'text-danger border-danger/40'
  if (status === 'YELLOW') return 'text-warm border-warm/40'
  if (status === 'GREEN') return 'text-ok border-ok/40'
  return 'text-[var(--text-secondary)] border-[var(--edge)]'
}

export function healthGlow(status) {
  if (status === 'RED') return 'glow-danger pulse-danger'
  if (status === 'YELLOW') return 'glow-warn'
  if (status === 'GREEN') return 'glow-ok'
  return ''
}

export function healthBorderAccent(status) {
  if (status === 'RED') return 'border-danger/25'
  if (status === 'YELLOW') return 'border-warm/20'
  if (status === 'GREEN') return 'border-ok/15'
  return 'border-edge'
}

export function healthLabel(status) {
  if (status === 'RED') return 'Critical'
  if (status === 'YELLOW') return 'Warning'
  if (status === 'GREEN') return 'Healthy'
  return 'Unknown'
}

export function healthIcon(status) {
  if (status === 'RED') return '!'
  if (status === 'YELLOW') return '~'
  if (status === 'GREEN') return '+'
  return '?'
}

export function tempText(v) {
  return v === null || v === undefined ? 'n/a' : `${v}°C`
}

export function formatPowerHours(hours) {
  if (hours == null) return 'n/a'
  if (hours < 1000) return `${hours}h`
  const years = (hours / 8760).toFixed(1)
  return `${hours.toLocaleString()}h (${years}y)`
}

export function driveType(device) {
  if (!device) return 'unknown'
  if (device.includes('nvme')) return 'nvme'
  return 'hdd'
}
