import { useState, useEffect } from 'react'
import { JoinPage } from './JoinPage'
import { GamePage } from './GamePage'

function getRoute(): { page: 'join'; token: string } | { page: 'game' } | { page: 'home' } {
  const path = window.location.pathname

  const joinMatch = path.match(/^\/join\/(.+)$/)
  if (joinMatch) return { page: 'join', token: joinMatch[1] }

  if (path === '/game' && localStorage.getItem('session_token')) {
    return { page: 'game' }
  }

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
    default:
      return (
        <main className="container" style={{ maxWidth: 480, textAlign: 'center' }}>
          <h1>CityQuiz</h1>
          <p>Scan your team's QR code or use the join link to get started.</p>
        </main>
      )
  }
}
