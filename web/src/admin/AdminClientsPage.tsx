import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Plus, Map, Gamepad2 } from 'lucide-react'
import { listClients, createClient, type ClientInfo } from './adminApi'
import { navigate } from '@/lib/navigate'
import { Button, MotionButton } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Dialog, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

export function AdminClientsPage() {
  const [clients, setClients] = useState<ClientInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [newSlug, setNewSlug] = useState('')
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    listClients()
      .then(setClients)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!newSlug.trim() || !newName.trim()) return
    setCreating(true)
    setError('')
    try {
      const client = await createClient(newSlug.trim(), newName.trim())
      setClients((prev) => [...prev, client])
      setNewSlug('')
      setNewName('')
      setDialogOpen(false)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Create failed')
    } finally {
      setCreating(false)
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
        <h2 className="text-xl font-semibold text-text-primary">Clients</h2>
        <MotionButton onClick={() => setDialogOpen(true)}>
          <Plus size={16} />
          Add Client
        </MotionButton>
      </div>

      {error && (
        <div className="mb-4">
          <Alert variant="error">{error}</Alert>
        </div>
      )}

      {clients.length === 0 ? (
        <p className="text-text-secondary">No clients yet.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Slug</TableHead>
              <TableHead>Name</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {clients.map((c, i) => (
              <motion.tr
                key={c.slug}
                className="border-b border-border transition-colors hover:bg-popover"
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.04, duration: 0.25 }}
              >
                <TableCell>
                  <code className="rounded bg-input px-1.5 py-0.5 text-xs font-mono text-accent">
                    {c.slug}
                  </code>
                </TableCell>
                <TableCell>{c.name}</TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-1.5">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => navigate(`/admin/clients/${c.slug}/scenarios`)}
                    >
                      <Map size={14} />
                      Scenarios
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => navigate(`/admin/clients/${c.slug}/games`)}
                    >
                      <Gamepad2 size={14} />
                      Games
                    </Button>
                  </div>
                </TableCell>
              </motion.tr>
            ))}
          </TableBody>
        </Table>
      )}

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)}>
        <DialogHeader>
          <DialogTitle>Add Client</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleCreate} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="slug">Slug</Label>
            <Input
              id="slug"
              type="text"
              value={newSlug}
              onChange={(e) => setNewSlug(e.target.value)}
              placeholder="e.g. acme"
              autoFocus
              required
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="e.g. Acme Tours"
              required
            />
          </div>
          <MotionButton type="submit" disabled={creating} className="w-full">
            {creating ? (
              <>
                <Spinner size={16} className="text-accent-foreground" />
                Creating...
              </>
            ) : (
              <>
                <Plus size={16} />
                Create Client
              </>
            )}
          </MotionButton>
        </form>
      </Dialog>
    </div>
  )
}
