import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { listGames, deleteGame } from './adminApi'
import type { GameSummary } from './adminTypes'
import { LoadingPage } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminGamesPage({ client }: { client: string }) {
  const { t } = useTranslation('admin')
  const [games, setGames] = useState<GameSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    listGames(client)
      .then((g) => g.sort((a, b) => b.createdAt.localeCompare(a.createdAt)))
      .then(setGames)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [client])

  async function handleDelete(id: string, scenarioName: string) {
    if (!confirm(t('games_delete_confirm', { name: scenarioName }))) return
    try {
      await deleteGame(client, id)
      setGames((prev) => prev.filter((g) => g.id !== id))
    } catch (e) {
      alert(e instanceof Error ? e.message : t('games_delete_failed'))
    }
  }

  if (loading) {
    return <LoadingPage message={t('games_loading')} />
  }

  if (error) {
    return <ErrorMessage message={error} />
  }

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h2 className="m-0">{t('games_title')}</h2>
        <button onClick={() => navigate(`/admin/clients/${client}/games/new`)} className="btn">
          {t('games_new')}
        </button>
      </div>

      {games.length === 0 ? (
        <p className="text-secondary">{t('games_empty')}</p>
      ) : (
        <table className="admin-table">
          <thead>
            <tr>
              <th>{t('games_col_scenario')}</th>
              <th>{t('games_col_mode')}</th>
              <th>{t('games_col_status')}</th>
              <th>{t('games_col_timer')}</th>
              <th>{t('games_col_teams')}</th>
              <th>{t('games_col_created')}</th>
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
                    {g.scenarioName}{g.supervised && ` ${t('games_supervised_suffix')}`}
                  </a>
                </td>
                <td>{g.mode || 'classic'}</td>
                <td>{t(`status_${g.status}`)}</td>
                <td>{t('games_timer_minutes', { minutes: g.timerMinutes })}</td>
                <td>{g.teamCount}</td>
                <td>{new Date(g.createdAt).toLocaleDateString()}</td>
                <td>
                  <button
                    className="btn-danger btn-sm"
                    onClick={() => handleDelete(g.id, g.scenarioName)}
                  >
                    {t('games_delete')}
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
