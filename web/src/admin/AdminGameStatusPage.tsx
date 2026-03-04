import { useState, useEffect } from 'react'
import { getGameStatus } from './adminApi'
import type { GameStatus } from './adminTypes'
import { LoadingPage } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

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

  if (error) return <ErrorMessage message={error} />
  if (!game) return <LoadingPage />

  const totalPlayers = game.teams.reduce((sum, t) => sum + t.players.length, 0)

  return (
    <>
      <div className="flex items-baseline gap-4 flex-wrap mb-2">
        <h2 className="m-0">{game.scenarioName}</h2>
        <span className="text-sm">
          <strong>Status:</strong> {game.status}
        </span>
        {game.timerEnabled && (
          <span className="text-sm">
            <strong>Timer:</strong> {game.timerMinutes}m (stage: {game.stageTimerMinutes}m)
          </span>
        )}
        <span className="text-sm">
          <strong>Players:</strong> {totalPlayers}
        </span>
        <span className="text-sm">
          <strong>Stages:</strong> {game.totalStages}
        </span>
      </div>
      {game.startedAt && (
        <p className="text-secondary text-xs mb-4">
          Started: {new Date(game.startedAt).toLocaleString()}
        </p>
      )}

      <div className="flex gap-2 mb-8">
        <button className="btn-secondary btn-sm" onClick={() => navigate(`/admin/clients/${client}/games/${id}/edit`)}>
          Edit Game
        </button>
        <button className="btn-ghost btn-sm" onClick={() => navigate(`/admin/clients/${client}/games`)}>
          Back to Games
        </button>
      </div>

      {game.teams.length === 0 ? (
        <p className="text-secondary">No teams yet.</p>
      ) : (
        <>
          <h3>Scoreboard</h3>
          <table className="admin-table mb-8">
            <thead>
              <tr>
                <th>Team</th>
                <th>Points</th>
                <th>Progress</th>
                <th>Players</th>
              </tr>
            </thead>
            <tbody>
              {[...game.teams]
                .sort((a, b) => b.completedStages - a.completedStages)
                .map((team) => (
                  <tr key={team.id}>
                    <td><strong>{team.name}</strong></td>
                    <td>{team.completedStages}</td>
                    <td>{team.completedStages}/{game.totalStages} stages</td>
                    <td>{team.players.length}</td>
                  </tr>
                ))}
            </tbody>
          </table>

          <h3>Team Details</h3>
          {game.teams.map((team) => (
            <div key={team.id} className="card">
              <div className="flex justify-between items-center mb-3">
                <div>
                  <strong>{team.name}</strong>
                  {team.guideName && <span className="text-secondary"> &mdash; Guide: {team.guideName}</span>}
                </div>
                <span className="font-bold">{team.completedStages} pts</span>
              </div>
              {team.players.length === 0 ? (
                <p className="text-secondary text-sm m-0">No players yet.</p>
              ) : (
                <table className="admin-table">
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
            </div>
          ))}
        </>
      )}
    </>
  )
}
