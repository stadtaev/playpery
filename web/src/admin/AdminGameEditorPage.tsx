import { useState, useEffect } from 'react'
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
  const [scenarios, setScenarios] = useState<ScenarioSummary[]>([])
  const [scenarioId, setScenarioId] = useState('')
  const [status, setStatus] = useState('draft')
  const [supervised, setSupervised] = useState(false)
  const [timerEnabled, setTimerEnabled] = useState(false)
  const [timerMinutes, setTimerMinutes] = useState(120)
  const [stageTimerMinutes, setStageTimerMinutes] = useState(10)
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
          setSupervised(g.supervised)
          setTimerEnabled(g.timerEnabled)
          setTimerMinutes(g.timerMinutes || 120)
          setStageTimerMinutes(g.stageTimerMinutes || 10)
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

    const data: GameRequest = { scenarioId, status, supervised, timerEnabled, timerMinutes, stageTimerMinutes }

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
      setError(e instanceof Error ? e.message : 'Save failed')
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
      setError(e instanceof Error ? e.message : 'Failed to add team')
    } finally {
      setAddingTeam(false)
    }
  }

  async function handleUpdateTeam(team: TeamItem) {
    if (!id) return
    const name = prompt('Team name:', team.name)
    if (name === null || name.trim() === '') return
    const guideName = prompt('Guide name:', team.guideName) ?? ''

    try {
      const updated = await updateTeam(client, id, team.id, { name: name.trim(), joinToken: team.joinToken, guideName: guideName.trim(), startStage: team.startStage })
      setTeams((prev) => prev.map((t) => (t.id === team.id ? updated : t)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Update failed')
    }
  }

  async function handleDeleteTeam(team: TeamItem) {
    if (!id) return
    if (!confirm(`Delete team "${team.name}"?`)) return
    try {
      await deleteTeam(client, id, team.id)
      setTeams((prev) => prev.filter((t) => t.id !== team.id))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  if (loading) {
    return <LoadingPage />
  }

  return (
    <>
      <div className="flex items-baseline gap-4 mb-4">
        <h2 className="m-0">{id ? 'Edit Game' : 'New Game'}</h2>
        {id && (
          <button className="btn-ghost btn-sm" onClick={() => navigate(`/admin/clients/${client}/games/${id}/status`)}>
            View Status
          </button>
        )}
      </div>
      {error && <ErrorMessage message={error} />}
      {startedAt && (
        <p className="text-secondary text-sm mb-4">Started: {new Date(startedAt).toLocaleString()}</p>
      )}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="input-label">Scenario</label>
          <select className="input" value={scenarioId} onChange={(e) => setScenarioId(e.target.value)} required disabled={!!id && status !== 'draft'}>
            {scenarios.map((s) => (
              <option key={s.id} value={s.id}>{s.name} ({s.city})</option>
            ))}
          </select>
          {!!id && status !== 'draft' && <p className="text-secondary text-xs mt-1">Scenario cannot be changed after game is activated</p>}
        </div>
        <div>
          <label className="input-label">Status</label>
          <select className="input" value={status} onChange={(e) => setStatus(e.target.value)}>
            {statuses.map((s) => (
              <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>
            ))}
          </select>
        </div>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={supervised} onChange={(e) => setSupervised(e.target.checked)} />
          <span className="text-sm">Supervised game</span>
        </label>
        <label className="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" checked={timerEnabled} onChange={(e) => setTimerEnabled(e.target.checked)} />
          <span className="text-sm">Enable timer</span>
        </label>
        {timerEnabled && (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="input-label">Game timer (minutes)</label>
              <input className="input" type="number" min="1" value={timerMinutes} onChange={(e) => setTimerMinutes(parseInt(e.target.value) || 120)} required />
            </div>
            <div>
              <label className="input-label">Stage timer (minutes)</label>
              <input className="input" type="number" min="1" value={stageTimerMinutes} onChange={(e) => setStageTimerMinutes(parseInt(e.target.value) || 10)} required />
            </div>
          </div>
        )}

        <div className="flex gap-4">
          <button type="submit" disabled={saving} className="btn">
            {saving ? <Spinner /> : id ? 'Update Game' : 'Create Game'}
          </button>
          <button type="button" className="btn-secondary" onClick={() => navigate(`/admin/clients/${client}/games`)}>
            Cancel
          </button>
        </div>
      </form>

      {id && (
        <>
          <hr />
          <h3>Teams</h3>

          {teams.length === 0 ? (
            <p className="text-secondary">No teams yet.</p>
          ) : (
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Join Link</th>
                  {supervised && <th>Supervisor Link</th>}
                  {selectedMode === 'math_puzzle' && <th>Team Secret</th>}
                  <th>Start Stage</th>
                  <th>Guide</th>
                  <th>Players</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {teams.map((t) => {
                  const joinPath = `/join/${client}/${t.joinToken}`
                  const joinUrl = `${window.location.origin}${joinPath}`
                  return (
                  <tr key={t.id}>
                    <td>{t.name}</td>
                    <td>
                      <a href={joinPath} target="_blank" rel="noopener noreferrer" className="text-xs break-all">
                        {joinUrl}
                      </a>
                    </td>
                    {supervised && (
                      <td>
                        {t.supervisorToken ? (() => {
                          const superPath = `/join/${client}/${t.supervisorToken}`
                          const superUrl = `${window.location.origin}${superPath}`
                          return (
                            <a href={superPath} target="_blank" rel="noopener noreferrer" className="text-xs break-all">
                              {superUrl}
                            </a>
                          )
                        })() : '-'}
                      </td>
                    )}
                    {selectedMode === 'math_puzzle' && <td>{t.teamSecret || '-'}</td>}
                    <td>{t.startStage ? `${t.startStage} — ${stages.find(s => s.stageNumber === t.startStage)?.location || ''}` : 'Default (1)'}</td>
                    <td>{t.guideName || '-'}</td>
                    <td>{t.playerCount}</td>
                    <td className="whitespace-nowrap">
                      <button
                        className="btn-ghost btn-sm mr-1"
                        onClick={() => handleUpdateTeam(t)}
                      >
                        Edit
                      </button>
                      <button
                        className="btn-danger btn-sm"
                        onClick={() => handleDeleteTeam(t)}
                        disabled={t.playerCount > 0}
                        title={t.playerCount > 0 ? 'Cannot delete team with players' : ''}
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                  )
                })}
              </tbody>
            </table>
          )}

          <details>
            <summary>Add Team</summary>
            <form onSubmit={handleAddTeam} className="mt-4 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="input-label">Team Name</label>
                  <input className="input" type="text" value={newTeamName} onChange={(e) => setNewTeamName(e.target.value)} required />
                </div>
                <div>
                  <label className="input-label">Starting Stage</label>
                  <select className="input" value={newTeamStartStage} onChange={(e) => setNewTeamStartStage(parseInt(e.target.value))}>
                    <option value={0}>Default (Stage 1)</option>
                    {stages.map((s) => (
                      <option key={s.stageNumber} value={s.stageNumber}>Stage {s.stageNumber} — {s.location}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="input-label">Join Token (optional)</label>
                  <input className="input" type="text" value={newTeamToken} onChange={(e) => setNewTeamToken(e.target.value)} placeholder="Auto-generated if blank" />
                </div>
                <div>
                  <label className="input-label">Guide Name (optional)</label>
                  <input className="input" type="text" value={newTeamGuide} onChange={(e) => setNewTeamGuide(e.target.value)} />
                </div>
              </div>
              <button type="submit" disabled={addingTeam} className="btn">
                {addingTeam ? <Spinner /> : 'Add Team'}
              </button>
            </form>
          </details>
        </>
      )}
    </>
  )
}
