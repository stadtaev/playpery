import { useTranslation } from 'react-i18next'
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
  const { t } = useTranslation('player')
  return (
    <div className="card">
      <div className="card-header">
        {t('stage_of', { current: stage.stageNumber, total: totalStages })}{role === 'supervisor' && <> &mdash; {stage.location}</>}
      </div>
      {stage.question && (
        <div className="mb-4">
          <p><strong>{t('question_label')}</strong> {stage.question}</p>
          {stage.questionImage && <img src={stage.questionImage} alt="" className="w-full mt-2" />}
        </div>
      )}
      {canAnswer ? (
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <input
              className="input"
              type="text"
              value={answer}
              onChange={(e) => onAnswerChange(e.target.value)}
              placeholder={t('answer_placeholder')}
              autoFocus
              required
            />
          </div>
          {feedback && (
            <p className={feedback.correct ? 'text-feedback-success' : 'text-feedback-error'}>{feedback.message}</p>
          )}
          <button type="submit" disabled={submitting} className="btn btn-accent w-full">
            {submitting ? <Spinner /> : t('submit_answer')}
          </button>
        </form>
      ) : (
        <p className="text-secondary italic">{t('waiting_for_supervisor_answer')}</p>
      )}
    </div>
  )
}
