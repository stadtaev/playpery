import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Plus, Trash2 } from 'lucide-react'
import { listScenarios, deleteScenario } from './adminApi'
import type { ScenarioSummary } from './adminTypes'
import { navigate } from '@/lib/navigate'
import { Button, MotionButton } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

const modeBadgeVariant: Record<string, 'default' | 'success' | 'warning' | 'info' | 'error'> = {
  classic: 'default',
  qr_quiz: 'info',
  qr_hunt: 'success',
  math_puzzle: 'warning',
  guided: 'error',
}

export function AdminScenariosPage() {
  const [scenarios, setScenarios] = useState<ScenarioSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<ScenarioSummary | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    listScenarios()
      .then(setScenarios)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  async function handleDelete() {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteScenario(deleteTarget.id)
      setScenarios((prev) => prev.filter((s) => s.id !== deleteTarget.id))
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
        <h2 className="text-xl font-semibold text-text-primary">Scenarios</h2>
        <MotionButton onClick={() => navigate('/admin/scenarios/new')}>
          <Plus size={16} />
          New Scenario
        </MotionButton>
      </div>

      {error && (
        <div className="mb-4">
          <Alert variant="error">{error}</Alert>
        </div>
      )}

      {scenarios.length === 0 ? (
        <p className="text-text-secondary">No scenarios yet.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>City</TableHead>
              <TableHead>Mode</TableHead>
              <TableHead className="text-center">Stages</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {scenarios.map((s, i) => {
              const mode = s.mode || 'classic'
              return (
                <motion.tr
                  key={s.id}
                  className="border-b border-border transition-colors hover:bg-popover"
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.04, duration: 0.25 }}
                >
                  <TableCell>
                    <button
                      className="text-accent hover:underline bg-transparent border-none cursor-pointer p-0 font-medium"
                      onClick={() => navigate(`/admin/scenarios/${s.id}/edit`)}
                    >
                      {s.name}
                    </button>
                  </TableCell>
                  <TableCell className="text-text-secondary">{s.city}</TableCell>
                  <TableCell>
                    <Badge variant={modeBadgeVariant[mode] ?? 'default'}>{mode}</Badge>
                  </TableCell>
                  <TableCell className="text-center">{s.stageCount}</TableCell>
                  <TableCell className="text-text-secondary">
                    {new Date(s.createdAt).toLocaleDateString()}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setDeleteTarget(s)}
                      className="text-text-muted hover:text-error"
                    >
                      <Trash2 size={14} />
                    </Button>
                  </TableCell>
                </motion.tr>
              )
            })}
          </TableBody>
        </Table>
      )}

      <Dialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)}>
        <DialogHeader>
          <DialogTitle>Delete Scenario</DialogTitle>
        </DialogHeader>
        <p className="text-text-secondary mb-4">
          Are you sure you want to delete "{deleteTarget?.name}"? This cannot be undone.
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
