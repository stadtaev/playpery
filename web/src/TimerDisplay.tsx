import { useState, useEffect } from 'react'

export function useCountdown(deadline: number | null) {
  const [remaining, setRemaining] = useState<number | null>(null)

  useEffect(() => {
    if (deadline === null) { setRemaining(null); return }
    function update() {
      setRemaining(Math.max(0, deadline! - Date.now()))
    }
    update()
    const id = setInterval(update, 1000)
    return () => clearInterval(id)
  }, [deadline])

  return remaining
}

function formatTime(ms: number): string {
  const mins = Math.floor(ms / 60000)
  const secs = Math.floor((ms % 60000) / 1000)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

export function TimerDisplay({ gameRemaining, stageRemaining }: { gameRemaining: number | null; stageRemaining: number | null }) {
  if (gameRemaining === null && stageRemaining === null) return null

  const minRemaining = Math.min(
    gameRemaining ?? Infinity,
    stageRemaining ?? Infinity,
  )
  const isUrgent = minRemaining <= 60000
  const color = isUrgent ? '#e53e3e' : '#38a169'

  return (
    <div style={{ position: 'fixed', top: '0.75rem', left: '0.75rem', zIndex: 1000, fontFamily: 'monospace', fontSize: '1.1rem', fontWeight: 'bold', color }}>
      {gameRemaining !== null && <div>Game: {formatTime(gameRemaining)}</div>}
      {stageRemaining !== null && <div>Stage: {formatTime(stageRemaining)}</div>}
    </div>
  )
}
