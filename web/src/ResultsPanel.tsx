import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { AnswerResult } from './useGameState'
import type { FunFact } from './types'

function normalizeFunFact(f: string | FunFact): FunFact {
  return typeof f === 'string' ? { text: f } : f
}

interface Props {
  stageNumber: number
  totalStages: number
  result: AnswerResult
  onContinue: () => void
}

export function ResultsPanel({ stageNumber, totalStages, result, onContinue }: Props) {
  const { t } = useTranslation('player')
  const [page, setPage] = useState(0)
  const funFacts = (result.funFacts ?? []).map(normalizeFunFact)
  const hasPages = funFacts.length > 0
  const isLastPage = !hasPages || page >= funFacts.length - 1
  const currentFact = hasPages ? funFacts[page] : null

  return (
    <div className="card">
      <div className="card-header">
        {t('stage_of', { current: stageNumber, total: totalStages })}
      </div>
      <div className="space-y-4">
        <p className={result.isCorrect ? 'text-feedback-success' : 'text-feedback-error'}>
          {result.isCorrect ? t('correct') : t('incorrect')}
        </p>
        <p>{t('correct_answer_is')} <strong>{result.correctAnswer}</strong></p>

        {currentFact && (
          <>
            <hr className="border-t border-gray-200" />
            <p>{currentFact.text}</p>
            {currentFact.image && (
              <img src={currentFact.image} alt="" className="w-full" />
            )}
            <p className="text-secondary text-sm text-center">{page + 1} / {funFacts.length}</p>
          </>
        )}

        {!isLastPage ? (
          <button className="btn w-full" onClick={() => setPage((p) => p + 1)}>
            {t('next')}
          </button>
        ) : (
          <button className="btn btn-accent w-full" onClick={onContinue}>
            {t('continue')}
          </button>
        )}
      </div>
    </div>
  )
}
