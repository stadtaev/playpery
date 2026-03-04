import type { StageInfo } from './types'

interface Props {
  stage: StageInfo
  totalStages: number
  isFirstStage: boolean
  onGoToStage: () => void
}

export function InterstitialPanel({ stage, totalStages, isFirstStage, onGoToStage }: Props) {
  return (
    <div className="card">
      <div className="card-header">
        Stage {stage.stageNumber} of {totalStages} &mdash; {stage.location}
      </div>
      <p className="mb-4"><strong>Clue:</strong> {stage.clue}</p>
      <button onClick={onGoToStage} className="btn btn-accent w-full">
        {isFirstStage ? 'Start Quest' : 'Go to Next Stage'}
      </button>
    </div>
  )
}
