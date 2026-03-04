import { useState, useEffect } from 'react'
import { getMe, logout } from './adminApi'
import { LoadingPage } from '../components/Spinner'

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
    return <LoadingPage />
  }

  if (!authed) return null

  return (
    <main className="page-wide">
      <nav className="flex justify-between items-center mb-8">
        <div className="flex items-center gap-6">
          <h1 className="m-0">
            <a href="/admin/clients" onClick={(e) => { e.preventDefault(); navigate('/admin/clients') }} className="no-underline">
              CityQuest Admin
            </a>
          </h1>
          <a href="/admin/scenarios" onClick={(e) => { e.preventDefault(); navigate('/admin/scenarios') }} className="text-sm uppercase tracking-widest font-bold">
            Scenarios
          </a>
          {client && (
            <>
              <span className="text-secondary">/</span>
              <strong>{client}</strong>
              <a href={`/admin/clients/${client}/games`} onClick={(e) => { e.preventDefault(); navigate(`/admin/clients/${client}/games`) }} className="text-sm uppercase tracking-widest font-bold">
                Games
              </a>
            </>
          )}
        </div>
        <button className="btn-secondary btn-sm" onClick={handleLogout}>
          Log out
        </button>
      </nav>
      {children}
    </main>
  )
}
