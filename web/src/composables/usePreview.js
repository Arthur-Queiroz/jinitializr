// fetchPreview asks the backend which files the given project config would
// produce. It returns the flat, sorted list of paths from POST /api/preview.
// The generator owns this list; the frontend never recomputes it.
export async function fetchPreview(config) {
  const res = await fetch('/api/preview', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const data = await res.json()
  return data.files ?? []
}
