import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { listScenarios, getGame, createGame, updateGame, createTeam, updateTeam, deleteTeam } from './adminApi'
import type { ScenarioSummary, Stage, GameRequest, TeamItem, TeamRequest } from './adminTypes'
import { LoadingPage, Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

const statuses = ['draft', 'active', 'paused', 'ended'] as const

export function AdminGameEditorPage({ client, id }: { client: string; id?: string }) {
  const { t } = useTranslation('admin')
  const [scenarios, setScenarios] = useState<ScenarioSummary[]>([])
  const [scenarioId, setScenarioId] = useState('')
  const [status, setStatus] = useState('draft')
  const [language, setLanguage] = useState('ru')
  const [supervised, setSupervised] = useState(true)
  const [timerEnabled, setTimerEnabled] = useState(false)
  const [timerMinutes, setTimerMinutes] = useState(120)
  const [stageTimerMinutes, setStageTimerMinutes] = useState(10)
  const [notes, setNotes] = useState('')
  const [startedAt, setStartedAt] = useState<string | null>(null)
  const [stages, setStages] = useState<Stage[]>([])
  const [teams, setTeams] = useState<TeamItem[]>([])
  const [loading, setLoading] = useState(!!id)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  // New team form state.
  const [newTeamName, setNewTeamName] = useState('')
  const [newTeamToken, setNewTeamToken] = useState('')
  const [newTeamGuide, setNewTeamGuide] = useState('')
  const [newTeamStartStage, setNewTeamStartStage] = useState(0)
  const [addingTeam, setAddingTeam] = useState(false)

  useEffect(() => {
    const loads: Promise<void>[] = [
      listScenarios().then((s) => {
        setScenarios(s)
        if (!id && s.length > 0) setScenarioId(s[0].id)
      }),
    ]

    if (id) {
      loads.push(
        getGame(client, id).then((g) => {
          setScenarioId(g.scenarioId)
          setStatus(g.status)
          setLanguage(g.language || 'ru')
          setSupervised(g.supervised)
          setTimerEnabled(g.timerEnabled)
          setTimerMinutes(g.timerMinutes || 120)
          setStageTimerMinutes(g.stageTimerMinutes || 10)
          setNotes(g.notes || '')
          setStartedAt(g.startedAt)
          setStages(g.stages || [])
          setTeams(g.teams)
        })
      )
    }

    Promise.all(loads)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [client, id])

  const selectedMode = scenarios.find((s) => s.id === scenarioId)?.mode || 'classic'

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')

    const data: GameRequest = { scenarioId, language, status, supervised, timerEnabled, timerMinutes, stageTimerMinutes, notes }

    try {
      if (id) {
        const updated = await updateGame(client, id, data)
        setStartedAt(updated.startedAt)
        setStages(updated.stages || [])
        setTeams(updated.teams)
      } else {
        const created = await createGame(client, data)
        setSaving(false)
        navigate(`/admin/clients/${client}/games/${created.id}/edit`)
        return
      }
      navigate(`/admin/clients/${client}/games`)
    } catch (e) {
      setError(e instanceof Error ? e.message : t('game_save_failed'))
      setSaving(false)
    }
  }

  async function handleAddTeam(e: React.FormEvent) {
    e.preventDefault()
    if (!id) return
    setAddingTeam(true)
    setError('')

    const data: TeamRequest = { name: newTeamName, joinToken: newTeamToken, guideName: newTeamGuide, startStage: newTeamStartStage }
    try {
      const team = await createTeam(client, id, data)
      setTeams((prev) => [...prev, team])
      setNewTeamName('')
      setNewTeamToken('')
      setNewTeamGuide('')
      setNewTeamStartStage(0)
    } catch (e) {
      setError(e instanceof Error ? e.message : t('game_add_team_failed'))
    } finally {
      setAddingTeam(false)
    }
  }

  async function handleUpdateTeam(team: TeamItem) {
    if (!id) return
    const name = prompt(t('teams_prompt_name'), team.name)
    if (name === null || name.trim() === '') return
    const guideName = prompt(t('teams_prompt_guide'), team.guideName) ?? ''

    try {
      const updated = await updateTeam(client, id, team.id, { name: name.trim(), joinToken: team.joinToken, guideName: guideName.trim(), startStage: team.startStage })
      setTeams((prev) => prev.map((t) => (t.id === team.id ? updated : t)))
    } catch (e) {
      alert(e instanceof Error ? e.message : t('game_update_failed'))
    }
  }

  async function handleDeleteTeam(team: TeamItem) {
    if (!id) return
    if (!confirm(t('teams_delete_confirm', { name: team.name }))) return
    try {
      await deleteTeam(client, id, team.id)
      setTeams((prev) => prev.filter((t) => t.id !== team.id))
    } catch (e) {
      alert(e instanceof Error ? e.message : t('game_delete_failed'))
    }
  }

  if (loading) {
    return <LoadingPage />
  }

  return (
    <>
      <div className="flex items-baseline gap-4 mb-4">
        <h2 className="m-0">{id ? t('game_edit_title') : t('game_new_title')}</h2>
        {id && (
          <button className="btn-ghost btn-sm" onClick={() => navigate(`/admin/clients/${client}/games/${id}/status`)}>
            {t('game_view_status')}
          </button>
        )}
      </div>
      {error && <ErrorMessage message={error} />}
      {startedAt && (
        <p className="text-secondary text-sm mb-4">{t('game_started', { date: new Date(startedAt).toLocaleString() })}</p>
      )}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="input-label">{t('game_scenario')}</label>
          <select className="input" value={scenarioId} onChange={(e) => setScenarioId(e.target.value)} required disabled={!!id && status !== 'draft'}>
            {scenarios.map((s) => (
              <option key={s.id} value={s.id}>{s.name} ({s.city})</option>
            ))}
          </select>
          {!!id && status !== 'draft' && <p className="text-secondary text-xs mt-1">{t('game_scenario_locked')}</p>}
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="input-label">{t('game_status')}</label>
            <select className="input" value={status} onChange={(e) => setStatus(e.target.value)}>
              {statuses.map((s) => (
                <option key={s} value={s}>{t(`status_${s}`)}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="input-label">{t('game_language')}</label>
            <select className="input" value={language} onChange={(e) => setLanguage(e.target.value)}>
              <option value="en">{t('lang_en')}</option>
              <option value="ru">{t('lang_ru')}</option>
            </select>
          </div>
        </div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={supervised} onChange={(e) => setSupervised(e.target.checked)} />
          <span className="text-sm">{t('game_supervised')}</span>
        </label>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={timerEnabled} onChange={(e) => setTimerEnabled(e.target.checked)} />
          <span className="text-sm">{t('game_timer_enable')}</span>
        </label>
        {timerEnabled && (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="input-label">{t('game_timer_minutes')}</label>
              <input className="input" type="number" min="1" value={timerMinutes} onChange={(e) => setTimerMinutes(parseInt(e.target.value) || 120)} required />
            </div>
            <div>
              <label className="input-label">{t('game_stage_timer_minutes')}</label>
              <input className="input" type="number" min="1" value={stageTimerMinutes} onChange={(e) => setStageTimerMinutes(parseInt(e.target.value) || 10)} required />
            </div>
          </div>
        )}

        <div>
          <label className="input-label">{t('game_notes')}</label>
          <textarea className="input" rows={3} value={notes} onChange={(e) => setNotes(e.target.value)} placeholder={t('game_notes_placeholder')} />
        </div>

        <div className="flex gap-4">
          <button type="submit" disabled={saving} className="btn">
            {saving ? <Spinner /> : id ? t('game_update') : t('game_create')}
          </button>
          <button type="button" className="btn-secondary" onClick={() => navigate(`/admin/clients/${client}/games`)}>
            {t('game_cancel')}
          </button>
        </div>
      </form>

      {id && (
        <>
          <hr />
          <h3>{t('teams_title')}</h3>

          {teams.length === 0 ? (
            <p className="text-secondary">{t('teams_empty')}</p>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th>{t('teams_col_name')}</th>
                  <th>{t('teams_col_join_link')}</th>
                  {supervised && <th>{t('teams_col_supervisor_link')}</th>}
                  {selectedMode === 'math_puzzle' && <th>{t('teams_col_team_secret')}</th>}
                  <th>{t('teams_col_start_stage')}</th>
                  <th>{t('teams_col_guide')}</th>
                  <th>{t('teams_col_players')}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {teams.map((tm) => {
                  const joinPath = `/join/${client}/${tm.joinToken}`
                  const joinUrl = `${window.location.origin}${joinPath}`
                  return (
                  <tr key={tm.id}>
                    <td>{tm.name}</td>
                    <td>
                      <a href={joinPath} target="_blank" rel="noopener noreferrer" className="text-xs break-all">
                        {joinUrl}
                      </a>
                    </td>
                    {supervised && (
                      <td>
                        {tm.supervisorToken ? (() => {
                          const superPath = `/join/${client}/${tm.supervisorToken}`
                          const superUrl = `${window.location.origin}${superPath}`
                          return (
                            <a href={superPath} target="_blank" rel="noopener noreferrer" className="text-xs break-all">
                              {superUrl}
                            </a>
                          )
                        })() : '-'}
                      </td>
                    )}
                    {selectedMode === 'math_puzzle' && <td>{tm.teamSecret || '-'}</td>}
                    <td>{tm.startStage ? t('teams_stage_location', { number: tm.startStage, location: stages.find(s => s.stageNumber === tm.startStage)?.location || '' }) : t('teams_default_stage')}</td>
                    <td>{tm.guideName || '-'}</td>
                    <td>{tm.playerCount}</td>
                    <td className="whitespace-nowrap">
                      <button
                        className="btn-ghost btn-sm mr-1"
                        onClick={() => handleUpdateTeam(tm)}
                      >
                        {t('teams_edit')}
                      </button>
                      <button
                        className="btn-danger btn-sm"
                        onClick={() => handleDeleteTeam(tm)}
                        disabled={tm.playerCount > 0}
                        title={tm.playerCount > 0 ? t('teams_delete_disabled') : ''}
                      >
                        {t('teams_delete')}
                      </button>
                    </td>
                  </tr>
                  )
                })}
              </tbody>
            </table>
          )}

          <details>
            <summary>{t('teams_add')}</summary>
            <form onSubmit={handleAddTeam} className="mt-4 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="input-label">{t('teams_team_name')}</label>
                  <input className="input" type="text" value={newTeamName} onChange={(e) => setNewTeamName(e.target.value)} required />
                </div>
                <div>
                  <label className="input-label">{t('teams_starting_stage')}</label>
                  <select className="input" value={newTeamStartStage} onChange={(e) => setNewTeamStartStage(parseInt(e.target.value))}>
                    <option value={0}>{t('teams_default_stage_option')}</option>
                    {stages.map((s) => (
                      <option key={s.stageNumber} value={s.stageNumber}>{t('teams_stage_option', { number: s.stageNumber, location: s.location })}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="input-label">{t('teams_join_token')}</label>
                  <input className="input" type="text" value={newTeamToken} onChange={(e) => setNewTeamToken(e.target.value)} placeholder={t('teams_join_token_placeholder')} />
                </div>
                <div>
                  <label className="input-label">{t('teams_guide_name')}</label>
                  <input className="input" type="text" value={newTeamGuide} onChange={(e) => setNewTeamGuide(e.target.value)} />
                </div>
              </div>
              <button type="submit" disabled={addingTeam} className="btn">
                {addingTeam ? <Spinner /> : t('teams_add')}
              </button>
            </form>
          </details>
        </>
      )}
    </>
  )
}
