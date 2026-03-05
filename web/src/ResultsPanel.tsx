import { useState } from 'react'
import type { AnswerResult } from './useGameState'

interface Props {
  stageNumber: number
  totalStages: number
  result: AnswerResult
  onContinue: () => void
}

export function ResultsPanel({ stageNumber, totalStages, result, onContinue }: Props) {
  const [page, setPage] = useState(0)
  const funFacts = result.funFacts ?? []
  const hasPages = funFacts.length > 0
  const isLastPage = !hasPages || page >= funFacts.length - 1

  return (
    <div className="card">
      <div className="card-header">
        Stage {stageNumber} of {totalStages}
      </div>
      <div className="space-y-4">
        <p className={result.isCorrect ? 'text-feedback-success' : 'text-feedback-error'}>
          {result.isCorrect ? 'Correct!' : 'Incorrect'}
        </p>
        <p>The correct answer is: <strong>{result.correctAnswer}</strong></p>

        {hasPages && (
          <>
            <hr className="border-t border-gray-200" />
            <p>{funFacts[page]}</p>
            <p className="text-secondary text-sm text-center">{page + 1} / {funFacts.length}</p>
          </>
        )}

        {!isLastPage ? (
          <button className="btn w-full" onClick={() => setPage((p) => p + 1)}>
            Next
          </button>
        ) : (
          <button className="btn btn-accent w-full" onClick={onContinue}>
            Continue
          </button>
        )}
      </div>
    </div>
  )
}
