import { useEffect } from 'react'

export function useGameEvents(client: string, onEvent: () => void) {
  useEffect(() => {
    const token = localStorage.getItem('session_token')
    if (!token) return

    const es = new EventSource(`/api/${client}/game/events?token=${token}`)

    es.addEventListener('state', () => {
      onEvent()
    })

    es.onerror = () => {
      // EventSource auto-reconnects. Nothing to do.
    }

    return () => es.close()
  }, [client, onEvent])
}
