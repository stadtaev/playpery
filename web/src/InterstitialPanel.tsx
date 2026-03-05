import type { StageInfo } from './types'

interface Props {
  stage: StageInfo
  totalStages: number
  isFirstStage: boolean
  role: string
  onGoToStage: () => void
}

export function InterstitialPanel({ stage, totalStages, isFirstStage, role, onGoToStage }: Props) {
  return (
    <div className="card">
      <div className="card-header">
        Stage {stage.stageNumber} of {totalStages}{role === 'supervisor' && <> &mdash; {stage.location}</>}
      </div>
      <p className="mb-4"><strong>Clue:</strong> {stage.clue}</p>
      <button onClick={onGoToStage} className="btn btn-accent w-full">
        {isFirstStage ? 'Start Quest' : 'Go to Next Stage'}
      </button>
    </div>
  )
}
