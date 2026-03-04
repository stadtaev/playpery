import type { Feedback } from './useGameState'

export function FeedbackMessage({ feedback }: { feedback: Feedback }) {
  return (
    <p className={feedback.correct ? 'text-feedback-success' : 'text-feedback-error'}>
      {feedback.message}
    </p>
  )
}
