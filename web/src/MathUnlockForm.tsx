import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'
import { Spinner } from './components/Spinner'

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
    <form onSubmit={onUnlock} className="space-y-4">
      <p>Your team secret is: <strong>{teamSecret}</strong></p>
      <p>Location number: <strong>{locationNumber}</strong></p>
      <p>Add them together and enter the result:</p>
      <div>
        <input
          className="input"
          type="text"
          value={unlockCode}
          onChange={(e) => onUnlockCodeChange(e.target.value)}
          placeholder="Calculated code..."
          autoFocus
          required
        />
      </div>
      {feedback && <FeedbackMessage feedback={feedback} />}
      <button type="submit" disabled={submitting} className="btn w-full">
        {submitting ? <Spinner /> : 'Submit Code'}
      </button>
    </form>
  )
}
