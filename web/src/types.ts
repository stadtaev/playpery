export interface TeamLookup {
  id: string
  name: string
  gameName: string
  role: string
}

export interface JoinResponse {
  token: string
  playerId: string
  teamId: string
  teamName: string
  role: string
}

export interface GameInfo {
  status: string
  timerEnabled: boolean
  timerMinutes: number
  stageTimerMinutes: number
  startedAt: string | null
  totalStages: number
}

export interface TeamInfo {
  id: string
  name: string
}

export interface StageInfo {
  stageNumber: number
  clue: string
  question: string
  location: string
}

export interface CompletedStage {
  stageNumber: number
  isCorrect: boolean
  answeredAt: string
}

export interface PlayerInfo {
  id: string
  name: string
}

export interface GameState {
  game: GameInfo
  team: TeamInfo
  role: string
  currentStage: StageInfo | null
  completedStages: CompletedStage[]
  players: PlayerInfo[]
}

export interface AnswerResponse {
  isCorrect: boolean
  stageNumber: number
  nextStage: StageInfo | null
  gameComplete: boolean
  correctAnswer?: string
}

export interface SSEEvent {
  type: 'stage_completed' | 'wrong_answer' | 'player_joined' | 'game_ended'
  stageNumber?: number
  playerName?: string
}
