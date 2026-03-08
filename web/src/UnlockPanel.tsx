import { useTranslation } from 'react-i18next'
import type { StageInfo, ScenarioMode } from './types'
import type { Feedback } from './useGameState'
import { QrUnlockForm } from './QrUnlockForm'
import { MathUnlockForm } from './MathUnlockForm'
import { SupervisedUnlockForm } from './SupervisedUnlockForm'

interface Props {
  stage: StageInfo
  totalStages: number
  mode: ScenarioMode
  role: string
  unlockCode: string
  onUnlockCodeChange: (code: string) => void
  onUnlock: (e: React.FormEvent) => void
  feedback: Feedback | null
  submitting: boolean
  teamSecret?: number
}

export function UnlockPanel({ stage, totalStages, mode, role, unlockCode, onUnlockCodeChange, onUnlock, feedback, submitting, teamSecret }: Props) {
  const { t } = useTranslation('player')
  const common = { onUnlock, feedback, submitting }

  return (
    <div className="card">
      <div className="card-header">
        {t('stage_of', { current: stage.stageNumber, total: totalStages })}{role === 'supervisor' && <> &mdash; {stage.location}</>}
      </div>
      {mode !== 'supervised' && (
        <p className="mb-4"><strong>{t('clue_label')}</strong> {stage.clue}</p>
      )}

      {(mode === 'qr_quiz' || mode === 'qr_hunt') && (
        <QrUnlockForm {...common} unlockCode={unlockCode} onUnlockCodeChange={onUnlockCodeChange} />
      )}
      {mode === 'math_puzzle' && (
        <MathUnlockForm {...common} unlockCode={unlockCode} onUnlockCodeChange={onUnlockCodeChange} teamSecret={teamSecret} locationNumber={stage.locationNumber} />
      )}
      {mode === 'supervised' && (
        <SupervisedUnlockForm {...common} role={role} clue={stage.clue} />
      )}
    </div>
  )
}
