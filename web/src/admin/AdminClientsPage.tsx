import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { listClients, createClient, type ClientInfo } from './adminApi'
import { LoadingPage, Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminClientsPage() {
  const { t } = useTranslation('admin')
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
      setError(e instanceof Error ? e.message : t('clients_create_failed'))
    } finally {
      setCreating(false)
    }
  }

  if (loading) {
    return <LoadingPage message={t('clients_loading')} />
  }

  return (
    <>
      <h2>{t('clients_title')}</h2>
      {error && <ErrorMessage message={error} />}

      {clients.length === 0 ? (
        <p className="text-secondary">{t('clients_empty')}</p>
      ) : (
        <table className="admin-table">
          <thead>
            <tr>
              <th>{t('clients_col_name')}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {clients.map((c) => (
              <tr key={c.slug}>
                <td>{c.name}</td>
                <td className="whitespace-nowrap">
                  <button
                    className="btn-ghost btn-sm mr-1"
                    onClick={() => navigate(`/admin/clients/${c.slug}/scenarios`)}
                  >
                    {t('clients_scenarios')}
                  </button>
                  <button
                    className="btn-ghost btn-sm"
                    onClick={() => navigate(`/admin/clients/${c.slug}/games`)}
                  >
                    {t('clients_games')}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <details>
        <summary>{t('clients_add')}</summary>
        <form onSubmit={handleCreate} className="mt-4 space-y-4">
          <div>
            <label className="input-label" htmlFor="name">{t('clients_col_name')}</label>
            <input id="name" className="input" type="text" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder={t('clients_name_placeholder')} required />
          </div>
          <div>
            <label className="input-label text-secondary" htmlFor="slug">{t('clients_slug')}</label>
            <input id="slug" className="input text-sm" type="text" value={newSlug} onChange={(e) => setNewSlug(e.target.value)} placeholder={t('clients_slug_placeholder')} required />
            <p className="text-secondary text-xs mt-1">{t('clients_slug_hint')}</p>
          </div>
          <button type="submit" disabled={creating} className="btn">
            {creating ? <Spinner /> : t('clients_create')}
          </button>
        </form>
      </details>
    </>
  )
}
