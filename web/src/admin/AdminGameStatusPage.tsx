import { useState, useEffect } from 'react'
import { getGameStatus } from './adminApi'
import type { GameStatus } from './adminTypes'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminGameStatusPage({ client, id }: { client: string; id: string }) {
  const [game, setGame] = useState<GameStatus | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true

    function load() {
      getGameStatus(client, id)
        .then((g) => { if (active) setGame(g) })
        .catch((e) => { if (active) setError(e.message) })
    }

    load()
    const interval = setInterval(load, 5000)
    return () => { active = false; clearInterval(interval) }
  }, [client, id])

  if (error) return <p role="alert" style={{ color: 'var(--pico-color-red-500)' }}>{error}</p>
  if (!game) return <p aria-busy="true">Loading...</p>

  const totalPlayers = game.teams.reduce((sum, t) => sum + t.players.length, 0)

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: '1rem', flexWrap: 'wrap' }}>
        <h2 style={{ margin: 0 }}>{game.scenarioName}</h2>
        <span>
          <strong>Status:</strong> {game.status}
        </span>
        <span>
          <strong>Timer:</strong> {game.timerMinutes}m
        </span>
        <span>
          <strong>Players:</strong> {totalPlayers}
        </span>
        <span>
          <strong>Stages:</strong> {game.totalStages}
        </span>
      </div>
      {game.startedAt && (
        <p style={{ fontSize: 'small', marginTop: '0.5rem' }}>
          Started: {new Date(game.startedAt).toLocaleString()}
        </p>
      )}

      <div style={{ display: 'flex', gap: '0.5rem', margin: '1rem 0' }}>
        <button className="outline" style={{ width: 'auto' }} onClick={() => navigate(`/admin/clients/${client}/games/${id}/edit`)}>
          Edit Game
        </button>
        <button className="outline secondary" style={{ width: 'auto' }} onClick={() => navigate(`/admin/clients/${client}/games`)}>
          Back to Games
        </button>
      </div>

      {game.teams.length === 0 ? (
        <p>No teams yet.</p>
      ) : (
        game.teams.map((team) => (
          <article key={team.id} style={{ marginBottom: '1rem' }}>
            <header>
              <strong>{team.name}</strong>
              {team.guideName && <span> &mdash; Guide: {team.guideName}</span>}
              <span style={{ float: 'right' }}>
                {team.completedStages}/{game.totalStages} stages
              </span>
            </header>
            {team.players.length === 0 ? (
              <p style={{ margin: 0 }}>No players yet.</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>Player</th>
                    <th>Joined</th>
                  </tr>
                </thead>
                <tbody>
                  {team.players.map((p, i) => (
                    <tr key={i}>
                      <td>{p.name}</td>
                      <td>{new Date(p.joinedAt).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </article>
        ))
      )}
    </>
  )
}
