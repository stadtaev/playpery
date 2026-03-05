import { useState, useEffect, useCallback } from 'react'
import { getGameState, submitAnswer, unlockStage } from './api'
import { useGameEvents } from './useGameEvents'
import { useCountdown } from './TimerDisplay'
import type { GameState } from './types'

export type StagePhase = 'interstitial' | 'unlocking' | 'answering'
export type Feedback = { correct: boolean; message: string }

export function useGameState() {
  const client = sessionStorage.getItem('client') || 'demo'
  const [state, setState] = useState<GameState | null>(null)
  const [answer, setAnswer] = useState('')
  const [unlockCode, setUnlockCode] = useState('')
  const [feedback, setFeedback] = useState<Feedback | null>(null)
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

  // If the current stage arrives already unlocked, skip from unlocking to answering.
  useEffect(() => {
    if (!state?.currentStage) return
    const mode = state.game.mode
    if (mode === 'classic') return
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
      const code = mode === 'supervised' ? '' : unlockCode.trim()
      const resp = await unlockStage(client, code)
      setUnlockCode('')
      if (resp.stageComplete) {
        setFeedback({ correct: true, message: `Stage ${resp.stageNumber} complete!` })
        setTimeout(() => {
          setStagePhase('interstitial')
          fetchState()
        }, 1500)
      } else {
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

  function handleContinue() {
    setFeedback(null)
    setStagePhase('interstitial')
    fetchState()
  }

  function handleLogout() {
    sessionStorage.removeItem('session_token')
    sessionStorage.removeItem('team_name')
    sessionStorage.removeItem('client')
    window.history.replaceState(null, '', '/')
    window.dispatchEvent(new PopStateEvent('popstate'))
  }

  return {
    state,
    error,
    stagePhase,
    answer,
    setAnswer,
    unlockCode,
    setUnlockCode,
    feedback,
    submitting,
    gameRemaining,
    stageRemaining,
    handleGoToStage,
    handleUnlock,
    handleSubmit,
    handleContinue,
    handleLogout,
  }
}
