export interface AdminMe {
  id: string
  email: string
}

export interface ScenarioSummary {
  id: string
  name: string
  city: string
  description: string
  stageCount: number
  createdAt: string
}

export interface Stage {
  stageNumber: number
  location: string
  clue: string
  question: string
  correctAnswer: string
  lat: number
  lng: number
}

export interface ScenarioDetail {
  id: string
  name: string
  city: string
  description: string
  stages: Stage[]
  createdAt: string
}

export interface ScenarioRequest {
  name: string
  city: string
  description: string
  stages: Stage[]
}

export interface GameSummary {
  id: string
  scenarioId: string
  scenarioName: string
  status: string
  timerEnabled: boolean
  timerMinutes: number
  stageTimerMinutes: number
  teamCount: number
  createdAt: string
}

export interface TeamItem {
  id: string
  name: string
  joinToken: string
  guideName: string
  playerCount: number
  createdAt: string
}

export interface GameDetail {
  id: string
  scenarioId: string
  scenarioName: string
  status: string
  timerEnabled: boolean
  timerMinutes: number
  stageTimerMinutes: number
  teams: TeamItem[]
  createdAt: string
}

export interface GameRequest {
  scenarioId: string
  status: string
  timerEnabled: boolean
  timerMinutes: number
  stageTimerMinutes: number
}

export interface TeamRequest {
  name: string
  joinToken: string
  guideName: string
}

export interface GameStatus {
  id: string
  scenarioName: string
  status: string
  timerEnabled: boolean
  timerMinutes: number
  stageTimerMinutes: number
  startedAt: string | null
  totalStages: number
  teams: TeamStatus[]
}

export interface TeamStatus {
  id: string
  name: string
  guideName: string
  completedStages: number
  players: PlayerStatus[]
}

export interface PlayerStatus {
  name: string
  joinedAt: string
}
