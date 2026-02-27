import { useState, useEffect } from 'react'
import { listScenarios, getGame, createGame, updateGame, createTeam, updateTeam, deleteTeam } from './adminApi'
import type { ScenarioSummary, GameRequest, TeamItem, TeamRequest } from './adminTypes'

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
  const [teams, setTeams] = useState<TeamItem[]>([])
  const [loading, setLoading] = useState(!!id)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  // New team form state.
  const [newTeamName, setNewTeamName] = useState('')
  const [newTeamToken, setNewTeamToken] = useState('')
  const [newTeamGuide, setNewTeamGuide] = useState('')
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

    const data: TeamRequest = { name: newTeamName, joinToken: newTeamToken, guideName: newTeamGuide }
    try {
      const team = await createTeam(client, id, data)
      setTeams((prev) => [...prev, team])
      setNewTeamName('')
      setNewTeamToken('')
      setNewTeamGuide('')
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
      const updated = await updateTeam(client, id, team.id, { name: name.trim(), joinToken: team.joinToken, guideName: guideName.trim() })
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
    return <p aria-busy="true">Loading...</p>
  }

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: '1rem' }}>
        <h2 style={{ margin: 0 }}>{id ? 'Edit Game' : 'New Game'}</h2>
        {id && (
          <button className="outline" style={{ width: 'auto', padding: '0.25rem 0.75rem', fontSize: 'small' }} onClick={() => navigate(`/admin/clients/${client}/games/${id}/status`)}>
            View Status
          </button>
        )}
      </div>
      {error && <p role="alert" style={{ color: 'var(--pico-color-red-500)' }}>{error}</p>}
      {startedAt && (
        <p><small>Started: {new Date(startedAt).toLocaleString()}</small></p>
      )}
      <form onSubmit={handleSubmit}>
        <label>
          Scenario
          <select value={scenarioId} onChange={(e) => setScenarioId(e.target.value)} required disabled={!!id && status !== 'draft'}>
            {scenarios.map((s) => (
              <option key={s.id} value={s.id}>{s.name} ({s.city})</option>
            ))}
          </select>
          {!!id && status !== 'draft' && <small>Scenario cannot be changed after game is activated</small>}
        </label>
        <label>
          Status
          <select value={status} onChange={(e) => setStatus(e.target.value)}>
            {statuses.map((s) => (
              <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>
            ))}
          </select>
        </label>
        <label>
          <input type="checkbox" checked={supervised} onChange={(e) => setSupervised(e.target.checked)} />
          Supervised game
        </label>
        <label>
          <input type="checkbox" checked={timerEnabled} onChange={(e) => setTimerEnabled(e.target.checked)} />
          Enable timer
        </label>
        {timerEnabled && (
          <div className="grid">
            <label>
              Game timer (minutes)
              <input type="number" min="1" value={timerMinutes} onChange={(e) => setTimerMinutes(parseInt(e.target.value) || 120)} required />
            </label>
            <label>
              Stage timer (minutes)
              <input type="number" min="1" value={stageTimerMinutes} onChange={(e) => setStageTimerMinutes(parseInt(e.target.value) || 10)} required />
            </label>
          </div>
        )}

        <div style={{ display: 'flex', gap: '1rem' }}>
          <button type="submit" disabled={saving} aria-busy={saving}>
            {saving ? 'Saving...' : id ? 'Update Game' : 'Create Game'}
          </button>
          <button type="button" className="secondary" onClick={() => navigate(`/admin/clients/${client}/games`)}>
            Cancel
          </button>
        </div>
      </form>

      {id && (
        <>
          <hr />
          <h3>Teams</h3>

          {teams.length === 0 ? (
            <p>No teams yet.</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Join Link</th>
                  {supervised && <th>Supervisor Link</th>}
                  {selectedMode === 'math_puzzle' && <th>Team Secret</th>}
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
                      <a href={joinPath} target="_blank" rel="noopener noreferrer" style={{ fontSize: 'small', wordBreak: 'break-all' }}>
                        {joinUrl}
                      </a>
                    </td>
                    {supervised && (
                      <td>
                        {t.supervisorToken ? (() => {
                          const superPath = `/join/${client}/${t.supervisorToken}`
                          const superUrl = `${window.location.origin}${superPath}`
                          return (
                            <a href={superPath} target="_blank" rel="noopener noreferrer" style={{ fontSize: 'small', wordBreak: 'break-all' }}>
                              {superUrl}
                            </a>
                          )
                        })() : '-'}
                      </td>
                    )}
                    {selectedMode === 'math_puzzle' && <td>{t.teamSecret || '-'}</td>}
                    <td>{t.guideName || '-'}</td>
                    <td>{t.playerCount}</td>
                    <td style={{ whiteSpace: 'nowrap' }}>
                      <button
                        className="outline"
                        onClick={() => handleUpdateTeam(t)}
                        style={{ width: 'auto', padding: '0.25rem 0.5rem', fontSize: 'small', marginRight: '0.25rem' }}
                      >
                        Edit
                      </button>
                      <button
                        className="outline secondary"
                        onClick={() => handleDeleteTeam(t)}
                        disabled={t.playerCount > 0}
                        title={t.playerCount > 0 ? 'Cannot delete team with players' : ''}
                        style={{ width: 'auto', padding: '0.25rem 0.5rem', fontSize: 'small' }}
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
            <form onSubmit={handleAddTeam} style={{ marginTop: '0.5rem' }}>
              <div className="grid">
                <label>
                  Team Name
                  <input type="text" value={newTeamName} onChange={(e) => setNewTeamName(e.target.value)} required />
                </label>
                <label>
                  Join Token (optional)
                  <input type="text" value={newTeamToken} onChange={(e) => setNewTeamToken(e.target.value)} placeholder="Auto-generated if blank" />
                </label>
                <label>
                  Guide Name (optional)
                  <input type="text" value={newTeamGuide} onChange={(e) => setNewTeamGuide(e.target.value)} />
                </label>
              </div>
              <button type="submit" disabled={addingTeam} aria-busy={addingTeam} style={{ width: 'auto' }}>
                {addingTeam ? 'Adding...' : 'Add Team'}
              </button>
            </form>
          </details>
        </>
      )}
    </>
  )
}
