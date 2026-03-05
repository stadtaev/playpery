import type { StageInfo } from './types'
import type { Feedback } from './useGameState'
import { Spinner } from './components/Spinner'

interface Props {
  stage: StageInfo
  totalStages: number
  role: string
  answer: string
  onAnswerChange: (answer: string) => void
  onSubmit: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
  canAnswer: boolean
}

export function AnswerPanel({ stage, totalStages, role, answer, onAnswerChange, onSubmit, feedback, submitting, canAnswer }: Props) {
  return (
    <div className="card">
      <div className="card-header">
        Stage {stage.stageNumber} of {totalStages}{role === 'supervisor' && <> &mdash; {stage.location}</>}
      </div>
      {stage.question && (
        <p className="mb-4"><strong>Question:</strong> {stage.question}</p>
      )}
      {canAnswer ? (
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <input
              className="input"
              type="text"
              value={answer}
              onChange={(e) => onAnswerChange(e.target.value)}
              placeholder="Your answer..."
              autoFocus
              required
            />
          </div>
          {feedback && (
            <p className={feedback.correct ? 'text-feedback-success' : 'text-feedback-error'}>{feedback.message}</p>
          )}
          <button type="submit" disabled={submitting} className="btn btn-accent w-full">
            {submitting ? <Spinner /> : 'Submit Answer'}
          </button>
        </form>
      ) : (
        <p className="text-secondary italic">Waiting for the supervisor to submit the answer...</p>
      )}
    </div>
  )
}
