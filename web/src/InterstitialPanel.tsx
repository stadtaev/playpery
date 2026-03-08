import { useTranslation } from 'react-i18next'
import type { StageInfo } from './types'

interface Props {
  stage: StageInfo
  totalStages: number
  isFirstStage: boolean
  role: string
  onGoToStage: () => void
}

export function InterstitialPanel({ stage, totalStages, isFirstStage, role, onGoToStage }: Props) {
  const { t } = useTranslation('player')
  return (
    <div className="card">
      <div className="card-header">
        {t('stage_of', { current: stage.stageNumber, total: totalStages })}{role === 'supervisor' && <> &mdash; {stage.location}</>}
      </div>
      <button onClick={onGoToStage} className="btn btn-accent w-full">
        {isFirstStage ? t('start_quest') : t('go_to_next_stage')}
      </button>
    </div>
  )
}
