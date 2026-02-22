import { useState, useEffect } from 'react'
import { JoinPage } from './JoinPage'
import { GamePage } from './GamePage'
import { AdminLoginPage } from './admin/AdminLoginPage'
import { AdminLayout } from './admin/AdminLayout'
import { AdminClientsPage } from './admin/AdminClientsPage'
import { AdminScenariosPage } from './admin/AdminScenariosPage'
import { AdminScenarioEditorPage } from './admin/AdminScenarioEditorPage'
import { AdminGamesPage } from './admin/AdminGamesPage'
import { AdminGameEditorPage } from './admin/AdminGameEditorPage'

type Route =
  | { page: 'join'; client: string; token: string }
  | { page: 'game' }
  | { page: 'home' }
  | { page: 'admin-login' }
  | { page: 'admin-clients' }
  | { page: 'admin-scenarios'; client: string }
  | { page: 'admin-scenario-new'; client: string }
  | { page: 'admin-scenario-edit'; client: string; id: string }
  | { page: 'admin-games'; client: string }
  | { page: 'admin-game-new'; client: string }
  | { page: 'admin-game-edit'; client: string; id: string }

function getRoute(): Route {
  const path = window.location.pathname

  // /join/{client}/{token}
  const joinMatch = path.match(/^\/join\/([^/]+)\/(.+)$/)
  if (joinMatch) return { page: 'join', client: joinMatch[1], token: joinMatch[2] }

  if (path === '/game' && localStorage.getItem('session_token')) {
    return { page: 'game' }
  }

  if (path === '/admin/login') return { page: 'admin-login' }
  if (path === '/admin' || path === '/admin/clients') return { page: 'admin-clients' }

  // /admin/clients/{client}/scenarios
  const scenariosMatch = path.match(/^\/admin\/clients\/([^/]+)\/scenarios$/)
  if (scenariosMatch) return { page: 'admin-scenarios', client: scenariosMatch[1] }

  const scenarioNewMatch = path.match(/^\/admin\/clients\/([^/]+)\/scenarios\/new$/)
  if (scenarioNewMatch) return { page: 'admin-scenario-new', client: scenarioNewMatch[1] }

  const scenarioEditMatch = path.match(/^\/admin\/clients\/([^/]+)\/scenarios\/(.+)\/edit$/)
  if (scenarioEditMatch) return { page: 'admin-scenario-edit', client: scenarioEditMatch[1], id: scenarioEditMatch[2] }

  // /admin/clients/{client}/games
  const gamesMatch = path.match(/^\/admin\/clients\/([^/]+)\/games$/)
  if (gamesMatch) return { page: 'admin-games', client: gamesMatch[1] }

  const gameNewMatch = path.match(/^\/admin\/clients\/([^/]+)\/games\/new$/)
  if (gameNewMatch) return { page: 'admin-game-new', client: gameNewMatch[1] }

  const gameEditMatch = path.match(/^\/admin\/clients\/([^/]+)\/games\/(.+)\/edit$/)
  if (gameEditMatch) return { page: 'admin-game-edit', client: gameEditMatch[1], id: gameEditMatch[2] }

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
      return <JoinPage client={route.client} joinToken={route.token} />
    case 'game':
      return <GamePage />
    case 'admin-login':
      return <AdminLoginPage />
    case 'admin-clients':
      return <AdminLayout><AdminClientsPage /></AdminLayout>
    case 'admin-scenarios':
      return <AdminLayout client={route.client}><AdminScenariosPage client={route.client} /></AdminLayout>
    case 'admin-scenario-new':
      return <AdminLayout client={route.client}><AdminScenarioEditorPage client={route.client} /></AdminLayout>
    case 'admin-scenario-edit':
      return <AdminLayout client={route.client}><AdminScenarioEditorPage client={route.client} id={route.id} /></AdminLayout>
    case 'admin-games':
      return <AdminLayout client={route.client}><AdminGamesPage client={route.client} /></AdminLayout>
    case 'admin-game-new':
      return <AdminLayout client={route.client}><AdminGameEditorPage client={route.client} /></AdminLayout>
    case 'admin-game-edit':
      return <AdminLayout client={route.client}><AdminGameEditorPage client={route.client} id={route.id} /></AdminLayout>
    default:
      return (
        <main className="container" style={{ maxWidth: 480, textAlign: 'center' }}>
          <h1>CityQuiz</h1>
          <p>Scan your team's QR code or use the join link to get started.</p>
        </main>
      )
  }
}
