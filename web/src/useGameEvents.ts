import { useEffect } from 'react'
import { getSession } from './lib/session'

export function useGameEvents(client: string, onEvent: (eventType?: string) => void) {
  useEffect(() => {
    const session = getSession()
    if (!session) return

    const es = new EventSource(`/api/${client}/game/events?token=${session.token}`)

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
