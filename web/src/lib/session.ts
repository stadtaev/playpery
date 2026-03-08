/**
 * Session management with tab isolation.
 *
 * Session data (token, client, role, team) is stored in localStorage so it
 * survives tab closure. A per-tab pointer in sessionStorage tracks which
 * session is active in the current tab, preventing player/supervisor tabs
 * from overwriting each other.
 */

interface SessionData {
  token: string
  client: string
  teamId: string
  teamName: string
  role: string
  language?: string
}

const ACTIVE_KEY = 'cq_active_session'

function storageKey(teamId: string, role: string): string {
  return `cq_${teamId}_${role}`
}

export function saveSession(data: SessionData): void {
  const key = storageKey(data.teamId, data.role)
  localStorage.setItem(key, JSON.stringify(data))
  sessionStorage.setItem(ACTIVE_KEY, key)
}

export function getSession(): SessionData | null {
  const key = sessionStorage.getItem(ACTIVE_KEY)
  if (!key) return null
  const raw = localStorage.getItem(key)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export function clearSession(): void {
  const key = sessionStorage.getItem(ACTIVE_KEY)
  if (key) localStorage.removeItem(key)
  sessionStorage.removeItem(ACTIVE_KEY)
}
