import { useState, useEffect } from 'react'
import { listClients, createClient, type ClientInfo } from './adminApi'

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
    return <p aria-busy="true">Loading clients...</p>
  }

  return (
    <>
      <h2>Clients</h2>
      {error && <p role="alert" style={{ color: 'var(--pico-color-red-500)' }}>{error}</p>}

      {clients.length === 0 ? (
        <p>No clients yet.</p>
      ) : (
        <table>
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
                <td><code>{c.slug}</code></td>
                <td>{c.name}</td>
                <td>
                  <button
                    className="outline"
                    onClick={() => navigate(`/admin/clients/${c.slug}/scenarios`)}
                    style={{ width: 'auto', padding: '0.25rem 0.5rem', fontSize: 'small', marginRight: '0.25rem' }}
                  >
                    Scenarios
                  </button>
                  <button
                    className="outline"
                    onClick={() => navigate(`/admin/clients/${c.slug}/games`)}
                    style={{ width: 'auto', padding: '0.25rem 0.5rem', fontSize: 'small' }}
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
        <form onSubmit={handleCreate} style={{ marginTop: '0.5rem' }}>
          <div className="grid">
            <label>
              Slug
              <input type="text" value={newSlug} onChange={(e) => setNewSlug(e.target.value)} placeholder="e.g. acme" required />
            </label>
            <label>
              Name
              <input type="text" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="e.g. Acme Tours" required />
            </label>
          </div>
          <button type="submit" disabled={creating} aria-busy={creating} style={{ width: 'auto' }}>
            {creating ? 'Creating...' : 'Create Client'}
          </button>
        </form>
      </details>
    </>
  )
}
