import { useState, useEffect } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { JoinPage } from './JoinPage'
import { GamePage } from './GamePage'
import { AdminLoginPage } from './admin/AdminLoginPage'
import { AdminLayout } from './admin/AdminLayout'
import { AdminClientsPage } from './admin/AdminClientsPage'
import { AdminScenariosPage } from './admin/AdminScenariosPage'
import { AdminScenarioEditorPage } from './admin/AdminScenarioEditorPage'
import { AdminGamesPage } from './admin/AdminGamesPage'
import { AdminGameEditorPage } from './admin/AdminGameEditorPage'
import { AdminGameStatusPage } from './admin/AdminGameStatusPage'

type Route =
  | { page: 'join'; client: string; token: string }
  | { page: 'game' }
  | { page: 'home' }
  | { page: 'admin-login' }
  | { page: 'admin-clients' }
  | { page: 'admin-scenarios' }
  | { page: 'admin-scenario-new' }
  | { page: 'admin-scenario-edit'; id: string }
  | { page: 'admin-games'; client: string }
  | { page: 'admin-game-new'; client: string }
  | { page: 'admin-game-edit'; client: string; id: string }
  | { page: 'admin-game-status'; client: string; id: string }

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

  // /admin/scenarios (global)
  if (path === '/admin/scenarios') return { page: 'admin-scenarios' }

  const scenarioNewMatch = path.match(/^\/admin\/scenarios\/new$/)
  if (scenarioNewMatch) return { page: 'admin-scenario-new' }

  const scenarioEditMatch = path.match(/^\/admin\/scenarios\/(.+)\/edit$/)
  if (scenarioEditMatch) return { page: 'admin-scenario-edit', id: scenarioEditMatch[1] }

  // /admin/clients/{client}/games
  const gamesMatch = path.match(/^\/admin\/clients\/([^/]+)\/games$/)
  if (gamesMatch) return { page: 'admin-games', client: gamesMatch[1] }

  const gameNewMatch = path.match(/^\/admin\/clients\/([^/]+)\/games\/new$/)
  if (gameNewMatch) return { page: 'admin-game-new', client: gameNewMatch[1] }

  const gameStatusMatch = path.match(/^\/admin\/clients\/([^/]+)\/games\/(.+)\/status$/)
  if (gameStatusMatch) return { page: 'admin-game-status', client: gameStatusMatch[1], id: gameStatusMatch[2] }

  const gameEditMatch = path.match(/^\/admin\/clients\/([^/]+)\/games\/(.+)\/edit$/)
  if (gameEditMatch) return { page: 'admin-game-edit', client: gameEditMatch[1], id: gameEditMatch[2] }

  return { page: 'home' }
}

const pageTransition = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
  transition: { duration: 0.15 },
}

function routeKey(route: Route): string {
  switch (route.page) {
    case 'join': return `join-${route.client}-${route.token}`
    case 'admin-scenario-edit': return `admin-scenario-edit-${route.id}`
    case 'admin-games': return `admin-games-${route.client}`
    case 'admin-game-new': return `admin-game-new-${route.client}`
    case 'admin-game-edit': return `admin-game-edit-${route.client}-${route.id}`
    case 'admin-game-status': return `admin-game-status-${route.client}-${route.id}`
    default: return route.page
  }
}

function renderRoute(route: Route) {
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
      return <AdminLayout><AdminScenariosPage /></AdminLayout>
    case 'admin-scenario-new':
      return <AdminLayout><AdminScenarioEditorPage /></AdminLayout>
    case 'admin-scenario-edit':
      return <AdminLayout><AdminScenarioEditorPage id={route.id} /></AdminLayout>
    case 'admin-games':
      return <AdminLayout client={route.client}><AdminGamesPage client={route.client} /></AdminLayout>
    case 'admin-game-new':
      return <AdminLayout client={route.client}><AdminGameEditorPage client={route.client} /></AdminLayout>
    case 'admin-game-status':
      return <AdminLayout client={route.client}><AdminGameStatusPage client={route.client} id={route.id} /></AdminLayout>
    case 'admin-game-edit':
      return <AdminLayout client={route.client}><AdminGameEditorPage client={route.client} id={route.id} /></AdminLayout>
    default:
      return (
        <div className="min-h-screen flex items-center justify-center bg-background px-4">
          <div className="text-center">
            <h1 className="text-2xl font-semibold text-text-primary mb-2">CityQuest</h1>
            <p className="text-text-secondary">Scan your team's QR code or use the join link to get started.</p>
          </div>
        </div>
      )
  }
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

  return (
    <AnimatePresence mode="wait">
      <motion.div key={routeKey(route)} {...pageTransition}>
        {renderRoute(route)}
      </motion.div>
    </AnimatePresence>
  )
}
