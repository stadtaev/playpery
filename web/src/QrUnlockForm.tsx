import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'
import { Spinner } from './components/Spinner'

interface Props {
  unlockCode: string
  onUnlockCodeChange: (code: string) => void
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
}

export function QrUnlockForm({ unlockCode, onUnlockCodeChange, onUnlock, feedback, submitting }: Props) {
  return (
    <form onSubmit={onUnlock} className="space-y-4">
      <p>Enter the code from the QR at this location:</p>
      <div>
        <input
          className="input"
          type="text"
          value={unlockCode}
          onChange={(e) => onUnlockCodeChange(e.target.value)}
          placeholder="QR code..."
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
