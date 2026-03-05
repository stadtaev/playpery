import { TimerDisplay } from './TimerDisplay'
import { InterstitialPanel } from './InterstitialPanel'
import { UnlockPanel } from './UnlockPanel'
import { AnswerPanel } from './AnswerPanel'
import { ResultsPanel } from './ResultsPanel'
import { useGameState } from './useGameState'
import { PageContainer } from './components/PageContainer'
import { LoadingPage } from './components/Spinner'
import { ErrorMessage } from './components/ErrorMessage'

export function GamePage() {
  const {
    state, error, stagePhase,
    answer, setAnswer,
    unlockCode, setUnlockCode,
    feedback, answerResult, submitting,
    gameRemaining, stageRemaining,
    handleGoToStage, handleUnlock, handleSubmit, handleContinue, handleLogout,
  } = useGameState()

  if (error) {
    return (
      <PageContainer>
        <h1>CityQuest</h1>
        <ErrorMessage message={error} />
        <button onClick={handleLogout} className="btn">Back to start</button>
      </PageContainer>
    )
  }

  if (!state) {
    return <LoadingPage message="Loading game..." />
  }

  const { game, team, role, currentStage, completedStages, players } = state
  const isEnded = game.status === 'ended' || (!currentStage && completedStages.length === game.totalStages)
  const mode = game.mode || 'classic'
  const canAnswer = !game.supervised || role === 'supervisor'

  return (
    <PageContainer size="md">
      {game.timerEnabled && !isEnded && (
        <TimerDisplay gameRemaining={gameRemaining} stageRemaining={stageRemaining} />
      )}
      <nav className="flex justify-between items-center mb-6">
        <h1 className="m-0">CityQuest</h1>
        <span className="text-secondary text-sm uppercase tracking-widest font-bold">{team.name}</span>
      </nav>

      {isEnded && (
        <div className="card">
          <div className="card-header">Game Over!</div>
          <p>
            Your team answered {completedStages.filter((s) => s.isCorrect).length} of {game.totalStages} correctly.
          </p>
        </div>
      )}

      {currentStage && !isEnded && stagePhase === 'interstitial' && (
        <InterstitialPanel
          stage={currentStage}
          totalStages={game.totalStages}
          isFirstStage={completedStages.length === 0}
          role={role}
          onGoToStage={handleGoToStage}
        />
      )}

      {currentStage && !isEnded && stagePhase === 'unlocking' && (
        <UnlockPanel
          stage={currentStage}
          totalStages={game.totalStages}
          mode={mode}
          role={role}
          unlockCode={unlockCode}
          onUnlockCodeChange={setUnlockCode}
          onUnlock={handleUnlock}
          feedback={feedback}
          submitting={submitting}
          teamSecret={state.teamSecret}
        />
      )}

      {currentStage && !isEnded && stagePhase === 'answering' && (
        <AnswerPanel
          stage={currentStage}
          totalStages={game.totalStages}
          role={role}
          answer={answer}
          onAnswerChange={setAnswer}
          onSubmit={handleSubmit}
          feedback={feedback}
          submitting={submitting}
          canAnswer={canAnswer}
        />
      )}

      {currentStage && !isEnded && stagePhase === 'results' && answerResult && (
        <ResultsPanel
          stageNumber={currentStage.stageNumber}
          totalStages={game.totalStages}
          result={answerResult}
          onContinue={handleContinue}
        />
      )}

      {completedStages.length > 0 && (
        <details open={isEnded}>
          <summary>Completed Stages ({completedStages.length})</summary>
          <ul className="mt-3 space-y-1">
            {completedStages.map((s) => (
              <li key={s.stageNumber} className={s.isCorrect ? 'text-success' : 'text-error'}>
                Stage {s.stageNumber} &mdash; {s.isCorrect ? 'correct' : 'incorrect'}
              </li>
            ))}
          </ul>
        </details>
      )}

      <details>
        <summary>Team ({(() => { const n = players.filter((p) => p.role !== 'supervisor').length; return `${n} player${n !== 1 ? 's' : ''}` })()})</summary>
        <div className="mt-3 space-y-1">
          {players.filter((p) => p.role === 'supervisor').map((p) => (
            <p key={p.id} className="flex items-center gap-2">
              <span className="inline-flex items-center justify-center w-5 h-5 bg-blue-600 text-white text-xs font-bold rounded-full">i</span>
              <span>Supervisor: {p.name}</span>
            </p>
          ))}
          <ul className="space-y-1">
            {players.filter((p) => p.role !== 'supervisor').map((p) => (
              <li key={p.id}>{p.name}</li>
            ))}
          </ul>
        </div>
      </details>

      <p className="text-center mt-8">
        <a href="#" onClick={(e) => { e.preventDefault(); handleLogout() }} className="text-secondary text-xs uppercase tracking-widest">
          Leave game
        </a>
      </p>
    </PageContainer>
  )
}
