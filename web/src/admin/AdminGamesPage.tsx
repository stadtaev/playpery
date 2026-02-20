import { useState, useEffect } from 'react'
import { listGames, deleteGame } from './adminApi'
import type { GameSummary } from './adminTypes'

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

export function AdminGamesPage() {
  const [games, setGames] = useState<GameSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    listGames()
      .then(setGames)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  async function handleDelete(id: string, scenarioName: string) {
    if (!confirm(`Delete game "${scenarioName}"?`)) return
    try {
      await deleteGame(id)
      setGames((prev) => prev.filter((g) => g.id !== id))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  if (loading) {
    return <p aria-busy="true">Loading games...</p>
  }

  if (error) {
    return <p role="alert">{error}</p>
  }

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h2 style={{ margin: 0 }}>Games</h2>
        <button onClick={() => navigate('/admin/games/new')} style={{ width: 'auto' }}>
          New Game
        </button>
      </div>

      {games.length === 0 ? (
        <p>No games yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Scenario</th>
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
                    href={`/admin/games/${g.id}/edit`}
                    onClick={(e) => { e.preventDefault(); navigate(`/admin/games/${g.id}/edit`) }}
                  >
                    {g.scenarioName}
                  </a>
                </td>
                <td>{statusLabels[g.status] || g.status}</td>
                <td>{g.timerMinutes}m</td>
                <td>{g.teamCount}</td>
                <td>{new Date(g.createdAt).toLocaleDateString()}</td>
                <td>
                  <button
                    className="outline secondary"
                    onClick={() => handleDelete(g.id, g.scenarioName)}
                    style={{ width: 'auto', padding: '0.25rem 0.5rem', fontSize: 'small' }}
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
