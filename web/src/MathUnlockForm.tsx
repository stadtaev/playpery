import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'

interface Props {
  unlockCode: string
  onUnlockCodeChange: (code: string) => void
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
  teamSecret?: number
  locationNumber?: number
}

export function MathUnlockForm({ unlockCode, onUnlockCodeChange, onUnlock, feedback, submitting, teamSecret, locationNumber }: Props) {
  return (
    <form onSubmit={onUnlock}>
      <p>Your team secret is: <strong>{teamSecret}</strong></p>
      <p>Location number: <strong>{locationNumber}</strong></p>
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
  )
}
