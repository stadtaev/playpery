import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Pencil, Trash2, Copy, Check, Plus, ExternalLink, ArrowLeft, Eye } from 'lucide-react'
import { listScenarios, getGame, createGame, updateGame, createTeam, updateTeam, deleteTeam } from './adminApi'
import type { ScenarioSummary, GameRequest, TeamItem, TeamRequest } from './adminTypes'
import { navigate } from '@/lib/navigate'
import { Button, MotionButton } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

const statuses = ['draft', 'active', 'paused', 'ended'] as const

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  function handleCopy() {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={handleCopy}
      className="h-7 w-7 text-text-muted hover:text-accent"
      title="Copy link"
    >
      {copied ? (
        <motion.span initial={{ scale: 0.5 }} animate={{ scale: 1 }}>
          <Check size={14} className="text-success" />
        </motion.span>
      ) : (
        <Copy size={14} />
      )}
    </Button>
  )
}

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

  // New team form state
  const [newTeamName, setNewTeamName] = useState('')
  const [newTeamToken, setNewTeamToken] = useState('')
  const [newTeamGuide, setNewTeamGuide] = useState('')
  const [addingTeam, setAddingTeam] = useState(false)

  // Inline edit state
  const [editingTeamId, setEditingTeamId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [editGuide, setEditGuide] = useState('')
  const [editSaving, setEditSaving] = useState(false)

  // Delete dialog state
  const [deleteTarget, setDeleteTarget] = useState<TeamItem | null>(null)
  const [deleting, setDeleting] = useState(false)

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

  async function handleAddTeam() {
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

  function startEditTeam(team: TeamItem) {
    setEditingTeamId(team.id)
    setEditName(team.name)
    setEditGuide(team.guideName)
  }

  function cancelEdit() {
    setEditingTeamId(null)
    setEditName('')
    setEditGuide('')
  }

  async function saveEditTeam(team: TeamItem) {
    if (!id) return
    setEditSaving(true)
    try {
      const updated = await updateTeam(client, id, team.id, {
        name: editName.trim(),
        joinToken: team.joinToken,
        guideName: editGuide.trim(),
      })
      setTeams((prev) => prev.map((t) => (t.id === team.id ? updated : t)))
      setEditingTeamId(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Update failed')
    } finally {
      setEditSaving(false)
    }
  }

  async function handleDeleteTeam() {
    if (!id || !deleteTarget) return
    setDeleting(true)
    try {
      await deleteTeam(client, id, deleteTarget.id)
      setTeams((prev) => prev.filter((t) => t.id !== deleteTarget.id))
      setDeleteTarget(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Delete failed')
      setDeleteTarget(null)
    } finally {
      setDeleting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Spinner size={32} />
      </div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <div className="flex items-center gap-3 mb-6">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate(`/admin/clients/${client}/games`)}
          className="text-text-muted hover:text-text-primary"
        >
          <ArrowLeft size={18} />
        </Button>
        <h2 className="text-xl font-semibold text-text-primary">
          {id ? 'Edit Game' : 'New Game'}
        </h2>
        {id && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/admin/clients/${client}/games/${id}/status`)}
          >
            <Eye size={14} />
            View Status
          </Button>
        )}
        {startedAt && (
          <span className="ml-auto text-xs text-text-muted">
            Started {new Date(startedAt).toLocaleString()}
          </span>
        )}
      </div>

      {error && (
        <div className="mb-4">
          <Alert variant="error">{error}</Alert>
        </div>
      )}

      <Card className="mb-6">
        <CardHeader>
          <h3 className="text-sm font-medium text-text-secondary uppercase tracking-wider">Game Settings</h3>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div className="space-y-1.5">
                <Label>Scenario</Label>
                <Select
                  value={scenarioId}
                  onChange={(e) => setScenarioId(e.target.value)}
                  required
                  disabled={!!id && status !== 'draft'}
                >
                  {scenarios.map((s) => (
                    <option key={s.id} value={s.id}>{s.name} ({s.city})</option>
                  ))}
                </Select>
                {!!id && status !== 'draft' && (
                  <p className="text-xs text-text-muted">Cannot change after activation</p>
                )}
              </div>
              <div className="space-y-1.5">
                <Label>Status</Label>
                <Select value={status} onChange={(e) => setStatus(e.target.value)}>
                  {statuses.map((s) => (
                    <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>
                  ))}
                </Select>
              </div>
            </div>

            <div className="flex flex-wrap gap-6 mb-4">
              <div className="flex items-center gap-2">
                <Switch checked={supervised} onCheckedChange={setSupervised} />
                <Label className="cursor-pointer" onClick={() => setSupervised(!supervised)}>
                  Supervised game
                </Label>
              </div>
              <div className="flex items-center gap-2">
                <Switch checked={timerEnabled} onCheckedChange={setTimerEnabled} />
                <Label className="cursor-pointer" onClick={() => setTimerEnabled(!timerEnabled)}>
                  Enable timer
                </Label>
              </div>
            </div>

            {timerEnabled && (
              <motion.div
                className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4"
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
              >
                <div className="space-y-1.5">
                  <Label>Game timer (minutes)</Label>
                  <Input
                    type="number"
                    min="1"
                    value={timerMinutes}
                    onChange={(e) => setTimerMinutes(parseInt(e.target.value) || 120)}
                    required
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Stage timer (minutes)</Label>
                  <Input
                    type="number"
                    min="1"
                    value={stageTimerMinutes}
                    onChange={(e) => setStageTimerMinutes(parseInt(e.target.value) || 10)}
                    required
                  />
                </div>
              </motion.div>
            )}

            <div className="flex gap-3">
              <MotionButton type="submit" disabled={saving}>
                {saving ? (
                  <>
                    <Spinner size={16} />
                    Saving...
                  </>
                ) : id ? (
                  'Update Game'
                ) : (
                  'Create Game'
                )}
              </MotionButton>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate(`/admin/clients/${client}/games`)}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {id && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium text-text-secondary uppercase tracking-wider">
                Teams
                {teams.length > 0 && (
                  <Badge variant="default" className="ml-2 normal-case">{teams.length}</Badge>
                )}
              </h3>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Join Link</TableHead>
                  {supervised && <TableHead>Supervisor Link</TableHead>}
                  {selectedMode === 'math_puzzle' && <TableHead>Secret</TableHead>}
                  <TableHead>Guide</TableHead>
                  <TableHead className="text-center">Players</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {teams.map((t, i) => {
                  const isEditing = editingTeamId === t.id
                  const joinPath = `/join/${client}/${t.joinToken}`
                  const joinUrl = `${window.location.origin}${joinPath}`

                  return (
                    <motion.tr
                      key={t.id}
                      className="border-b border-border transition-colors hover:bg-popover"
                      initial={{ opacity: 0, y: 8 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: i * 0.04, duration: 0.25 }}
                    >
                      <TableCell>
                        {isEditing ? (
                          <Input
                            value={editName}
                            onChange={(e) => setEditName(e.target.value)}
                            className="h-8 w-36"
                            autoFocus
                          />
                        ) : (
                          <span className="font-medium">{t.name}</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <a
                            href={joinPath}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs text-accent hover:underline truncate max-w-[160px]"
                            title={joinUrl}
                          >
                            <ExternalLink size={12} className="inline mr-1" />
                            {t.joinToken}
                          </a>
                          <CopyButton text={joinUrl} />
                        </div>
                      </TableCell>
                      {supervised && (
                        <TableCell>
                          {t.supervisorToken ? (() => {
                            const superPath = `/join/${client}/${t.supervisorToken}`
                            const superUrl = `${window.location.origin}${superPath}`
                            return (
                              <div className="flex items-center gap-1">
                                <a
                                  href={superPath}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="text-xs text-accent hover:underline truncate max-w-[160px]"
                                  title={superUrl}
                                >
                                  <ExternalLink size={12} className="inline mr-1" />
                                  {t.supervisorToken}
                                </a>
                                <CopyButton text={superUrl} />
                              </div>
                            )
                          })() : (
                            <span className="text-text-muted">-</span>
                          )}
                        </TableCell>
                      )}
                      {selectedMode === 'math_puzzle' && (
                        <TableCell className="text-text-secondary">
                          {t.teamSecret ?? '-'}
                        </TableCell>
                      )}
                      <TableCell>
                        {isEditing ? (
                          <Input
                            value={editGuide}
                            onChange={(e) => setEditGuide(e.target.value)}
                            className="h-8 w-28"
                            placeholder="Guide name"
                          />
                        ) : (
                          <span className="text-text-secondary">{t.guideName || '-'}</span>
                        )}
                      </TableCell>
                      <TableCell className="text-center">
                        <Badge variant={t.playerCount > 0 ? 'info' : 'default'}>
                          {t.playerCount}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          {isEditing ? (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => saveEditTeam(t)}
                                disabled={editSaving || !editName.trim()}
                                className="h-7 text-xs text-success hover:text-success"
                              >
                                {editSaving ? <Spinner size={14} /> : <Check size={14} />}
                                Save
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={cancelEdit}
                                className="h-7 text-xs"
                              >
                                Cancel
                              </Button>
                            </>
                          ) : (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => startEditTeam(t)}
                                className="h-7 w-7 text-text-muted hover:text-accent"
                                title="Edit team"
                              >
                                <Pencil size={14} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => setDeleteTarget(t)}
                                disabled={t.playerCount > 0}
                                className="h-7 w-7 text-text-muted hover:text-error disabled:opacity-30"
                                title={t.playerCount > 0 ? 'Cannot delete team with players' : 'Delete team'}
                              >
                                <Trash2 size={14} />
                              </Button>
                            </>
                          )}
                        </div>
                      </TableCell>
                    </motion.tr>
                  )
                })}

                {/* Add team inline row */}
                <TableRow className="hover:bg-transparent">
                  <TableCell>
                    <Input
                      value={newTeamName}
                      onChange={(e) => setNewTeamName(e.target.value)}
                      onKeyDown={(e) => { if (e.key === 'Enter' && newTeamName.trim()) handleAddTeam() }}
                      placeholder="Team name"
                      className="h-8 w-36"
                    />
                  </TableCell>
                  <TableCell>
                    <Input
                      value={newTeamToken}
                      onChange={(e) => setNewTeamToken(e.target.value)}
                      onKeyDown={(e) => { if (e.key === 'Enter' && newTeamName.trim()) handleAddTeam() }}
                      placeholder="Auto-generated"
                      className="h-8 w-32"
                    />
                  </TableCell>
                  {supervised && <TableCell />}
                  {selectedMode === 'math_puzzle' && <TableCell />}
                  <TableCell>
                    <Input
                      value={newTeamGuide}
                      onChange={(e) => setNewTeamGuide(e.target.value)}
                      onKeyDown={(e) => { if (e.key === 'Enter' && newTeamName.trim()) handleAddTeam() }}
                      placeholder="Guide name"
                      className="h-8 w-28"
                    />
                  </TableCell>
                  <TableCell />
                  <TableCell className="text-right">
                    <MotionButton
                      size="sm"
                      onClick={() => handleAddTeam()}
                      disabled={addingTeam || !newTeamName.trim()}
                      className="h-7 text-xs"
                    >
                      {addingTeam ? (
                        <Spinner size={14} />
                      ) : (
                        <Plus size={14} />
                      )}
                      Add
                    </MotionButton>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Delete team confirmation dialog */}
      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogHeader>
          <DialogTitle>Delete Team</DialogTitle>
        </DialogHeader>
        <p className="text-text-secondary mb-4">
          Are you sure you want to delete team "{deleteTarget?.name}"? This cannot be undone.
        </p>
        <div className="flex gap-3 justify-end">
          <Button variant="outline" onClick={() => setDeleteTarget(null)}>
            Cancel
          </Button>
          <MotionButton variant="destructive" onClick={handleDeleteTeam} disabled={deleting}>
            {deleting ? (
              <>
                <Spinner size={16} />
                Deleting...
              </>
            ) : (
              <>
                <Trash2 size={16} />
                Delete
              </>
            )}
          </MotionButton>
        </div>
      </Dialog>
    </motion.div>
  )
}
