import type { TeamLookup, JoinResponse, GameState, AnswerResponse, UnlockResponse } from './types'

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(path, opts)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

function authHeaders(): HeadersInit {
  const token = localStorage.getItem('session_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export function lookupTeam(client: string, joinToken: string): Promise<TeamLookup> {
  return request(`/api/${client}/teams/${joinToken}`)
}

export function joinTeam(client: string, joinToken: string, playerName: string): Promise<JoinResponse> {
  return request(`/api/${client}/join`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ joinToken, playerName }),
  })
}

export function getGameState(client: string): Promise<GameState> {
  return request(`/api/${client}/game/state`, { headers: authHeaders() })
}

export function submitAnswer(client: string, answer: string): Promise<AnswerResponse> {
  return request(`/api/${client}/game/answer`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ answer }),
  })
}

export function unlockStage(client: string, code: string): Promise<UnlockResponse> {
  return request(`/api/${client}/game/unlock`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ code }),
  })
}
