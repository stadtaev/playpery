import { useEffect } from 'react'

export function useGameEvents(onEvent: () => void) {
  useEffect(() => {
    const token = localStorage.getItem('session_token')
    if (!token) return

    const es = new EventSource(`/api/game/events?token=${token}`)

    es.addEventListener('state', () => {
      onEvent()
    })

    es.onerror = () => {
      // EventSource auto-reconnects. Nothing to do.
    }

    return () => es.close()
  }, [onEvent])
}
