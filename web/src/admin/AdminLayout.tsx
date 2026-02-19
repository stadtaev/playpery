import { useState, useEffect } from 'react'
import { getMe, logout } from './adminApi'

export function AdminLayout({ children }: { children: React.ReactNode }) {
  const [authed, setAuthed] = useState<boolean | null>(null)

  useEffect(() => {
    getMe()
      .then(() => setAuthed(true))
      .catch(() => {
        setAuthed(false)
        window.history.replaceState(null, '', '/admin/login')
        window.dispatchEvent(new PopStateEvent('popstate'))
      })
  }, [])

  async function handleLogout() {
    await logout().catch(() => {})
    window.history.replaceState(null, '', '/admin/login')
    window.dispatchEvent(new PopStateEvent('popstate'))
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
        <h1 style={{ margin: 0 }}>
          <a href="/admin/scenarios" onClick={(e) => { e.preventDefault(); window.history.pushState(null, '', '/admin/scenarios'); window.dispatchEvent(new PopStateEvent('popstate')) }} style={{ textDecoration: 'none' }}>
            CityQuiz Admin
          </a>
        </h1>
        <button className="outline secondary" onClick={handleLogout} style={{ width: 'auto' }}>
          Log out
        </button>
      </nav>
      {children}
    </main>
  )
}
