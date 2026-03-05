async function req(path) {
  const res = await fetch(`/api/v1${path}`)
  if (!res.ok) {
    throw new Error(`API error ${res.status}`)
  }
  return res.json()
}

export const api = {
  drives: () => req('/drives'),
  drive: (id) => req(`/drives/${id}`),
  history: (id) => req(`/drives/${id}/history`),
  attributes: (id) => req(`/drives/${id}/attributes`),
  tests: (id, page = 1, pageSize = 10) => req(`/drives/${id}/tests?page=${page}&page_size=${pageSize}`)
}
