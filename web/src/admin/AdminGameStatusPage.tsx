import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getGameStatus } from './adminApi'
import type { GameStatus } from './adminTypes'
import { LoadingPage } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminGameStatusPage({ client, id }: { client: string; id: string }) {
  const { t } = useTranslation('admin')
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
          {t('status_label')} {t(`status_${game.status}`)}
        </span>
        {game.timerEnabled && (
          <span className="text-sm">
            {t('timer_label', { game: game.timerMinutes, stage: game.stageTimerMinutes })}
          </span>
        )}
        <span className="text-sm">
          {t('players_label', { count: totalPlayers })}
        </span>
        <span className="text-sm">
          {t('stages_label', { count: game.totalStages })}
        </span>
      </div>
      {game.startedAt && (
        <p className="text-secondary text-xs mb-4">
          {t('started_label', { date: new Date(game.startedAt).toLocaleString() })}
        </p>
      )}

      <div className="flex gap-2 mb-8">
        <button className="btn-secondary btn-sm" onClick={() => navigate(`/admin/clients/${client}/games/${id}/edit`)}>
          {t('status_edit_game')}
        </button>
        <button className="btn-ghost btn-sm" onClick={() => navigate(`/admin/clients/${client}/games`)}>
          {t('status_back_to_games')}
        </button>
      </div>

      {game.teams.length === 0 ? (
        <p className="text-secondary">{t('teams_empty')}</p>
      ) : (
        <>
          <h3>{t('scoreboard_title')}</h3>
          <table className="admin-table mb-8">
            <thead>
              <tr>
                <th>{t('scoreboard_col_team')}</th>
                <th>{t('scoreboard_col_points')}</th>
                <th>{t('scoreboard_col_progress')}</th>
                <th>{t('scoreboard_col_players')}</th>
              </tr>
            </thead>
            <tbody>
              {[...game.teams]
                .sort((a, b) => b.completedStages - a.completedStages)
                .map((team) => (
                  <tr key={team.id}>
                    <td><strong>{team.name}</strong></td>
                    <td>{team.completedStages}</td>
                    <td>{t('scoreboard_progress', { completed: team.completedStages, total: game.totalStages })}</td>
                    <td>{team.players.length}</td>
                  </tr>
                ))}
            </tbody>
          </table>

          <h3>{t('team_details_title')}</h3>
          {game.teams.map((team) => (
            <div key={team.id} className="card">
              <div className="flex justify-between items-center mb-3">
                <div>
                  <strong>{team.name}</strong>
                  {team.guideName && <span className="text-secondary"> &mdash; {t('team_guide', { name: team.guideName })}</span>}
                </div>
                <span className="font-bold">{t('scoreboard_pts', { count: team.completedStages })}</span>
              </div>
              {team.players.length === 0 ? (
                <p className="text-secondary text-sm m-0">{t('team_no_players')}</p>
              ) : (
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>{t('team_col_player')}</th>
                      <th>{t('team_col_joined')}</th>
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
