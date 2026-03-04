import { TimerDisplay } from './TimerDisplay'
import { InterstitialPanel } from './InterstitialPanel'
import { UnlockPanel } from './UnlockPanel'
import { AnswerPanel } from './AnswerPanel'
import { useGameState } from './useGameState'

export function GamePage() {
  const {
    state, error, stagePhase,
    answer, setAnswer,
    unlockCode, setUnlockCode,
    feedback, submitting,
    gameRemaining, stageRemaining,
    handleGoToStage, handleUnlock, handleSubmit, handleContinue, handleLogout,
  } = useGameState()

  if (error) {
    return (
      <main className="container">
        <h1>CityQuest</h1>
        <p role="alert">{error}</p>
        <button onClick={handleLogout}>Back to start</button>
      </main>
    )
  }

  if (!state) {
    return (
      <main className="container">
        <p aria-busy="true">Loading game...</p>
      </main>
    )
  }

  const { game, team, role, currentStage, completedStages, players } = state
  const isEnded = game.status === 'ended' || (!currentStage && completedStages.length === game.totalStages)
  const mode = game.mode || 'classic'
  const canAnswer = !game.supervised || role === 'supervisor' || mode === 'supervised'

  return (
    <main className="container" style={{ maxWidth: 600 }}>
      {game.timerEnabled && !isEnded && (
        <TimerDisplay gameRemaining={gameRemaining} stageRemaining={stageRemaining} />
      )}
      <nav style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 style={{ margin: 0 }}>CityQuest</h1>
        <small>{team.name}</small>
      </nav>

      {isEnded && (
        <article>
          <header>Game Over!</header>
          <p>
            Your team answered {completedStages.filter((s) => s.isCorrect).length} of {game.totalStages} correctly.
          </p>
        </article>
      )}

      {currentStage && !isEnded && stagePhase === 'interstitial' && (
        <InterstitialPanel
          stage={currentStage}
          totalStages={game.totalStages}
          isFirstStage={completedStages.length === 0}
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
          answer={answer}
          onAnswerChange={setAnswer}
          onSubmit={handleSubmit}
          onContinue={handleContinue}
          feedback={feedback}
          submitting={submitting}
          canAnswer={canAnswer}
        />
      )}

      {completedStages.length > 0 && (
        <details open={isEnded}>
          <summary>Completed Stages ({completedStages.length})</summary>
          <ul>
            {completedStages.map((s) => (
              <li key={s.stageNumber} style={{ color: s.isCorrect ? 'var(--pico-color-green-500)' : 'var(--pico-color-red-500)' }}>
                Stage {s.stageNumber} &mdash; {s.isCorrect ? 'correct' : 'incorrect'}
              </li>
            ))}
          </ul>
        </details>
      )}

      <details>
        <summary>Team ({players.length} players)</summary>
        <ul>
          {players.map((p) => (
            <li key={p.id}>{p.name}</li>
          ))}
        </ul>
      </details>

      <p style={{ textAlign: 'center', marginTop: '2rem' }}>
        <a href="#" onClick={(e) => { e.preventDefault(); handleLogout() }} style={{ fontSize: 'small' }}>
          Leave game
        </a>
      </p>
    </main>
  )
}
