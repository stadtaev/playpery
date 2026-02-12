// Package cityquiz defines the core domain types and service interfaces.
// It has zero external dependencies — everything here is pure Go.
package cityquiz

import "time"

type Client struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}

type Scenario struct {
	ID          string
	Name        string
	City        string
	Description string
	Stages      []Stage
	CreatedAt   time.Time
}

type Stage struct {
	StageNumber   int
	Location      string
	Clue          string
	Question      string
	CorrectAnswer string
	Lat           float64
	Lng           float64
}

type Game struct {
	ID           string
	ScenarioID   string
	ClientID     string
	Status       GameStatus
	ScheduledAt  *time.Time
	StartedAt    *time.Time
	EndedAt      *time.Time
	TimerMinutes int
	CreatedAt    time.Time
}

type GameStatus string

const (
	GameStatusDraft  GameStatus = "draft"
	GameStatusActive GameStatus = "active"
	GameStatusPaused GameStatus = "paused"
	GameStatusEnded  GameStatus = "ended"
)

type Team struct {
	ID        string
	GameID    string
	Name      string
	JoinToken string
	GuideName string
	CreatedAt time.Time
}

type Player struct {
	ID        string
	TeamID    string
	Name      string
	SessionID string
	JoinedAt  time.Time
}

type StageResult struct {
	ID          string
	GameID      string
	TeamID      string
	StageNumber int
	Answer      string
	IsCorrect   bool
	AnsweredAt  *time.Time
}
