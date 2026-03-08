import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getMe, logout } from './adminApi'
import { LoadingPage } from '../components/Spinner'

const ADMIN_LANG_KEY = 'cq_admin_lang'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminLayout({ client, children }: { client?: string; children: React.ReactNode }) {
  const { t, i18n } = useTranslation('admin')
  const [authed, setAuthed] = useState<boolean | null>(null)

  useEffect(() => {
    const saved = localStorage.getItem(ADMIN_LANG_KEY)
    if (saved) i18n.changeLanguage(saved)
  }, [i18n])

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

  function toggleLanguage() {
    const next = i18n.language === 'ru' ? 'en' : 'ru'
    i18n.changeLanguage(next)
    localStorage.setItem(ADMIN_LANG_KEY, next)
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
              {t('nav_title')}
            </a>
          </h1>
          <a href="/admin/scenarios" onClick={(e) => { e.preventDefault(); navigate('/admin/scenarios') }} className="text-sm uppercase tracking-widest font-bold">
            {t('nav_scenarios')}
          </a>
          {client && (
            <>
              <span className="text-secondary">/</span>
              <strong>{client}</strong>
              <a href={`/admin/clients/${client}/games`} onClick={(e) => { e.preventDefault(); navigate(`/admin/clients/${client}/games`) }} className="text-sm uppercase tracking-widest font-bold">
                {t('nav_games')}
              </a>
            </>
          )}
        </div>
        <div className="flex items-center gap-3">
          <button className="btn-ghost btn-sm" onClick={toggleLanguage}>
            {i18n.language === 'ru' ? t('lang_en') : t('lang_ru')}
          </button>
          <button className="btn-secondary btn-sm" onClick={handleLogout}>
            {t('nav_logout')}
          </button>
        </div>
      </nav>
      {children}
    </main>
  )
}
