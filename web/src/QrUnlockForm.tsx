import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'

interface Props {
  unlockCode: string
  onUnlockCodeChange: (code: string) => void
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
}

export function QrUnlockForm({ unlockCode, onUnlockCodeChange, onUnlock, feedback, submitting }: Props) {
  return (
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
  )
}
