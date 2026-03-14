// Session management — init player, persist token in localStorage

const STORAGE_KEY = 'hexslayer_session'

export function getSession() {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function saveSession(session) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
}

export function clearSession() {
  localStorage.removeItem(STORAGE_KEY)
}

export async function initSession() {
  const existing = getSession()
  if (existing?.sessionToken) {
    // Validate the token is still valid against an auth-protected endpoint
    const check = await fetch('/api/map/zones?lat=0&lng=0', {
      headers: { Authorization: `Bearer ${existing.sessionToken}` },
    }).catch(() => null)
    if (check && check.status !== 401) return existing
    // Token invalid — clear and re-init
    clearSession()
  }

  const res = await fetch('/api/player/init', { method: 'POST' })
  if (!res.ok) throw new Error('Failed to init player')

  const data = await res.json()
  const session = {
    playerId: data.playerId,
    sessionToken: data.sessionToken,
    name: data.name,
  }
  saveSession(session)
  return session
}
