import { useTranslation } from 'react-i18next'
import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'
import { Spinner } from './components/Spinner'

interface Props {
  role: string
  clue: string
  clueImage?: string
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
}

export function SupervisedUnlockForm({ role, clue, clueImage, onUnlock, feedback, submitting }: Props) {
  const { t } = useTranslation('player')
  if (role !== 'supervisor') {
    return (
      <div className="space-y-4">
        <p><strong>{t('clue_label')}</strong> {clue}</p>
        {clueImage && <img src={clueImage} alt="" className="w-full" />}
        <p className="text-secondary italic">{t('supervised_waiting')}</p>
      </div>
    )
  }
  return (
    <form onSubmit={onUnlock} className="space-y-4">
      <p><strong>{t('clue_label')}</strong> {clue}</p>
      {clueImage && <img src={clueImage} alt="" className="w-full" />}
      {feedback && <FeedbackMessage feedback={feedback} />}
      <button type="submit" disabled={submitting} className="btn w-full">
        {submitting ? <Spinner /> : t('unlock_stage')}
      </button>
    </form>
  )
}
