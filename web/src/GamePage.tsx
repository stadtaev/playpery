import { useState, useEffect, useCallback } from 'react'
import { getGameState, submitAnswer } from './api'
import { useGameEvents } from './useGameEvents'
import type { GameState } from './types'

function TimeRemaining({ startedAt, timerMinutes }: { startedAt: string; timerMinutes: number }) {
  const [remaining, setRemaining] = useState('')

  useEffect(() => {
    function update() {
      const start = new Date(startedAt).getTime()
      const end = start + timerMinutes * 60 * 1000
      const left = Math.max(0, end - Date.now())
      const mins = Math.floor(left / 60000)
      const secs = Math.floor((left % 60000) / 1000)
      setRemaining(`${mins}:${secs.toString().padStart(2, '0')}`)
    }
    update()
    const id = setInterval(update, 1000)
    return () => clearInterval(id)
  }, [startedAt, timerMinutes])

  return <span>{remaining}</span>
}

export function GamePage() {
  const client = localStorage.getItem('client') || 'demo'
  const [state, setState] = useState<GameState | null>(null)
  const [answer, setAnswer] = useState('')
  const [feedback, setFeedback] = useState<{ correct: boolean; message: string } | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

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

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!answer.trim() || submitting) return
    setSubmitting(true)
    setFeedback(null)
    try {
      const resp = await submitAnswer(client, answer.trim())
      if (resp.isCorrect) {
        setFeedback({ correct: true, message: `Stage ${resp.stageNumber} complete!` })
        setAnswer('')
        fetchState()
      } else {
        console.log('[debug] correct answer:', resp.correctAnswer)
        setFeedback({ correct: false, message: 'Wrong answer, try again!' })
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
        <h1>CityQuiz</h1>
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
      <nav style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>CityQuiz</h1>
        <small>
          {team.name} &middot;{' '}
          {game.startedAt && (
            <TimeRemaining startedAt={game.startedAt} timerMinutes={game.timerMinutes} />
          )}
        </small>
      </nav>

      {isEnded && (
        <article>
          <header>Game Over!</header>
          <p>
            Your team completed {completedStages.length} of {game.totalStages} stages.
          </p>
        </article>
      )}

      {currentStage && !isEnded && (
        <article>
          <header>
            Stage {currentStage.stageNumber} of {game.totalStages} &mdash; {currentStage.location}
          </header>
          <p><strong>Clue:</strong> {currentStage.clue}</p>
          <p><strong>Question:</strong> {currentStage.question}</p>
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
              <small style={{ color: feedback.correct ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
                {feedback.message}
              </small>
            )}
            <button type="submit" disabled={submitting} aria-busy={submitting}>
              Submit Answer
            </button>
          </form>
        </article>
      )}

      {completedStages.length > 0 && (
        <details open={isEnded}>
          <summary>Completed Stages ({completedStages.length})</summary>
          <ul>
            {completedStages.map((s) => (
              <li key={s.stageNumber}>Stage {s.stageNumber} &mdash; completed</li>
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
