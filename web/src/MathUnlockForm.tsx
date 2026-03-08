import { useTranslation, Trans } from 'react-i18next'
import type { Feedback } from './useGameState'
import { FeedbackMessage } from './FeedbackMessage'
import { Spinner } from './components/Spinner'

interface Props {
  unlockCode: string
  onUnlockCodeChange: (code: string) => void
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
  teamSecret?: number
  locationNumber?: number
}

export function MathUnlockForm({ unlockCode, onUnlockCodeChange, onUnlock, feedback, submitting, teamSecret, locationNumber }: Props) {
  const { t } = useTranslation('player')
  return (
    <form onSubmit={onUnlock} className="space-y-4">
      <p><Trans i18nKey="math_team_secret" ns="player" values={{ number: teamSecret }} /></p>
      <p><Trans i18nKey="math_location_number" ns="player" values={{ number: locationNumber }} /></p>
      <p>{t('math_instruction')}</p>
      <div>
        <input
          className="input"
          type="text"
          value={unlockCode}
          onChange={(e) => onUnlockCodeChange(e.target.value)}
          placeholder={t('math_placeholder')}
          autoFocus
          required
        />
      </div>
      {feedback && <FeedbackMessage feedback={feedback} />}
      <button type="submit" disabled={submitting} className="btn w-full">
        {submitting ? <Spinner /> : t('submit_code')}
      </button>
    </form>
  )
}
