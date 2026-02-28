import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Clock, MapPin, LogOut, Users, Trophy, CheckCircle2, XCircle,
  Lock, Unlock, Send, ArrowRight, UserCog, Hash, QrCode, Play,
} from 'lucide-react'
import { getGameState, submitAnswer, unlockStage } from './api'
import { useGameEvents } from './useGameEvents'
import type { GameState } from './types'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { MotionButton, Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

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

  return (
    <div className={`fixed top-3 left-3 z-50 flex flex-col gap-1 rounded-lg border border-border bg-card/80 backdrop-blur-sm px-3 py-2 font-mono text-sm ${isUrgent ? 'animate-pulse' : ''}`}>
      {gameRemaining !== null && (
        <div className={`flex items-center gap-1.5 ${isUrgent ? 'text-error' : 'text-success'}`}>
          <Clock size={14} />
          <span>Game: {formatTime(gameRemaining)}</span>
        </div>
      )}
      {stageRemaining !== null && (
        <div className={`flex items-center gap-1.5 ${isUrgent ? 'text-error' : 'text-accent'}`}>
          <Clock size={14} />
          <span>Stage: {formatTime(stageRemaining)}</span>
        </div>
      )}
    </div>
  )
}

function ProgressDots({ totalStages, completedStages }: { totalStages: number; completedStages: { stageNumber: number; isCorrect: boolean }[] }) {
  const completed = new Map(completedStages.map(s => [s.stageNumber, s.isCorrect]))

  return (
    <div className="flex items-center justify-center gap-1.5 py-3">
      {Array.from({ length: totalStages }, (_, i) => {
        const num = i + 1
        const result = completed.get(num)
        return (
          <div
            key={num}
            className={`h-2.5 w-2.5 rounded-full transition-colors ${
              result === true
                ? 'bg-success'
                : result === false
                  ? 'bg-error'
                  : 'bg-border'
            }`}
            title={`Stage ${num}${result === true ? ' - correct' : result === false ? ' - incorrect' : ''}`}
          />
        )
      })}
    </div>
  )
}

const phaseTransition = {
  initial: { opacity: 0, x: 30 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: -30 },
  transition: { duration: 0.25, ease: 'easeInOut' as const },
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

  const onSSEEvent = useCallback((eventType?: string) => {
    fetchState()
    if (eventType === 'stage_unlocked') {
      setStagePhase((prev) => prev === 'unlocking' ? 'answering' : prev)
      setUnlockCode('')
    }
  }, [fetchState])

  useGameEvents(client, onSSEEvent)

  const currentStageNumber = state?.currentStage?.stageNumber ?? null
  useEffect(() => {
    setStagePhase('interstitial')
    setPhaseStartedAt(null)
    setFeedback(null)
    setAnswer('')
    setUnlockCode('')
  }, [currentStageNumber])

  useEffect(() => {
    if (!state?.currentStage) return
    const mode = state.game.mode
    if (mode === 'classic') return
    if (!state.currentStage.locked && stagePhase === 'unlocking') {
      setStagePhase('answering')
    }
  }, [state?.currentStage?.locked, state?.game.mode, stagePhase])

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
      const code = mode === 'guided' ? '' : unlockCode.trim()
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

  function handleLogout() {
    localStorage.removeItem('session_token')
    localStorage.removeItem('team_name')
    localStorage.removeItem('client')
    window.history.replaceState(null, '', '/')
    window.dispatchEvent(new PopStateEvent('popstate'))
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background px-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, ease: 'easeOut' }}
          className="w-full max-w-md"
        >
          <Card>
            <CardHeader>
              <h1 className="text-xl font-semibold text-text-primary">CityQuest</h1>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <Alert variant="error">{error}</Alert>
              <Button variant="ghost" onClick={handleLogout} className="w-full">
                <LogOut size={16} />
                Back to start
              </Button>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    )
  }

  if (!state) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background px-4">
        <Spinner />
      </div>
    )
  }

  const { game, team, role, currentStage, completedStages, players } = state
  const isEnded = game.status === 'ended' || (!currentStage && completedStages.length === game.totalStages)
  const mode = game.mode || 'classic'
  const canAnswer = !game.supervised || role === 'supervisor' || mode === 'guided'

  return (
    <div className="min-h-screen bg-background px-4 py-6">
      <div className="mx-auto max-w-lg flex flex-col gap-4">
        {game.timerEnabled && !isEnded && (
          <TimerDisplay gameRemaining={gameRemaining} stageRemaining={stageRemaining} />
        )}

        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold text-text-primary">CityQuest</h1>
          <div className="flex items-center gap-2">
            <Badge>
              <Users size={12} className="mr-1" />
              {team.name}
            </Badge>
            {role === 'supervisor' && (
              <Badge variant="warning">
                <UserCog size={12} className="mr-1" />
                Guide
              </Badge>
            )}
          </div>
        </div>

        {/* Progress dots */}
        {game.totalStages > 0 && (
          <ProgressDots totalStages={game.totalStages} completedStages={completedStages} />
        )}

        {/* Game Over */}
        {isEnded && (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.4 }}
          >
            <Card>
              <CardHeader>
                <div className="flex items-center gap-2">
                  <Trophy size={20} className="text-accent" />
                  <h2 className="text-lg font-semibold text-text-primary">Game Over!</h2>
                </div>
              </CardHeader>
              <CardContent className="flex flex-col gap-4">
                <p className="text-text-secondary">
                  Your team answered {completedStages.filter((s) => s.isCorrect).length} of {game.totalStages} correctly.
                </p>
                {completedStages.length > 0 && (
                  <div className="flex flex-col gap-1.5">
                    {completedStages.map((s) => (
                      <div key={s.stageNumber} className="flex items-center gap-2 text-sm">
                        {s.isCorrect ? (
                          <CheckCircle2 size={14} className="text-success" />
                        ) : (
                          <XCircle size={14} className="text-error" />
                        )}
                        <span className="text-text-secondary">
                          Stage {s.stageNumber} — {s.isCorrect ? 'correct' : 'incorrect'}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Stage phases */}
        <AnimatePresence mode="wait">
          {currentStage && !isEnded && stagePhase === 'interstitial' && (
            <motion.div key="interstitial" {...phaseTransition}>
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <MapPin size={18} className="text-accent" />
                      <h2 className="text-lg font-semibold text-text-primary">
                        Stage {currentStage.stageNumber}
                        <span className="text-text-muted font-normal"> / {game.totalStages}</span>
                      </h2>
                    </div>
                  </div>
                  <p className="text-sm text-text-muted">{currentStage.location}</p>
                </CardHeader>
                <CardContent className="flex flex-col gap-4">
                  <p className="text-text-secondary">{currentStage.clue}</p>
                  <MotionButton onClick={handleGoToStage} className="w-full">
                    {completedStages.length === 0 ? (
                      <>
                        <Play size={16} />
                        Start Quest
                      </>
                    ) : (
                      <>
                        <ArrowRight size={16} />
                        Go to Next Stage
                      </>
                    )}
                  </MotionButton>
                </CardContent>
              </Card>
            </motion.div>
          )}

          {currentStage && !isEnded && stagePhase === 'unlocking' && (
            <motion.div key="unlocking" {...phaseTransition}>
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Lock size={18} className="text-accent" />
                    <h2 className="text-lg font-semibold text-text-primary">
                      Stage {currentStage.stageNumber}
                      <span className="text-text-muted font-normal"> / {game.totalStages}</span>
                    </h2>
                  </div>
                  <p className="text-sm text-text-muted">{currentStage.location}</p>
                </CardHeader>
                <CardContent className="flex flex-col gap-4">
                  <p className="text-text-secondary">{currentStage.clue}</p>

                  {(mode === 'qr_quiz' || mode === 'qr_hunt') && (
                    <form onSubmit={handleUnlock} className="flex flex-col gap-3">
                      <p className="text-sm text-text-muted flex items-center gap-1.5">
                        <QrCode size={14} />
                        Enter the code from the QR at this location:
                      </p>
                      <Input
                        type="text"
                        value={unlockCode}
                        onChange={(e) => setUnlockCode(e.target.value)}
                        placeholder="QR code..."
                        autoFocus
                        required
                      />
                      <AnimatePresence>
                        {feedback && (
                          <motion.div
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: 'auto' }}
                            exit={{ opacity: 0, height: 0 }}
                          >
                            <Alert variant={feedback.correct ? 'success' : 'error'}>
                              {feedback.message}
                            </Alert>
                          </motion.div>
                        )}
                      </AnimatePresence>
                      <MotionButton type="submit" disabled={submitting} className="w-full">
                        {submitting ? (
                          <><Spinner size={16} className="text-accent-foreground" /> Submitting...</>
                        ) : (
                          <><Unlock size={16} /> Submit Code</>
                        )}
                      </MotionButton>
                    </form>
                  )}

                  {mode === 'math_puzzle' && (
                    <form onSubmit={handleUnlock} className="flex flex-col gap-3">
                      <div className="flex items-center gap-3 rounded-lg bg-input p-3 text-sm">
                        <Hash size={14} className="text-accent" />
                        <div>
                          <p className="text-text-secondary">Team secret: <strong className="text-text-primary">{state.teamSecret}</strong></p>
                          <p className="text-text-secondary">Location number: <strong className="text-text-primary">{state.currentStage?.locationNumber}</strong></p>
                          <p className="text-text-muted text-xs mt-1">Add them together and enter the result</p>
                        </div>
                      </div>
                      <Input
                        type="text"
                        value={unlockCode}
                        onChange={(e) => setUnlockCode(e.target.value)}
                        placeholder="Calculated code..."
                        autoFocus
                        required
                      />
                      <AnimatePresence>
                        {feedback && (
                          <motion.div
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: 'auto' }}
                            exit={{ opacity: 0, height: 0 }}
                          >
                            <Alert variant={feedback.correct ? 'success' : 'error'}>
                              {feedback.message}
                            </Alert>
                          </motion.div>
                        )}
                      </AnimatePresence>
                      <MotionButton type="submit" disabled={submitting} className="w-full">
                        {submitting ? (
                          <><Spinner size={16} className="text-accent-foreground" /> Submitting...</>
                        ) : (
                          <><Unlock size={16} /> Submit Code</>
                        )}
                      </MotionButton>
                    </form>
                  )}

                  {mode === 'guided' && (
                    role === 'supervisor' ? (
                      <form onSubmit={handleUnlock} className="flex flex-col gap-3">
                        <p className="text-sm text-text-muted">Unlock this stage for your team:</p>
                        <AnimatePresence>
                          {feedback && (
                            <motion.div
                              initial={{ opacity: 0, height: 0 }}
                              animate={{ opacity: 1, height: 'auto' }}
                              exit={{ opacity: 0, height: 0 }}
                            >
                              <Alert variant={feedback.correct ? 'success' : 'error'}>
                                {feedback.message}
                              </Alert>
                            </motion.div>
                          )}
                        </AnimatePresence>
                        <MotionButton type="submit" disabled={submitting} className="w-full">
                          {submitting ? (
                            <><Spinner size={16} className="text-accent-foreground" /> Unlocking...</>
                          ) : (
                            <><Unlock size={16} /> Unlock Stage</>
                          )}
                        </MotionButton>
                      </form>
                    ) : (
                      <div className="flex items-center gap-2 rounded-lg bg-input p-3 text-sm text-text-muted">
                        <Lock size={14} />
                        Waiting for the guide to unlock this stage...
                      </div>
                    )
                  )}
                </CardContent>
              </Card>
            </motion.div>
          )}

          {currentStage && !isEnded && stagePhase === 'answering' && (
            <motion.div key="answering" {...phaseTransition}>
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <MapPin size={18} className="text-accent" />
                    <h2 className="text-lg font-semibold text-text-primary">
                      Stage {currentStage.stageNumber}
                      <span className="text-text-muted font-normal"> / {game.totalStages}</span>
                    </h2>
                  </div>
                  <p className="text-sm text-text-muted">{currentStage.location}</p>
                </CardHeader>
                <CardContent className="flex flex-col gap-4">
                  <p className="text-text-secondary">{currentStage.clue}</p>
                  {currentStage.question && (
                    <div className="rounded-lg bg-input p-3 text-sm text-text-primary font-medium">
                      {currentStage.question}
                    </div>
                  )}

                  <AnimatePresence mode="wait">
                    {feedback && !feedback.correct ? (
                      <motion.div
                        key="wrong"
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -10 }}
                        className="flex flex-col gap-3"
                      >
                        <Alert variant="error">{feedback.message}</Alert>
                        <Button
                          variant="secondary"
                          onClick={() => { setFeedback(null); setStagePhase('interstitial'); fetchState() }}
                          className="w-full"
                        >
                          <ArrowRight size={16} />
                          Continue
                        </Button>
                      </motion.div>
                    ) : canAnswer ? (
                      <motion.form
                        key="form"
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        onSubmit={handleSubmit}
                        className="flex flex-col gap-3"
                      >
                        <Input
                          type="text"
                          value={answer}
                          onChange={(e) => setAnswer(e.target.value)}
                          placeholder="Your answer..."
                          autoFocus
                          required
                        />
                        <AnimatePresence>
                          {feedback && (
                            <motion.div
                              initial={{ opacity: 0, height: 0 }}
                              animate={{ opacity: 1, height: 'auto' }}
                              exit={{ opacity: 0, height: 0 }}
                            >
                              <Alert variant="success">{feedback.message}</Alert>
                            </motion.div>
                          )}
                        </AnimatePresence>
                        <MotionButton type="submit" disabled={submitting} className="w-full">
                          {submitting ? (
                            <><Spinner size={16} className="text-accent-foreground" /> Submitting...</>
                          ) : (
                            <><Send size={16} /> Submit Answer</>
                          )}
                        </MotionButton>
                      </motion.form>
                    ) : (
                      <motion.div
                        key="waiting"
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        className="flex items-center gap-2 rounded-lg bg-input p-3 text-sm text-text-muted"
                      >
                        <UserCog size={14} />
                        Waiting for the supervisor to submit the answer...
                      </motion.div>
                    )}
                  </AnimatePresence>
                </CardContent>
              </Card>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Completed stages (non-ended state) */}
        {completedStages.length > 0 && !isEnded && (
          <Card>
            <CardContent className="pt-5">
              <details>
                <summary className="cursor-pointer text-sm font-medium text-text-secondary hover:text-text-primary transition-colors">
                  Completed Stages ({completedStages.length})
                </summary>
                <div className="mt-3 flex flex-col gap-1.5">
                  {completedStages.map((s) => (
                    <div key={s.stageNumber} className="flex items-center gap-2 text-sm">
                      {s.isCorrect ? (
                        <CheckCircle2 size={14} className="text-success" />
                      ) : (
                        <XCircle size={14} className="text-error" />
                      )}
                      <span className="text-text-secondary">
                        Stage {s.stageNumber} — {s.isCorrect ? 'correct' : 'incorrect'}
                      </span>
                    </div>
                  ))}
                </div>
              </details>
            </CardContent>
          </Card>
        )}

        {/* Team members */}
        <Card>
          <CardContent className="pt-5">
            <details>
              <summary className="cursor-pointer text-sm font-medium text-text-secondary hover:text-text-primary transition-colors">
                <Users size={14} className="inline mr-1.5 -mt-0.5" />
                Team ({players.length} players)
              </summary>
              <div className="mt-3 flex flex-col gap-1">
                {players.map((p) => (
                  <span key={p.id} className="text-sm text-text-muted pl-5">{p.name}</span>
                ))}
              </div>
            </details>
          </CardContent>
        </Card>

        {/* Leave game */}
        <div className="flex justify-center pt-2 pb-4">
          <button
            onClick={handleLogout}
            className="text-xs text-text-muted hover:text-text-secondary transition-colors flex items-center gap-1 cursor-pointer"
          >
            <LogOut size={12} />
            Leave game
          </button>
        </div>
      </div>
    </div>
  )
}
