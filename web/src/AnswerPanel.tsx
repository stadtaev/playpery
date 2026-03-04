import type { StageInfo } from './types'
import type { Feedback } from './useGameState'

interface Props {
  stage: StageInfo
  totalStages: number
  answer: string
  onAnswerChange: (answer: string) => void
  onSubmit: (e: React.FormEvent) => void
  onContinue: () => void
  feedback: Feedback | null
  submitting: boolean
  canAnswer: boolean
}

export function AnswerPanel({ stage, totalStages, answer, onAnswerChange, onSubmit, onContinue, feedback, submitting, canAnswer }: Props) {
  return (
    <article>
      <header>
        Stage {stage.stageNumber} of {totalStages} &mdash; {stage.location}
      </header>
      <p><strong>Clue:</strong> {stage.clue}</p>
      {stage.question && (
        <p><strong>Question:</strong> {stage.question}</p>
      )}
      {feedback && !feedback.correct ? (
        <>
          <p style={{ color: 'var(--pico-color-red-500)' }}>
            {feedback.message}
          </p>
          <button onClick={onContinue}>
            Continue
          </button>
        </>
      ) : canAnswer ? (
        <form onSubmit={onSubmit}>
          <input
            type="text"
            value={answer}
            onChange={(e) => onAnswerChange(e.target.value)}
            placeholder="Your answer..."
            autoFocus
            required
          />
          {feedback && (
            <small style={{ color: 'var(--pico-color-green-500)' }}>
              {feedback.message}
            </small>
          )}
          <button type="submit" disabled={submitting} aria-busy={submitting}>
            Submit Answer
          </button>
        </form>
      ) : (
        <p><em>Waiting for the supervisor to submit the answer...</em></p>
      )}
    </article>
  )
}
