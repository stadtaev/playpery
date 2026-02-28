import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Plus, Trash2, Users, Timer, Eye } from 'lucide-react'
import { listGames, deleteGame } from './adminApi'
import type { GameSummary } from './adminTypes'
import { navigate } from '@/lib/navigate'
import { Button, MotionButton } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

const statusBadgeVariant: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  draft: 'default',
  active: 'success',
  paused: 'warning',
  ended: 'error',
}

const modeBadgeVariant: Record<string, 'default' | 'success' | 'warning' | 'info' | 'error'> = {
  classic: 'default',
  qr_quiz: 'info',
  qr_hunt: 'success',
  math_puzzle: 'warning',
  guided: 'error',
}

export function AdminGamesPage({ client }: { client: string }) {
  const [games, setGames] = useState<GameSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<GameSummary | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    listGames(client)
      .then(setGames)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [client])

  async function handleDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteGame(client, deleteTarget.id)
      setGames((prev) => prev.filter((g) => g.id !== deleteTarget.id))
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
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-text-primary">Games</h2>
        <MotionButton onClick={() => navigate(`/admin/clients/${client}/games/new`)}>
          <Plus size={16} />
          New Game
        </MotionButton>
      </div>

      {error && (
        <div className="mb-4">
          <Alert variant="error">{error}</Alert>
        </div>
      )}

      {games.length === 0 ? (
        <p className="text-text-secondary">No games yet.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Scenario</TableHead>
              <TableHead>Mode</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-center">Timer</TableHead>
              <TableHead className="text-center">Teams</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {games.map((g, i) => {
              const mode = g.mode || 'classic'
              return (
                <motion.tr
                  key={g.id}
                  className="border-b border-border transition-colors hover:bg-popover"
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.04, duration: 0.25 }}
                >
                  <TableCell>
                    <button
                      className="text-accent hover:underline bg-transparent border-none cursor-pointer p-0 font-medium"
                      onClick={() => navigate(`/admin/clients/${client}/games/${g.id}/edit`)}
                    >
                      {g.scenarioName}
                    </button>
                    {g.supervised && (
                      <span className="ml-2 inline-flex items-center text-text-muted text-xs">
                        <Eye size={12} className="mr-0.5" />
                        supervised
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <Badge variant={modeBadgeVariant[mode] ?? 'default'}>{mode}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={statusBadgeVariant[g.status] ?? 'default'}>{g.status}</Badge>
                  </TableCell>
                  <TableCell className="text-center text-text-secondary">
                    {g.timerEnabled ? (
                      <span className="inline-flex items-center gap-1">
                        <Timer size={12} />
                        {g.timerMinutes}m
                      </span>
                    ) : (
                      <span className="text-text-muted">-</span>
                    )}
                  </TableCell>
                  <TableCell className="text-center">
                    <span className="inline-flex items-center gap-1 text-text-secondary">
                      <Users size={12} />
                      {g.teamCount}
                    </span>
                  </TableCell>
                  <TableCell className="text-text-secondary">
                    {new Date(g.createdAt).toLocaleDateString()}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => navigate(`/admin/clients/${client}/games/${g.id}/status`)}
                        className="text-text-muted hover:text-accent"
                        title="View status"
                      >
                        <Eye size={14} />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteTarget(g)}
                        className="text-text-muted hover:text-error"
                      >
                        <Trash2 size={14} />
                      </Button>
                    </div>
                  </TableCell>
                </motion.tr>
              )
            })}
          </TableBody>
        </Table>
      )}

      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogHeader>
          <DialogTitle>Delete Game</DialogTitle>
        </DialogHeader>
        <p className="text-text-secondary mb-4">
          Are you sure you want to delete the game for "{deleteTarget?.scenarioName}"? This cannot be undone.
        </p>
        <div className="flex gap-3 justify-end">
          <Button variant="outline" onClick={() => setDeleteTarget(null)}>
            Cancel
          </Button>
          <MotionButton variant="destructive" onClick={handleDelete} disabled={deleting}>
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
    </div>
  )
}
