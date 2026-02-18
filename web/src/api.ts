import type { TeamLookup, JoinResponse, GameState, AnswerResponse } from './types'

const BASE = '/api'

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, opts)
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

export function lookupTeam(joinToken: string): Promise<TeamLookup> {
  return request(`/teams/${joinToken}`)
}

export function joinTeam(joinToken: string, playerName: string): Promise<JoinResponse> {
  return request('/join', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ joinToken, playerName }),
  })
}

export function getGameState(): Promise<GameState> {
  return request('/game/state', { headers: authHeaders() })
}

export function submitAnswer(answer: string): Promise<AnswerResponse> {
  return request('/game/answer', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ answer }),
  })
}
