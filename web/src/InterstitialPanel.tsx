import type { StageInfo } from './types'

interface Props {
  stage: StageInfo
  totalStages: number
  isFirstStage: boolean
  onGoToStage: () => void
}

export function InterstitialPanel({ stage, totalStages, isFirstStage, onGoToStage }: Props) {
  return (
    <article>
      <header>
        Stage {stage.stageNumber} of {totalStages} &mdash; {stage.location}
      </header>
      <p><strong>Clue:</strong> {stage.clue}</p>
      <button onClick={onGoToStage}>
        {isFirstStage ? 'Start Quest' : 'Go to Next Stage'}
      </button>
    </article>
  )
}
