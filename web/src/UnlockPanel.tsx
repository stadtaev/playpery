import type { StageInfo, ScenarioMode } from './types'
import type { Feedback } from './useGameState'

interface Props {
  stage: StageInfo
  totalStages: number
  mode: ScenarioMode
  role: string
  unlockCode: string
  onUnlockCodeChange: (code: string) => void
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
  teamSecret?: number
}

function FeedbackMessage({ feedback }: { feedback: Feedback }) {
  return (
    <small style={{ color: feedback.correct ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
      {feedback.message}
    </small>
  )
}

export function UnlockPanel({ stage, totalStages, mode, role, unlockCode, onUnlockCodeChange, onUnlock, feedback, submitting, teamSecret }: Props) {
  return (
    <article>
      <header>
        Stage {stage.stageNumber} of {totalStages} &mdash; {stage.location}
      </header>
      <p><strong>Clue:</strong> {stage.clue}</p>

      {(mode === 'qr_quiz' || mode === 'qr_hunt') && (
        <form onSubmit={onUnlock}>
          <p>Enter the code from the QR at this location:</p>
          <input
            type="text"
            value={unlockCode}
            onChange={(e) => onUnlockCodeChange(e.target.value)}
            placeholder="QR code..."
            autoFocus
            required
          />
          {feedback && <FeedbackMessage feedback={feedback} />}
          <button type="submit" disabled={submitting} aria-busy={submitting}>
            Submit Code
          </button>
        </form>
      )}

      {mode === 'math_puzzle' && (
        <form onSubmit={onUnlock}>
          <p>Your team secret is: <strong>{teamSecret}</strong></p>
          <p>Location number: <strong>{stage.locationNumber}</strong></p>
          <p>Add them together and enter the result:</p>
          <input
            type="text"
            value={unlockCode}
            onChange={(e) => onUnlockCodeChange(e.target.value)}
            placeholder="Calculated code..."
            autoFocus
            required
          />
          {feedback && <FeedbackMessage feedback={feedback} />}
          <button type="submit" disabled={submitting} aria-busy={submitting}>
            Submit Code
          </button>
        </form>
      )}

      {mode === 'supervised' && (
        role === 'supervisor' ? (
          <form onSubmit={onUnlock}>
            <p>Unlock this stage for your team:</p>
            {feedback && <FeedbackMessage feedback={feedback} />}
            <button type="submit" disabled={submitting} aria-busy={submitting}>
              Unlock Stage
            </button>
          </form>
        ) : (
          <p><em>Waiting for the guide to unlock this stage...</em></p>
        )
      )}
    </article>
  )
}
