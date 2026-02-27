import { useState, useEffect, useCallback } from 'react'
import { getGameState, submitAnswer, unlockStage } from './api'
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

type StagePhase = 'interstitial' | 'unlocking' | 'answering'

export function GamePage() {
  const client = localStorage.getItem('client') || 'demo'
  const [state, setState] = useState<GameState | null>(null)
  const [answer, setAnswer] = useState('')
  const [unlockCode, setUnlockCode] = useState('')
  const [feedback, setFeedback] = useState<{ correct: boolean; message: string } | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [stagePhase, setStagePhase] = useState<StagePhase>('interstitial')
  const [phaseStartedAt, setPhaseStartedAt] = useState<number | null>(null)

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

  // SSE handler: refetch state, and if stage was unlocked, transition phase.
  // Only transition if we're in the unlocking phase — otherwise we'd skip interstitial.
  const onSSEEvent = useCallback((eventType?: string) => {
    fetchState()
    if (eventType === 'stage_unlocked') {
      setStagePhase((prev) => prev === 'unlocking' ? 'answering' : prev)
      setUnlockCode('')
    }
  }, [fetchState])

  useGameEvents(client, onSSEEvent)

  // Reset to interstitial when stage changes (e.g. teammate answered via SSE).
  const currentStageNumber = state?.currentStage?.stageNumber ?? null
  useEffect(() => {
    setStagePhase('interstitial')
    setPhaseStartedAt(null)
    setFeedback(null)
    setAnswer('')
    setUnlockCode('')
  }, [currentStageNumber])

  // If the current stage arrives already unlocked (e.g. via SSE), skip from unlocking to answering.
  useEffect(() => {
    if (!state?.currentStage) return
    const mode = state.game.mode
    if (mode === 'classic') return // classic doesn't use locked
    if (!state.currentStage.locked && stagePhase === 'unlocking') {
      setStagePhase('answering')
    }
  }, [state?.currentStage?.locked, state?.game.mode, stagePhase])

  // Compute timer deadlines.
  const timerActive = state?.game.timerEnabled && state.game.status === 'active'

  const gameDeadline = timerActive && state.game.startedAt
    ? new Date(state.game.startedAt).getTime() + state.game.timerMinutes * 60000
    : null

  const stageDeadline = (timerActive && phaseStartedAt && state.game.stageTimerMinutes)
    ? phaseStartedAt + state.game.stageTimerMinutes * 60000
    : null

  const gameRemaining = useCountdown(gameDeadline)
  const stageRemaining = useCountdown(stageDeadline)

  function handleGoToStage() {
    const mode = state?.game.mode || 'classic'
    const now = Date.now()
    setPhaseStartedAt(now)
    setFeedback(null)
    if (mode === 'classic') {
      setStagePhase('answering')
    } else {
      // Non-classic: if stage is already unlocked, go straight to answering
      if (state?.currentStage && !state.currentStage.locked) {
        setStagePhase('answering')
      } else {
        setStagePhase('unlocking')
      }
    }
  }

  async function handleUnlock(e: React.FormEvent) {
    e.preventDefault()
    if (submitting) return
    setSubmitting(true)
    setFeedback(null)
    try {
      const mode = state!.game.mode
      const code = mode === 'guided' ? '' : unlockCode.trim()
      const resp = await unlockStage(client, code)
      setUnlockCode('')
      if (resp.stageComplete) {
        // Stage auto-completed (qr_hunt, math_puzzle, guided without questions)
        setFeedback({ correct: true, message: `Stage ${resp.stageNumber} complete!` })
        setTimeout(() => {
          setStagePhase('interstitial')
          fetchState()
        }, 1500)
      } else {
        // Unlocked, now answer the question
        setStagePhase('answering')
      }
    } catch (e) {
      setFeedback({ correct: false, message: e instanceof Error ? e.message : 'Error' })
    } finally {
      setSubmitting(false)
    }
  }

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
        setStagePhase('interstitial')
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

  const { game, team, role, currentStage, completedStages, players } = state
  const isEnded = game.status === 'ended' || (!currentStage && completedStages.length === game.totalStages)
  const canAnswer = !game.supervised || role === 'supervisor'
  const mode = game.mode || 'classic'

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

      {currentStage && !isEnded && stagePhase === 'interstitial' && (
        <article>
          <header>
            Stage {currentStage.stageNumber} of {game.totalStages} &mdash; {currentStage.location}
          </header>
          <p><strong>Clue:</strong> {currentStage.clue}</p>
          <button onClick={handleGoToStage}>
            {completedStages.length === 0 ? 'Start Quest' : 'Go to Next Stage'}
          </button>
        </article>
      )}

      {currentStage && !isEnded && stagePhase === 'unlocking' && (
        <article>
          <header>
            Stage {currentStage.stageNumber} of {game.totalStages} &mdash; {currentStage.location}
          </header>
          <p><strong>Clue:</strong> {currentStage.clue}</p>

          {(mode === 'qr_quiz' || mode === 'qr_hunt') && (
            <form onSubmit={handleUnlock}>
              <p>Enter the code from the QR at this location:</p>
              <input
                type="text"
                value={unlockCode}
                onChange={(e) => setUnlockCode(e.target.value)}
                placeholder="QR code..."
                autoFocus
                required
              />
              {feedback && (
                <small style={{ color: feedback.correct ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
                  {feedback.message}
                </small>
              )}
              <button type="submit" disabled={submitting} aria-busy={submitting}>
                Submit Code
              </button>
            </form>
          )}

          {mode === 'math_puzzle' && (
            <form onSubmit={handleUnlock}>
              <p>Your team secret is: <strong>{state.teamSecret}</strong></p>
              <p>Location number: <strong>{state.currentStage?.locationNumber}</strong></p>
              <p>Add them together and enter the result:</p>
              <input
                type="text"
                value={unlockCode}
                onChange={(e) => setUnlockCode(e.target.value)}
                placeholder="Calculated code..."
                autoFocus
                required
              />
              {feedback && (
                <small style={{ color: feedback.correct ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
                  {feedback.message}
                </small>
              )}
              <button type="submit" disabled={submitting} aria-busy={submitting}>
                Submit Code
              </button>
            </form>
          )}

          {mode === 'guided' && (
            role === 'supervisor' ? (
              <form onSubmit={handleUnlock}>
                <p>Unlock this stage for your team:</p>
                {feedback && (
                  <small style={{ color: feedback.correct ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
                    {feedback.message}
                  </small>
                )}
                <button type="submit" disabled={submitting} aria-busy={submitting}>
                  Unlock Stage
                </button>
              </form>
            ) : (
              <p><em>Waiting for the guide to unlock this stage...</em></p>
            )
          )}
        </article>
      )}

      {currentStage && !isEnded && stagePhase === 'answering' && (
        <article>
          <header>
            Stage {currentStage.stageNumber} of {game.totalStages} &mdash; {currentStage.location}
          </header>
          <p><strong>Clue:</strong> {currentStage.clue}</p>
          {currentStage.question && (
            <p><strong>Question:</strong> {currentStage.question}</p>
          )}
          {feedback && !feedback.correct ? (
            <>
              <p style={{ color: 'var(--pico-color-red-500)' }}>
                {feedback.message}
              </p>
              <button onClick={() => { setFeedback(null); setStagePhase('interstitial'); fetchState() }}>
                Continue
              </button>
            </>
          ) : canAnswer ? (
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
          ) : (
            <p><em>Waiting for the supervisor to submit the answer...</em></p>
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
