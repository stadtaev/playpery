import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'

interface Props {
  role: string
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
}

export function SupervisedUnlockForm({ role, onUnlock, feedback, submitting }: Props) {
  if (role !== 'supervisor') {
    return <p><em>Waiting for the guide to unlock this stage...</em></p>
  }
  return (
    <form onSubmit={onUnlock}>
      <p>Unlock this stage for your team:</p>
      {feedback && <FeedbackMessage feedback={feedback} />}
      <button type="submit" disabled={submitting} aria-busy={submitting}>
        Unlock Stage
      </button>
    </form>
  )
}
