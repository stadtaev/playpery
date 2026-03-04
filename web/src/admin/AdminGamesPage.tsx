import { useState, useEffect } from 'react'
import { listGames, deleteGame } from './adminApi'
import type { GameSummary } from './adminTypes'
import { LoadingPage } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

const statusLabels: Record<string, string> = {
  draft: 'Draft',
  active: 'Active',
  paused: 'Paused',
  ended: 'Ended',
}

export function AdminGamesPage({ client }: { client: string }) {
  const [games, setGames] = useState<GameSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    listGames(client)
      .then(setGames)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [client])

  async function handleDelete(id: string, scenarioName: string) {
    if (!confirm(`Delete game "${scenarioName}"?`)) return
    try {
      await deleteGame(client, id)
      setGames((prev) => prev.filter((g) => g.id !== id))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  if (loading) {
    return <LoadingPage message="Loading games..." />
  }

  if (error) {
    return <ErrorMessage message={error} />
  }

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h2 className="m-0">Games</h2>
        <button onClick={() => navigate(`/admin/clients/${client}/games/new`)} className="btn">
          New Game
        </button>
      </div>

      {games.length === 0 ? (
        <p className="text-secondary">No games yet.</p>
      ) : (
        <table className="admin-table">
          <thead>
            <tr>
              <th>Scenario</th>
              <th>Mode</th>
              <th>Status</th>
              <th>Timer</th>
              <th>Teams</th>
              <th>Created</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {games.map((g) => (
              <tr key={g.id}>
                <td>
                  <a
                    href={`/admin/clients/${client}/games/${g.id}/edit`}
                    onClick={(e) => { e.preventDefault(); navigate(`/admin/clients/${client}/games/${g.id}/edit`) }}
                  >
                    {g.scenarioName}{g.supervised && ' (supervised)'}
                  </a>
                </td>
                <td>{g.mode || 'classic'}</td>
                <td>{statusLabels[g.status] || g.status}</td>
                <td>{g.timerMinutes}m</td>
                <td>{g.teamCount}</td>
                <td>{new Date(g.createdAt).toLocaleDateString()}</td>
                <td>
                  <button
                    className="btn-danger btn-sm"
                    onClick={() => handleDelete(g.id, g.scenarioName)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  )
}
