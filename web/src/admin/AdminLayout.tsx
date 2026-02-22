import { useState, useEffect } from 'react'
import { getMe, logout } from './adminApi'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminLayout({ client, children }: { client?: string; children: React.ReactNode }) {
  const [authed, setAuthed] = useState<boolean | null>(null)

  useEffect(() => {
    getMe()
      .then(() => setAuthed(true))
      .catch(() => {
        setAuthed(false)
        navigate('/admin/login')
      })
  }, [])

  async function handleLogout() {
    await logout().catch(() => {})
    navigate('/admin/login')
  }

  if (authed === null) {
    return (
      <main className="container">
        <p aria-busy="true">Loading...</p>
      </main>
    )
  }

  if (!authed) return null

  return (
    <main className="container">
      <nav style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
          <h1 style={{ margin: 0 }}>
            <a href="/admin/clients" onClick={(e) => { e.preventDefault(); navigate('/admin/clients') }} style={{ textDecoration: 'none' }}>
              CityQuest Admin
            </a>
          </h1>
          <a href="/admin/scenarios" onClick={(e) => { e.preventDefault(); navigate('/admin/scenarios') }}>Scenarios</a>
          {client && (
            <>
              <span style={{ color: 'var(--pico-muted-color)' }}>/</span>
              <strong>{client}</strong>
              <a href={`/admin/clients/${client}/games`} onClick={(e) => { e.preventDefault(); navigate(`/admin/clients/${client}/games`) }}>Games</a>
            </>
          )}
        </div>
        <button className="outline secondary" onClick={handleLogout} style={{ width: 'auto' }}>
          Log out
        </button>
      </nav>
      {children}
    </main>
  )
}
