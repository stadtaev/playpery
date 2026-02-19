import { useState, useEffect } from 'react'
import { JoinPage } from './JoinPage'
import { GamePage } from './GamePage'
import { AdminLoginPage } from './admin/AdminLoginPage'
import { AdminLayout } from './admin/AdminLayout'
import { AdminScenariosPage } from './admin/AdminScenariosPage'
import { AdminScenarioEditorPage } from './admin/AdminScenarioEditorPage'

type Route =
  | { page: 'join'; token: string }
  | { page: 'game' }
  | { page: 'home' }
  | { page: 'admin-login' }
  | { page: 'admin-scenarios' }
  | { page: 'admin-scenario-new' }
  | { page: 'admin-scenario-edit'; id: string }

function getRoute(): Route {
  const path = window.location.pathname

  const joinMatch = path.match(/^\/join\/(.+)$/)
  if (joinMatch) return { page: 'join', token: joinMatch[1] }

  if (path === '/game' && localStorage.getItem('session_token')) {
    return { page: 'game' }
  }

  if (path === '/admin/login') return { page: 'admin-login' }
  if (path === '/admin/scenarios') return { page: 'admin-scenarios' }
  if (path === '/admin/scenarios/new') return { page: 'admin-scenario-new' }

  const editMatch = path.match(/^\/admin\/scenarios\/(.+)\/edit$/)
  if (editMatch) return { page: 'admin-scenario-edit', id: editMatch[1] }

  return { page: 'home' }
}

export default function App() {
  const [route, setRoute] = useState(getRoute)

  useEffect(() => {
    function onNav() {
      setRoute(getRoute())
    }
    window.addEventListener('popstate', onNav)
    return () => window.removeEventListener('popstate', onNav)
  }, [])

  switch (route.page) {
    case 'join':
      return <JoinPage joinToken={route.token} />
    case 'game':
      return <GamePage />
    case 'admin-login':
      return <AdminLoginPage />
    case 'admin-scenarios':
      return <AdminLayout><AdminScenariosPage /></AdminLayout>
    case 'admin-scenario-new':
      return <AdminLayout><AdminScenarioEditorPage /></AdminLayout>
    case 'admin-scenario-edit':
      return <AdminLayout><AdminScenarioEditorPage id={route.id} /></AdminLayout>
    default:
      return (
        <main className="container" style={{ maxWidth: 480, textAlign: 'center' }}>
          <h1>CityQuiz</h1>
          <p>Scan your team's QR code or use the join link to get started.</p>
        </main>
      )
  }
}
