import { useEffect } from 'react'

export function useGameEvents(client: string, onEvent: (eventType?: string) => void) {
  useEffect(() => {
    const token = localStorage.getItem('session_token')
    if (!token) return

    const es = new EventSource(`/api/${client}/game/events?token=${token}`)

    es.addEventListener('state', (e) => {
      let eventType: string | undefined
      try {
        const data = JSON.parse(e.data)
        eventType = data.type
      } catch {
        // ignore parse errors
      }
      onEvent(eventType)
    })

    es.onerror = () => {
      // EventSource auto-reconnects. Nothing to do.
    }

    return () => es.close()
  }, [client, onEvent])
}
