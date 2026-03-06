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

  return (
    <div className={`fixed top-0 left-0 z-50 flex flex-col`}>
      {gameRemaining !== null && (
        <div className={`px-3 py-1.5 font-mono text-sm font-bold text-white ${isUrgent ? 'bg-[#d4351c]' : 'bg-[#018849]'}`}>
          {formatTime(gameRemaining)}
        </div>
      )}
      {stageRemaining !== null && (
        <div className={`px-3 py-1.5 font-mono text-sm font-bold text-white ${stageRemaining <= 60000 ? 'bg-[#d4351c]' : 'bg-[#018849]'}`}>
          ⏱ {formatTime(stageRemaining)}
        </div>
      )}
    </div>
  )
}
