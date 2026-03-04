import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'
import { Spinner } from './components/Spinner'

interface Props {
  role: string
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
}

export function SupervisedUnlockForm({ role, onUnlock, feedback, submitting }: Props) {
  if (role !== 'supervisor') {
    return <p className="text-secondary italic">Waiting for the guide to unlock this stage...</p>
  }
  return (
    <form onSubmit={onUnlock} className="space-y-4">
      <p>Unlock this stage for your team:</p>
      {feedback && <FeedbackMessage feedback={feedback} />}
      <button type="submit" disabled={submitting} className="btn w-full">
        {submitting ? <Spinner /> : 'Unlock Stage'}
      </button>
    </form>
  )
}
