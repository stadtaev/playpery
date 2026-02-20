import type { AdminMe, ScenarioSummary, ScenarioDetail, ScenarioRequest, GameSummary, GameDetail, GameRequest, TeamItem, TeamRequest } from './adminTypes'

const BASE = '/api/admin'

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    credentials: 'same-origin',
    ...opts,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export function login(email: string, password: string): Promise<AdminMe> {
  return request('/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
}

export function logout(): Promise<void> {
  return request('/logout', { method: 'POST' })
}

export function getMe(): Promise<AdminMe> {
  return request('/me')
}

export function listScenarios(): Promise<ScenarioSummary[]> {
  return request('/scenarios')
}

export function getScenario(id: string): Promise<ScenarioDetail> {
  return request(`/scenarios/${id}`)
}

export function createScenario(data: ScenarioRequest): Promise<ScenarioDetail> {
  return request('/scenarios', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export function updateScenario(id: string, data: ScenarioRequest): Promise<ScenarioDetail> {
  return request(`/scenarios/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export function deleteScenario(id: string): Promise<void> {
  return request(`/scenarios/${id}`, { method: 'DELETE' })
}

export function listGames(): Promise<GameSummary[]> {
  return request('/games')
}

export function getGame(id: string): Promise<GameDetail> {
  return request(`/games/${id}`)
}

export function createGame(data: GameRequest): Promise<GameDetail> {
  return request('/games', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export function updateGame(id: string, data: GameRequest): Promise<GameDetail> {
  return request(`/games/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export function deleteGame(id: string): Promise<void> {
  return request(`/games/${id}`, { method: 'DELETE' })
}

export function createTeam(gameId: string, data: TeamRequest): Promise<TeamItem> {
  return request(`/games/${gameId}/teams`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export function updateTeam(gameId: string, teamId: string, data: TeamRequest): Promise<TeamItem> {
  return request(`/games/${gameId}/teams/${teamId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export function deleteTeam(gameId: string, teamId: string): Promise<void> {
  return request(`/games/${gameId}/teams/${teamId}`, { method: 'DELETE' })
}
