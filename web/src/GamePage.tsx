import { useState, useEffect, useCallback } from 'react'
import { getGameState, submitAnswer } from './api'
import { useGameEvents } from './useGameEvents'
import type { GameState } from './types'

function useCountdown(deadline: number | null) {
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

function TimerDisplay({ gameRemaining, stageRemaining }: { gameRemaining: number | null; stageRemaining: number | null }) {
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

export function GamePage() {
  const client = localStorage.getItem('client') || 'demo'
  const [state, setState] = useState<GameState | null>(null)
  const [answer, setAnswer] = useState('')
  const [feedback, setFeedback] = useState<{ correct: boolean; message: string } | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [stageStartedAt, setStageStartedAt] = useState<number | null>(null)

  const fetchState = useCallback(() => {
    getGameState(client)
      .then((s) => {
        setState(s)
        setError('')
      })
      .catch((e) => setError(e.message))
  }, [client])

  useEffect(() => {
    fetchState()
  }, [fetchState])

  useGameEvents(client, fetchState)

  // Reset interstitial when stage changes (e.g. teammate answered via SSE).
  const currentStageNumber = state?.currentStage?.stageNumber ?? null
  useEffect(() => {
    setStageStartedAt(null)
  }, [currentStageNumber])

  // Compute timer deadlines.
  const timerActive = state?.game.timerEnabled && state.game.status === 'active'

  const gameDeadline = timerActive && state.game.startedAt
    ? new Date(state.game.startedAt).getTime() + state.game.timerMinutes * 60000
    : null

  const stageDeadline = (timerActive && stageStartedAt && state.game.stageTimerMinutes)
    ? stageStartedAt + state.game.stageTimerMinutes * 60000
    : null

  const gameRemaining = useCountdown(gameDeadline)
  const stageRemaining = useCountdown(stageDeadline)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!answer.trim() || submitting) return
    setSubmitting(true)
    setFeedback(null)
    try {
      const resp = await submitAnswer(client, answer.trim())
      setAnswer('')
      if (resp.isCorrect) {
        setFeedback({ correct: true, message: `Stage ${resp.stageNumber} complete!` })
        setStageStartedAt(null)
        fetchState()
      } else {
        setFeedback({ correct: false, message: `Incorrect — the correct answer was: ${resp.correctAnswer}` })
      }
    } catch (e) {
      setFeedback({ correct: false, message: e instanceof Error ? e.message : 'Error' })
    } finally {
      setSubmitting(false)
    }
  }

  function handleLogout() {
    localStorage.removeItem('session_token')
    localStorage.removeItem('team_name')
    localStorage.removeItem('client')
    window.history.replaceState(null, '', '/')
    window.dispatchEvent(new PopStateEvent('popstate'))
  }

  if (error) {
    return (
      <main className="container">
        <h1>CityQuest</h1>
        <p role="alert">{error}</p>
        <button onClick={handleLogout}>Back to start</button>
      </main>
    )
  }

  if (!state) {
    return (
      <main className="container">
        <p aria-busy="true">Loading game...</p>
      </main>
    )
  }

  const { game, team, currentStage, completedStages, players } = state
  const isEnded = game.status === 'ended' || (!currentStage && completedStages.length === game.totalStages)

  return (
    <main className="container" style={{ maxWidth: 600 }}>
      {game.timerEnabled && !isEnded && (
        <TimerDisplay gameRemaining={gameRemaining} stageRemaining={stageRemaining} />
      )}
      <nav style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>CityQuest</h1>
        <small>{team.name}</small>
      </nav>

      {isEnded && (
        <article>
          <header>Game Over!</header>
          <p>
            Your team answered {completedStages.filter((s) => s.isCorrect).length} of {game.totalStages} correctly.
          </p>
        </article>
      )}

      {currentStage && !isEnded && stageStartedAt === null && (
        <article>
          <header>
            Stage {currentStage.stageNumber} of {game.totalStages} &mdash; {currentStage.location}
          </header>
          <p><strong>Clue:</strong> {currentStage.clue}</p>
          <button onClick={() => setStageStartedAt(Date.now())}>
            {completedStages.length === 0 ? 'Start Quest' : 'Go to Next Stage'}
          </button>
        </article>
      )}

      {currentStage && !isEnded && stageStartedAt !== null && (
        <article>
          <header>
            Stage {currentStage.stageNumber} of {game.totalStages} &mdash; {currentStage.location}
          </header>
          <p><strong>Clue:</strong> {currentStage.clue}</p>
          <p><strong>Question:</strong> {currentStage.question}</p>
          {feedback && !feedback.correct ? (
            <>
              <p style={{ color: 'var(--pico-color-red-500)' }}>
                {feedback.message}
              </p>
              <button onClick={() => { setFeedback(null); setStageStartedAt(null); fetchState() }}>
                Continue
              </button>
            </>
          ) : (
            <form onSubmit={handleSubmit}>
              <input
                type="text"
                value={answer}
                onChange={(e) => setAnswer(e.target.value)}
                placeholder="Your answer..."
                autoFocus
                required
              />
              {feedback && (
                <small style={{ color: 'var(--pico-color-green-500)' }}>
                  {feedback.message}
                </small>
              )}
              <button type="submit" disabled={submitting} aria-busy={submitting}>
                Submit Answer
              </button>
            </form>
          )}
        </article>
      )}

      {completedStages.length > 0 && (
        <details open={isEnded}>
          <summary>Completed Stages ({completedStages.length})</summary>
          <ul>
            {completedStages.map((s) => (
              <li key={s.stageNumber} style={{ color: s.isCorrect ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
                Stage {s.stageNumber} &mdash; {s.isCorrect ? 'correct' : 'incorrect'}
              </li>
            ))}
          </ul>
        </details>
      )}

      <details>
        <summary>Team ({players.length} players)</summary>
        <ul>
          {players.map((p) => (
            <li key={p.id}>{p.name}</li>
          ))}
        </ul>
      </details>

      <p style={{ textAlign: 'center', marginTop: '2rem' }}>
        <a href="#" onClick={(e) => { e.preventDefault(); handleLogout() }} style={{ fontSize: 'small' }}>
          Leave game
        </a>
      </p>
    </main>
  )
}
