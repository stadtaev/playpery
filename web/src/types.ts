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

export type ScenarioMode = 'classic' | 'qr_quiz' | 'qr_hunt' | 'math_puzzle' | 'guided'

export interface GameInfo {
  status: string
  mode: ScenarioMode
  hasQuestions?: boolean
  supervised: boolean
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
  question?: string
  location: string
  locked: boolean
  locationNumber?: number
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
  teamSecret?: number
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

export interface UnlockResponse {
  stageNumber: number
  unlocked: boolean
  stageComplete?: boolean
  nextStage?: StageInfo
  gameComplete?: boolean
  question?: string
}

export interface SSEEvent {
  type: 'stage_completed' | 'stage_unlocked' | 'wrong_answer' | 'player_joined' | 'game_ended'
  stageNumber?: number
  playerName?: string
}
