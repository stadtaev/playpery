import type { Feedback } from './useGameState'

export function FeedbackMessage({ feedback }: { feedback: Feedback }) {
  return (
    <small style={{ color: feedback.correct ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
      {feedback.message}
    </small>
  )
}
