import { useState, useEffect } from 'react'
import { lookupTeam, joinTeam } from './api'
import type { TeamLookup } from './types'

export function JoinPage({ joinToken }: { joinToken: string }) {
  const [team, setTeam] = useState<TeamLookup | null>(null)
  const [error, setError] = useState('')
  const [name, setName] = useState('')
  const [joining, setJoining] = useState(false)

  useEffect(() => {
    lookupTeam(joinToken)
      .then(setTeam)
      .catch((e) => setError(e.message))
  }, [joinToken])

  async function handleJoin(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setJoining(true)
    setError('')
    try {
      const resp = await joinTeam(joinToken, name.trim())
      localStorage.setItem('session_token', resp.token)
      localStorage.setItem('team_name', resp.teamName)
      window.history.replaceState(null, '', '/game')
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to join')
      setJoining(false)
    }
  }

  if (error && !team) {
    return (
      <main className="container">
        <h1>CityQuiz</h1>
        <p role="alert">{error}</p>
      </main>
    )
  }

  if (!team) {
    return (
      <main className="container">
        <p aria-busy="true">Loading team...</p>
      </main>
    )
  }

  return (
    <main className="container" style={{ maxWidth: 480 }}>
      <h1>CityQuiz</h1>
      <hgroup>
        <h2>Join {team.name}</h2>
        <p>{team.gameName}</p>
      </hgroup>
      <form onSubmit={handleJoin}>
        <label>
          Your name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter your name"
            autoFocus
            required
          />
        </label>
        {error && <small style={{ color: 'var(--pico-color-red-500)' }}>{error}</small>}
        <button type="submit" disabled={joining} aria-busy={joining}>
          {joining ? 'Joining...' : 'Join Game'}
        </button>
      </form>
    </main>
  )
}
