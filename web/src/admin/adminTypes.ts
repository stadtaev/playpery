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
