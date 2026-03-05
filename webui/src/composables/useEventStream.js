import { onUnmounted } from 'vue'

export function useEventStream(events, onEvent, { debounceMs = 300, filterDevice = null } = {}) {
  let sse = null
  let timer = null

  const schedule = () => {
    if (timer) return
    timer = setTimeout(async () => {
      timer = null
      await onEvent()
    }, debounceMs)
  }

  const makeHandler = () => {
    if (!filterDevice) return schedule
    return (event) => {
      try {
        const payload = JSON.parse(event.data || '{}')
        const device = filterDevice()
        if (device && payload.device && payload.device !== device) return
      } catch {}
      schedule()
    }
  }

  function connect() {
    sse = new EventSource('/api/v1/events')
    const handler = makeHandler()
    for (const ev of events) {
      sse.addEventListener(ev, handler)
    }
    sse.onerror = () => {}
  }

  function disconnect() {
    if (sse) {
      sse.close()
      sse = null
    }
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  onUnmounted(disconnect)

  return { connect, disconnect }
}
