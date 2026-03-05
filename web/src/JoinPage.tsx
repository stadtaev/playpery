import { useState, useEffect } from 'react'
import { lookupTeam, joinTeam } from './api'
import type { TeamLookup } from './types'
import { saveSession } from './lib/session'
import { PageContainer } from './components/PageContainer'
import { LoadingPage, Spinner } from './components/Spinner'
import { ErrorMessage } from './components/ErrorMessage'

export function JoinPage({ client, joinToken }: { client: string; joinToken: string }) {
  const [team, setTeam] = useState<TeamLookup | null>(null)
  const [error, setError] = useState('')
  const [name, setName] = useState('')
  const [joining, setJoining] = useState(false)

  useEffect(() => {
    lookupTeam(client, joinToken)
      .then(setTeam)
      .catch((e) => setError(e.message))
  }, [client, joinToken])

  async function handleJoin(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setJoining(true)
    setError('')
    try {
      const resp = await joinTeam(client, joinToken, name.trim())
      saveSession({
        token: resp.token,
        client,
        teamId: resp.teamId,
        teamName: resp.teamName,
        role: resp.role,
      })
      window.history.replaceState(null, '', '/game')
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to join')
      setJoining(false)
    }
  }

  if (error && !team) {
    return (
      <PageContainer>
        <h1>CityQuest</h1>
        <ErrorMessage message={error} />
      </PageContainer>
    )
  }

  if (!team) {
    return <LoadingPage message="Loading team..." />
  }

  return (
    <PageContainer>
      <h1>CityQuest</h1>
      <div className="mb-6">
        <h2 className="mb-1">Join {team.name}</h2>
        <p className="text-secondary">{team.gameName}</p>
      </div>
      {team.role === 'supervisor' && (
        <p>
          <span className="inline-block bg-primary text-white text-xs font-bold uppercase tracking-widest px-3 py-1">
            Joining as Supervisor
          </span>
        </p>
      )}
      <form onSubmit={handleJoin} className="space-y-4">
        <div>
          <label className="input-label" htmlFor="player-name">Your name</label>
          <input
            id="player-name"
            className="input"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter your name"
            autoFocus
            required
          />
        </div>
        {error && <p className="text-feedback-error">{error}</p>}
        <button type="submit" disabled={joining} className="btn btn-accent w-full">
          {joining ? <Spinner /> : 'Join Game'}
        </button>
      </form>
    </PageContainer>
  )
}
