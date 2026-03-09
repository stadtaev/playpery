import { useState, useEffect, useCallback, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { getGameState, submitAnswer, unlockStage } from './api'
import { useGameEvents } from './useGameEvents'
import { useCountdown } from './TimerDisplay'
import { getSession, clearSession } from './lib/session'
import type { GameState, FunFact } from './types'

export type StagePhase = 'interstitial' | 'unlocking' | 'answering' | 'results'
export type Feedback = { correct: boolean; message: string }
export type AnswerResult = { isCorrect: boolean; correctAnswer: string; funFacts?: FunFact[] }

export function useGameState() {
  const { t } = useTranslation('player')
  const client = getSession()?.client || 'demo'
  const [state, setState] = useState<GameState | null>(null)
  const [answer, setAnswer] = useState('')
  const [unlockCode, setUnlockCode] = useState('')
  const [feedback, setFeedback] = useState<Feedback | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [stagePhase, setStagePhase] = useState<StagePhase>('interstitial')
  const stagePhaseRef = useRef<StagePhase>('interstitial')
  const [answerResult, setAnswerResult] = useState<AnswerResult | null>(null)
  const answeringRef = useRef(false) // true while submitAnswer is in-flight

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

  // Keep ref in sync so SSE callback can read current phase without re-creating.
  function updateStagePhase(phase: StagePhase) {
    stagePhaseRef.current = phase
    setStagePhase(phase)
  }

  // SSE handler: refetch state, and if stage was unlocked, transition phase.
  // Skip refetch during 'results' phase — we already have the answer data and
  // the server has already advanced currentStage, which would reset the phase.
  const onSSEEvent = useCallback((eventType?: string) => {
    if (eventType === 'stage_unlocked') {
      fetchState()
      setStagePhase((prev) => {
        const next = prev === 'unlocking' ? 'answering' : prev
        stagePhaseRef.current = next
        return next
      })
      setUnlockCode('')
    } else if (eventType === 'stage_completed' || eventType === 'wrong_answer') {
      // For non-submitters: fetch new state and show results from server.
      if (stagePhaseRef.current !== 'results' && !answeringRef.current) {
        getGameState(client).then((s) => {
          setState(s)
          setError('')
          if (s.lastResult) {
            setAnswerResult({
              isCorrect: s.lastResult.isCorrect,
              correctAnswer: s.lastResult.correctAnswer,
              funFacts: s.lastResult.funFacts,
            })
            updateStagePhase('results')
          }
        }).catch((e) => setError(e.message))
      }
    } else if (stagePhaseRef.current !== 'results' && !answeringRef.current) {
      fetchState()
    }
  }, [client, fetchState])

  useGameEvents(client, onSSEEvent)

  // Reset to interstitial when stage changes (e.g. teammate answered via SSE).
  // Skip reset if currently showing results — handleContinue will reset explicitly.
  const currentStageNumber = state?.currentStage?.stageNumber ?? null
  useEffect(() => {
    if (stagePhaseRef.current === 'results' || answeringRef.current) return
    updateStagePhase('interstitial')
    setFeedback(null)
    setAnswerResult(null)
    setAnswer('')
    setUnlockCode('')
  }, [currentStageNumber])

  // If the current stage arrives already unlocked, skip from unlocking to answering.
  useEffect(() => {
    if (!state?.currentStage) return
    const mode = state.game.mode
    if (mode === 'classic') return
    if (!state.currentStage.locked && stagePhase === 'unlocking') {
      updateStagePhase('answering')
    }
  }, [state?.currentStage?.locked, state?.game.mode, stagePhase])

  // Compute timer deadlines.
  const timerActive = state?.game.timerEnabled && state.game.status === 'active'

  const gameDeadline = timerActive && state.game.startedAt
    ? new Date(state.game.startedAt).getTime() + state.game.timerMinutes * 60000
    : null

  const stageDeadline = (timerActive && state.stageUnlockedAt && state.game.stageTimerMinutes)
    ? new Date(state.stageUnlockedAt).getTime() + state.game.stageTimerMinutes * 60000
    : null

  const gameRemaining = useCountdown(gameDeadline)
  const stageRemaining = useCountdown(stageDeadline)

  function handleGoToStage() {
    const mode = state?.game.mode || 'classic'
    setFeedback(null)
    if (mode === 'classic') {
      updateStagePhase('answering')
    } else {
      if (state?.currentStage && !state.currentStage.locked) {
        updateStagePhase('answering')
      } else {
        updateStagePhase('unlocking')
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
        setFeedback({ correct: true, message: t('stage_complete', { number: resp.stageNumber }) })
        setTimeout(() => {
          updateStagePhase('interstitial')
          fetchState()
        }, 1500)
      } else {
        updateStagePhase('answering')
      }
    } catch (e) {
      setFeedback({ correct: false, message: e instanceof Error ? e.message : t('error_generic') })
    } finally {
      setSubmitting(false)
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!answer.trim() || submitting) return
    setSubmitting(true)
    setFeedback(null)
    answeringRef.current = true
    try {
      const resp = await submitAnswer(client, answer.trim())
      setAnswer('')
      setAnswerResult({
        isCorrect: resp.isCorrect,
        correctAnswer: resp.correctAnswer,
        funFacts: resp.funFacts,
      })
      updateStagePhase('results')
    } catch (e) {
      setFeedback({ correct: false, message: e instanceof Error ? e.message : t('error_generic') })
    } finally {
      answeringRef.current = false
      setSubmitting(false)
    }
  }

  function handleContinue() {
    setFeedback(null)
    setAnswerResult(null)
    updateStagePhase('interstitial')
    fetchState()
  }

  function handleLogout() {
    clearSession()
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
    answerResult,
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
