package server

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type sessionInfo struct {
	PlayerID string
	TeamID   string
	GameID   string
	Role     string
}

type gameStateData struct {
	Status            string
	Supervised        bool
	TimerEnabled      bool
	TimerMinutes      int
	StageTimerMinutes int
	StartedAt         *string
	StagesJSON        string
	TeamName          string
}

type Store interface {
	PlayerFromToken(ctx context.Context, token string) (sessionInfo, error)

	TeamLookup(ctx context.Context, joinToken string) (TeamLookupResponse, error)
	JoinTeam(ctx context.Context, gameID, teamID, playerName, role string) (playerID, sessionID string, err error)
	GameState(ctx context.Context, gameID, teamID string) (gameStateData, error)
	ExpireGame(ctx context.Context, gameID string) error
	CountAnsweredStages(ctx context.Context, gameID, teamID string) (int, error)
	CountCorrectAnswers(ctx context.Context, gameID, teamID string) (int, error)
	RecordAnswer(ctx context.Context, gameID, teamID string, stageNumber int, answer string, isCorrect bool) error
	ListPlayers(ctx context.Context, gameID, teamID string) ([]PlayerInfo, error)
	ListCompletedStages(ctx context.Context, gameID, teamID string) ([]CompletedStage, error)

	ListGames(ctx context.Context) ([]AdminGameSummary, error)
	CreateGame(ctx context.Context, req AdminGameRequest, stages []AdminStage) (AdminGameDetail, error)
	GetGame(ctx context.Context, id string) (AdminGameDetail, error)
	UpdateGame(ctx context.Context, id string, req AdminGameRequest) (AdminGameDetail, error)
	DeleteGame(ctx context.Context, id string) error
	GameHasPlayers(ctx context.Context, gameID string) (bool, error)
	DeleteTeamsByGame(ctx context.Context, gameID string) error

	ListTeams(ctx context.Context, gameID string) ([]AdminTeamItem, error)
	CreateTeam(ctx context.Context, gameID string, req AdminTeamRequest, token string) (AdminTeamItem, error)
	UpdateTeam(ctx context.Context, gameID, teamID string, req AdminTeamRequest) (AdminTeamItem, error)
	DeleteTeam(ctx context.Context, gameID, teamID string) error
	TeamHasPlayers(ctx context.Context, gameID, teamID string) (bool, error)
	GameExists(ctx context.Context, gameID string) (bool, error)
	GameStatus(ctx context.Context, gameID string) (AdminGameStatus, error)
}
