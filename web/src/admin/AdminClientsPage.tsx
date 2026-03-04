import { useState, useEffect } from 'react'
import { listClients, createClient, type ClientInfo } from './adminApi'
import { LoadingPage, Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminClientsPage() {
  const [clients, setClients] = useState<ClientInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
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
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Create failed')
    } finally {
      setCreating(false)
    }
  }

  if (loading) {
    return <LoadingPage message="Loading clients..." />
  }

  return (
    <>
      <h2>Clients</h2>
      {error && <ErrorMessage message={error} />}

      {clients.length === 0 ? (
        <p className="text-secondary">No clients yet.</p>
      ) : (
        <table className="admin-table">
          <thead>
            <tr>
              <th>Slug</th>
              <th>Name</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {clients.map((c) => (
              <tr key={c.slug}>
                <td><code className="text-sm">{c.slug}</code></td>
                <td>{c.name}</td>
                <td className="whitespace-nowrap">
                  <button
                    className="btn-ghost btn-sm mr-1"
                    onClick={() => navigate(`/admin/clients/${c.slug}/scenarios`)}
                  >
                    Scenarios
                  </button>
                  <button
                    className="btn-ghost btn-sm"
                    onClick={() => navigate(`/admin/clients/${c.slug}/games`)}
                  >
                    Games
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <details>
        <summary>Add Client</summary>
        <form onSubmit={handleCreate} className="mt-4 space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="input-label" htmlFor="slug">Slug</label>
              <input id="slug" className="input" type="text" value={newSlug} onChange={(e) => setNewSlug(e.target.value)} placeholder="e.g. acme" required />
            </div>
            <div>
              <label className="input-label" htmlFor="name">Name</label>
              <input id="name" className="input" type="text" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="e.g. Acme Tours" required />
            </div>
          </div>
          <button type="submit" disabled={creating} className="btn">
            {creating ? <Spinner /> : 'Create Client'}
          </button>
        </form>
      </details>
    </>
  )
}
