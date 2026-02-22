package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Document types stored as JSONB in per-model tables.

type scenarioDoc struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	City        string       `json:"city"`
	Description string       `json:"description"`
	Stages      []AdminStage `json:"stages"`
	CreatedAt   string       `json:"createdAt"`
}

type gameDoc struct {
	ID                string       `json:"id"`
	ScenarioID        string       `json:"scenarioId"`
	ScenarioName      string       `json:"scenarioName"`
	Status            string       `json:"status"`
	TimerEnabled      bool         `json:"timerEnabled"`
	TimerMinutes      int          `json:"timerMinutes"`
	StageTimerMinutes int          `json:"stageTimerMinutes"`
	Stages            []AdminStage `json:"stages"`
	StartedAt         *string      `json:"startedAt"`
	EndedAt           *string      `json:"endedAt"`
	CreatedAt         string       `json:"createdAt"`
	Teams             []teamDoc    `json:"teams"`
}

type teamDoc struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	JoinToken string           `json:"joinToken"`
	GuideName string           `json:"guideName"`
	CreatedAt string           `json:"createdAt"`
	Players   []playerDoc      `json:"players"`
	Results   []stageResultDoc `json:"results"`
}

type playerDoc struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SessionID string `json:"sessionId"`
	JoinedAt  string `json:"joinedAt"`
}

type stageResultDoc struct {
	StageNumber int    `json:"stageNumber"`
	Answer      string `json:"answer"`
	IsCorrect   bool   `json:"isCorrect"`
	AnsweredAt  string `json:"answeredAt"`
}

type playerSessionDoc struct {
	PlayerID string `json:"playerId"`
	TeamID   string `json:"teamId"`
	GameID   string `json:"gameId"`
}

// DocStore implements Store using per-model tables with JSONB data columns.
type DocStore struct {
	db *sql.DB
}

func NewDocStore(ctx context.Context, db *sql.DB) (*DocStore, error) {
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS games (
			id          TEXT PRIMARY KEY,
			scenario_id TEXT NOT NULL,
			status      TEXT NOT NULL,
			data        JSONB NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS player_sessions (
			id   TEXT PRIMARY KEY,
			data JSONB NOT NULL
		)`,
	} {
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return nil, fmt.Errorf("creating table: %w", err)
		}
	}

	return &DocStore{db: db}, nil
}

// Generic helpers — same shape, just take table instead of collection.

func (s *DocStore) get(ctx context.Context, table, id string, dest any) error {
	var data string
	err := s.db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT json(data) FROM %s WHERE id = ?`, table), id,
	).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (s *DocStore) del(ctx context.Context, table, id string) error {
	result, err := s.db.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, table), id,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Per-table put methods — different columns per table.

func (s *DocStore) putGame(ctx context.Context, g gameDoc) error {
	data, err := json.Marshal(g)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO games (id, scenario_id, status, data) VALUES (?, ?, ?, jsonb(?))
		 ON CONFLICT(id) DO UPDATE SET scenario_id = excluded.scenario_id, status = excluded.status, data = excluded.data`,
		g.ID, g.ScenarioID, g.Status, string(data),
	)
	return err
}

func (s *DocStore) putSession(ctx context.Context, table, id string, doc any) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		fmt.Sprintf(`INSERT OR REPLACE INTO %s (id, data) VALUES (?, jsonb(?))`, table),
		id, string(data),
	)
	return err
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func nowUTC() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

// allGames loads all game documents into memory.
func (s *DocStore) allGames(ctx context.Context) ([]gameDoc, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT json(data) FROM games ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []gameDoc
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var g gameDoc
		if err := json.Unmarshal([]byte(data), &g); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

// getGame is a convenience wrapper that returns the gameDoc by ID.
// Backfills defaults for documents created before new fields existed.
func (s *DocStore) getGame(ctx context.Context, id string) (gameDoc, error) {
	var g gameDoc
	err := s.get(ctx, "games", id, &g)
	if err == nil && !g.TimerEnabled && g.TimerMinutes > 0 {
		g.TimerEnabled = true
		if g.StageTimerMinutes == 0 {
			g.StageTimerMinutes = 10
		}
	}
	return g, err
}

// modifyGame loads a game, applies fn, and saves it in a transaction.
func (s *DocStore) modifyGame(ctx context.Context, gameID string, fn func(*gameDoc) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var data string
	err = tx.QueryRowContext(ctx,
		`SELECT json(data) FROM games WHERE id = ?`, gameID,
	).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	var g gameDoc
	if err := json.Unmarshal([]byte(data), &g); err != nil {
		return err
	}

	if err := fn(&g); err != nil {
		return err
	}

	jsonData, err := json.Marshal(g)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`UPDATE games SET scenario_id = ?, status = ?, data = jsonb(?) WHERE id = ?`,
		g.ScenarioID, g.Status, string(jsonData), g.ID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Player auth

func (s *DocStore) PlayerFromToken(ctx context.Context, token string) (playerSession, error) {
	var ps playerSessionDoc
	err := s.get(ctx, "player_sessions", token, &ps)
	if errors.Is(err, ErrNotFound) {
		return playerSession{}, errNoSession
	}
	if err != nil {
		return playerSession{}, err
	}
	return playerSession{PlayerID: ps.PlayerID, TeamID: ps.TeamID, GameID: ps.GameID}, nil
}

// Player game flow

func (s *DocStore) TeamLookup(ctx context.Context, joinToken string) (TeamLookupResponse, error) {
	// Materialize active games first — SQLite can't have concurrent cursors.
	rows, err := s.db.QueryContext(ctx,
		`SELECT json(data) FROM games WHERE status = 'active'`,
	)
	if err != nil {
		return TeamLookupResponse{}, err
	}
	defer rows.Close()

	var games []gameDoc
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return TeamLookupResponse{}, err
		}
		var g gameDoc
		if err := json.Unmarshal([]byte(data), &g); err != nil {
			return TeamLookupResponse{}, err
		}
		games = append(games, g)
	}

	for _, g := range games {
		for _, t := range g.Teams {
			if t.JoinToken == joinToken {
				return TeamLookupResponse{
					ID:       t.ID,
					Name:     t.Name,
					GameName: g.ScenarioName,
					GameID:   g.ID,
				}, nil
			}
		}
	}
	return TeamLookupResponse{}, ErrNotFound
}

func (s *DocStore) JoinTeam(ctx context.Context, gameID, teamID, playerName string) (string, string, error) {
	playerID := newID()
	sessionID := newID()
	now := nowUTC()

	err := s.modifyGame(ctx, gameID, func(g *gameDoc) error {
		for i := range g.Teams {
			if g.Teams[i].ID == teamID {
				g.Teams[i].Players = append(g.Teams[i].Players, playerDoc{
					ID:        playerID,
					Name:      playerName,
					SessionID: sessionID,
					JoinedAt:  now,
				})
				return nil
			}
		}
		return ErrNotFound
	})
	if err != nil {
		return "", "", err
	}

	err = s.putSession(ctx, "player_sessions", sessionID, playerSessionDoc{
		PlayerID: playerID,
		TeamID:   teamID,
		GameID:   gameID,
	})
	if err != nil {
		return "", "", err
	}

	return playerID, sessionID, nil
}

func (s *DocStore) GameState(ctx context.Context, gameID, teamID string) (gameStateData, error) {
	g, err := s.getGame(ctx, gameID)
	if err != nil {
		return gameStateData{}, err
	}

	stagesJSON, _ := json.Marshal(g.Stages)

	var teamName string
	for _, t := range g.Teams {
		if t.ID == teamID {
			teamName = t.Name
			break
		}
	}

	var d gameStateData
	d.Status = g.Status
	d.TimerEnabled = g.TimerEnabled
	d.TimerMinutes = g.TimerMinutes
	d.StageTimerMinutes = g.StageTimerMinutes
	d.StartedAt = g.StartedAt
	d.StagesJSON = string(stagesJSON)
	d.TeamName = teamName
	return d, nil
}

func (s *DocStore) ExpireGame(ctx context.Context, gameID string) error {
	now := nowUTC()
	return s.modifyGame(ctx, gameID, func(g *gameDoc) error {
		if g.Status == "active" {
			g.Status = "ended"
			g.EndedAt = &now
		}
		return nil
	})
}

func (s *DocStore) CountCorrectAnswers(ctx context.Context, gameID, teamID string) (int, error) {
	g, err := s.getGame(ctx, gameID)
	if err != nil {
		return 0, err
	}
	for _, t := range g.Teams {
		if t.ID == teamID {
			count := 0
			for _, r := range t.Results {
				if r.IsCorrect {
					count++
				}
			}
			return count, nil
		}
	}
	return 0, nil
}

func (s *DocStore) RecordAnswer(ctx context.Context, gameID, teamID string, stageNumber int, answer string, isCorrect bool) error {
	now := nowUTC()
	return s.modifyGame(ctx, gameID, func(g *gameDoc) error {
		for i := range g.Teams {
			if g.Teams[i].ID == teamID {
				g.Teams[i].Results = append(g.Teams[i].Results, stageResultDoc{
					StageNumber: stageNumber,
					Answer:      answer,
					IsCorrect:   isCorrect,
					AnsweredAt:  now,
				})
				return nil
			}
		}
		return ErrNotFound
	})
}

func (s *DocStore) ListPlayers(ctx context.Context, gameID, teamID string) ([]PlayerInfo, error) {
	g, err := s.getGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	for _, t := range g.Teams {
		if t.ID == teamID {
			players := make([]PlayerInfo, len(t.Players))
			for i, p := range t.Players {
				players[i] = PlayerInfo{ID: p.ID, Name: p.Name}
			}
			return players, nil
		}
	}
	return nil, nil
}

func (s *DocStore) ListCompletedStages(ctx context.Context, gameID, teamID string) ([]CompletedStage, error) {
	g, err := s.getGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	for _, t := range g.Teams {
		if t.ID == teamID {
			var completed []CompletedStage
			for _, r := range t.Results {
				if r.IsCorrect {
					completed = append(completed, CompletedStage{
						StageNumber: r.StageNumber,
						IsCorrect:   true,
						AnsweredAt:  r.AnsweredAt,
					})
				}
			}
			return completed, nil
		}
	}
	return nil, nil
}

// Admin games

func (s *DocStore) ListGames(ctx context.Context) ([]AdminGameSummary, error) {
	allGames, err := s.allGames(ctx)
	if err != nil {
		return nil, err
	}

	var games []AdminGameSummary
	for _, g := range allGames {
		games = append(games, AdminGameSummary{
			ID:                g.ID,
			ScenarioID:        g.ScenarioID,
			ScenarioName:      g.ScenarioName,
			Status:            g.Status,
			TimerEnabled:      g.TimerEnabled,
			TimerMinutes:      g.TimerMinutes,
			StageTimerMinutes: g.StageTimerMinutes,
			TeamCount:         len(g.Teams),
			CreatedAt:         g.CreatedAt,
		})
	}
	// Sort newest first.
	for i, j := 0, len(games)-1; i < j; i, j = i+1, j-1 {
		games[i], games[j] = games[j], games[i]
	}
	return games, nil
}

func (s *DocStore) CreateGame(ctx context.Context, req AdminGameRequest, stages []AdminStage) (AdminGameDetail, error) {
	id := newID()
	now := nowUTC()
	doc := gameDoc{
		ID:                id,
		ScenarioID:        req.ScenarioID,
		ScenarioName:      req.ScenarioName,
		Status:            req.Status,
		TimerEnabled:      req.TimerEnabled,
		TimerMinutes:      req.TimerMinutes,
		StageTimerMinutes: req.StageTimerMinutes,
		Stages:            stages,
		CreatedAt:         now,
		Teams:             []teamDoc{},
	}
	if err := s.putGame(ctx, doc); err != nil {
		return AdminGameDetail{}, err
	}
	return AdminGameDetail{
		ID:                id,
		ScenarioID:        req.ScenarioID,
		ScenarioName:      req.ScenarioName,
		Status:            req.Status,
		TimerEnabled:      req.TimerEnabled,
		TimerMinutes:      req.TimerMinutes,
		StageTimerMinutes: req.StageTimerMinutes,
		Teams:             []AdminTeamItem{},
		CreatedAt:         now,
	}, nil
}

func (s *DocStore) GetGame(ctx context.Context, id string) (AdminGameDetail, error) {
	g, err := s.getGame(ctx, id)
	if err != nil {
		return AdminGameDetail{}, err
	}

	teams := make([]AdminTeamItem, len(g.Teams))
	for i, t := range g.Teams {
		teams[i] = AdminTeamItem{
			ID:          t.ID,
			Name:        t.Name,
			JoinToken:   t.JoinToken,
			GuideName:   t.GuideName,
			PlayerCount: len(t.Players),
			CreatedAt:   t.CreatedAt,
		}
	}

	return AdminGameDetail{
		ID:                g.ID,
		ScenarioID:        g.ScenarioID,
		ScenarioName:      g.ScenarioName,
		Status:            g.Status,
		TimerEnabled:      g.TimerEnabled,
		TimerMinutes:      g.TimerMinutes,
		StageTimerMinutes: g.StageTimerMinutes,
		Teams:             teams,
		CreatedAt:         g.CreatedAt,
	}, nil
}

func (s *DocStore) UpdateGame(ctx context.Context, id string, req AdminGameRequest) (AdminGameDetail, error) {
	g, err := s.getGame(ctx, id)
	if err != nil {
		return AdminGameDetail{}, err
	}
	g.ScenarioID = req.ScenarioID
	g.Status = req.Status
	g.TimerEnabled = req.TimerEnabled
	g.TimerMinutes = req.TimerMinutes
	g.StageTimerMinutes = req.StageTimerMinutes
	if err := s.putGame(ctx, g); err != nil {
		return AdminGameDetail{}, err
	}
	return AdminGameDetail{
		ID:                id,
		ScenarioID:        req.ScenarioID,
		Status:            req.Status,
		TimerEnabled:      req.TimerEnabled,
		TimerMinutes:      req.TimerMinutes,
		StageTimerMinutes: req.StageTimerMinutes,
		CreatedAt:         g.CreatedAt,
	}, nil
}

func (s *DocStore) DeleteGame(ctx context.Context, id string) error {
	return s.del(ctx, "games", id)
}

func (s *DocStore) GameHasPlayers(ctx context.Context, gameID string) (bool, error) {
	g, err := s.getGame(ctx, gameID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, t := range g.Teams {
		if len(t.Players) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (s *DocStore) DeleteTeamsByGame(ctx context.Context, gameID string) error {
	return s.modifyGame(ctx, gameID, func(g *gameDoc) error {
		g.Teams = []teamDoc{}
		return nil
	})
}

// Admin teams

func (s *DocStore) ListTeams(ctx context.Context, gameID string) ([]AdminTeamItem, error) {
	g, err := s.getGame(ctx, gameID)
	if errors.Is(err, ErrNotFound) {
		return []AdminTeamItem{}, nil
	}
	if err != nil {
		return nil, err
	}
	teams := make([]AdminTeamItem, len(g.Teams))
	for i, t := range g.Teams {
		teams[i] = AdminTeamItem{
			ID:          t.ID,
			Name:        t.Name,
			JoinToken:   t.JoinToken,
			GuideName:   t.GuideName,
			PlayerCount: len(t.Players),
			CreatedAt:   t.CreatedAt,
		}
	}
	return teams, nil
}

func (s *DocStore) CreateTeam(ctx context.Context, gameID string, req AdminTeamRequest, token string) (AdminTeamItem, error) {
	// Check join token uniqueness across all games.
	games, err := s.allGames(ctx)
	if err != nil {
		return AdminTeamItem{}, err
	}
	for _, g := range games {
		for _, t := range g.Teams {
			if t.JoinToken == token {
				return AdminTeamItem{}, fmt.Errorf("UNIQUE constraint failed: join_token %q", token)
			}
		}
	}

	teamID := newID()
	now := nowUTC()
	newTeam := teamDoc{
		ID:        teamID,
		Name:      req.Name,
		JoinToken: token,
		GuideName: req.GuideName,
		CreatedAt: now,
		Players:   []playerDoc{},
		Results:   []stageResultDoc{},
	}

	err = s.modifyGame(ctx, gameID, func(g *gameDoc) error {
		g.Teams = append(g.Teams, newTeam)
		return nil
	})
	if err != nil {
		return AdminTeamItem{}, err
	}

	return AdminTeamItem{
		ID:          teamID,
		Name:        req.Name,
		JoinToken:   token,
		GuideName:   req.GuideName,
		PlayerCount: 0,
		CreatedAt:   now,
	}, nil
}

func (s *DocStore) UpdateTeam(ctx context.Context, gameID, teamID string, req AdminTeamRequest) (AdminTeamItem, error) {
	var result AdminTeamItem
	err := s.modifyGame(ctx, gameID, func(g *gameDoc) error {
		for i := range g.Teams {
			if g.Teams[i].ID == teamID {
				g.Teams[i].Name = req.Name
				g.Teams[i].GuideName = req.GuideName
				result = AdminTeamItem{
					ID:          teamID,
					Name:        req.Name,
					JoinToken:   g.Teams[i].JoinToken,
					GuideName:   req.GuideName,
					PlayerCount: len(g.Teams[i].Players),
					CreatedAt:   g.Teams[i].CreatedAt,
				}
				return nil
			}
		}
		return ErrNotFound
	})
	return result, err
}

func (s *DocStore) DeleteTeam(ctx context.Context, gameID, teamID string) error {
	return s.modifyGame(ctx, gameID, func(g *gameDoc) error {
		for i := range g.Teams {
			if g.Teams[i].ID == teamID {
				g.Teams = append(g.Teams[:i], g.Teams[i+1:]...)
				return nil
			}
		}
		return ErrNotFound
	})
}

func (s *DocStore) TeamHasPlayers(ctx context.Context, gameID, teamID string) (bool, error) {
	g, err := s.getGame(ctx, gameID)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, t := range g.Teams {
		if t.ID == teamID {
			return len(t.Players) > 0, nil
		}
	}
	return false, nil
}

func (s *DocStore) GameExists(ctx context.Context, gameID string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM games WHERE id = ?`, gameID,
	).Scan(&n)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *DocStore) GameStatus(ctx context.Context, gameID string) (AdminGameStatus, error) {
	g, err := s.getGame(ctx, gameID)
	if err != nil {
		return AdminGameStatus{}, err
	}

	teams := make([]AdminTeamStatus, len(g.Teams))
	for i, t := range g.Teams {
		players := make([]AdminPlayerStatus, len(t.Players))
		for j, p := range t.Players {
			players[j] = AdminPlayerStatus{
				Name:     p.Name,
				JoinedAt: p.JoinedAt,
			}
		}

		completed := 0
		for _, r := range t.Results {
			if r.IsCorrect {
				completed++
			}
		}

		teams[i] = AdminTeamStatus{
			ID:              t.ID,
			Name:            t.Name,
			GuideName:       t.GuideName,
			CompletedStages: completed,
			Players:         players,
		}
	}

	return AdminGameStatus{
		ID:                g.ID,
		ScenarioName:      g.ScenarioName,
		Status:            g.Status,
		TimerEnabled:      g.TimerEnabled,
		TimerMinutes:      g.TimerMinutes,
		StageTimerMinutes: g.StageTimerMinutes,
		StartedAt:         g.StartedAt,
		TotalStages:       len(g.Stages),
		Teams:             teams,
	}, nil
}

// SeedDemoGame creates the demo game if no games exist, snapshotting the given scenario stages.
func (s *DocStore) SeedDemoGame(ctx context.Context, sc *scenarioDoc) error {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM games`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := nowUTC()
	game := gameDoc{
		ID:                "g0000000deadbeef",
		ScenarioID:        sc.ID,
		ScenarioName:      sc.Name,
		Status:            "active",
		TimerEnabled:      true,
		TimerMinutes:      120,
		StageTimerMinutes: 10,
		Stages:       sc.Stages,
		StartedAt:    &now,
		CreatedAt:    now,
		Teams: []teamDoc{
			{
				ID:        "t000000000incas",
				Name:      "Los Incas",
				JoinToken: "incas-2025",
				CreatedAt: now,
				Players:   []playerDoc{},
				Results:   []stageResultDoc{},
			},
			{
				ID:        "t00000000condor",
				Name:      "Los Condores",
				JoinToken: "condores-2025",
				CreatedAt: now,
				Players:   []playerDoc{},
				Results:   []stageResultDoc{},
			},
		},
	}
	return s.putGame(ctx, game)
}

// Ensure DocStore implements Store at compile time.
var _ Store = (*DocStore)(nil)
