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
  timerMinutes: number
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
  timerMinutes: number
  teams: TeamItem[]
  createdAt: string
}

export interface GameRequest {
  scenarioId: string
  status: string
  timerMinutes: number
}

export interface TeamRequest {
  name: string
  joinToken: string
  guideName: string
}
