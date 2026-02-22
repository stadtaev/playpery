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
	ID           string    `json:"id"`
	ScenarioID   string    `json:"scenarioId"`
	Status       string    `json:"status"`
	TimerMinutes int       `json:"timerMinutes"`
	StartedAt    *string   `json:"startedAt"`
	EndedAt      *string   `json:"endedAt"`
	CreatedAt    string    `json:"createdAt"`
	Teams        []teamDoc `json:"teams"`
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
		`CREATE TABLE IF NOT EXISTS scenarios (
			id   TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			data JSONB NOT NULL
		)`,
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

func (s *DocStore) putScenario(ctx context.Context, sc scenarioDoc) error {
	data, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO scenarios (id, name, data) VALUES (?, ?, jsonb(?))
		 ON CONFLICT(id) DO UPDATE SET name = excluded.name, data = excluded.data`,
		sc.ID, sc.Name, string(data),
	)
	return err
}

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

// allScenarios loads all scenario documents into memory.
func (s *DocStore) allScenarios(ctx context.Context) ([]scenarioDoc, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT json(data) FROM scenarios ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []scenarioDoc
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var sc scenarioDoc
		if err := json.Unmarshal([]byte(data), &sc); err != nil {
			return nil, err
		}
		scenarios = append(scenarios, sc)
	}
	return scenarios, nil
}

// getGame is a convenience wrapper that returns the gameDoc by ID.
func (s *DocStore) getGame(ctx context.Context, id string) (gameDoc, error) {
	var g gameDoc
	err := s.get(ctx, "games", id, &g)
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
				var sc scenarioDoc
				if err := s.get(ctx, "scenarios", g.ScenarioID, &sc); err != nil {
					return TeamLookupResponse{}, err
				}
				return TeamLookupResponse{
					ID:       t.ID,
					Name:     t.Name,
					GameName: sc.Name,
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

	var sc scenarioDoc
	if err := s.get(ctx, "scenarios", g.ScenarioID, &sc); err != nil {
		return gameStateData{}, err
	}

	stagesJSON, _ := json.Marshal(sc.Stages)

	var teamName string
	for _, t := range g.Teams {
		if t.ID == teamID {
			teamName = t.Name
			break
		}
	}

	var d gameStateData
	d.Status = g.Status
	d.TimerMinutes = g.TimerMinutes
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

// Admin scenarios

func (s *DocStore) ListScenarios(ctx context.Context) ([]AdminScenarioSummary, error) {
	all, err := s.allScenarios(ctx)
	if err != nil {
		return nil, err
	}
	var scenarios []AdminScenarioSummary
	for _, sc := range all {
		scenarios = append(scenarios, AdminScenarioSummary{
			ID:          sc.ID,
			Name:        sc.Name,
			City:        sc.City,
			Description: sc.Description,
			StageCount:  len(sc.Stages),
			CreatedAt:   sc.CreatedAt,
		})
	}
	// Sort newest first.
	for i, j := 0, len(scenarios)-1; i < j; i, j = i+1, j-1 {
		scenarios[i], scenarios[j] = scenarios[j], scenarios[i]
	}
	return scenarios, nil
}

func (s *DocStore) CreateScenario(ctx context.Context, req AdminScenarioRequest) (AdminScenarioDetail, error) {
	id := newID()
	now := nowUTC()
	doc := scenarioDoc{
		ID:          id,
		Name:        req.Name,
		City:        req.City,
		Description: req.Description,
		Stages:      req.Stages,
		CreatedAt:   now,
	}
	if err := s.putScenario(ctx, doc); err != nil {
		return AdminScenarioDetail{}, err
	}
	return AdminScenarioDetail{
		ID:          id,
		Name:        req.Name,
		City:        req.City,
		Description: req.Description,
		Stages:      req.Stages,
		CreatedAt:   now,
	}, nil
}

func (s *DocStore) GetScenario(ctx context.Context, id string) (AdminScenarioDetail, error) {
	var sc scenarioDoc
	if err := s.get(ctx, "scenarios", id, &sc); err != nil {
		return AdminScenarioDetail{}, err
	}
	stages := sc.Stages
	if stages == nil {
		stages = []AdminStage{}
	}
	return AdminScenarioDetail{
		ID:          sc.ID,
		Name:        sc.Name,
		City:        sc.City,
		Description: sc.Description,
		Stages:      stages,
		CreatedAt:   sc.CreatedAt,
	}, nil
}

func (s *DocStore) UpdateScenario(ctx context.Context, id string, req AdminScenarioRequest) (AdminScenarioDetail, error) {
	var sc scenarioDoc
	if err := s.get(ctx, "scenarios", id, &sc); err != nil {
		return AdminScenarioDetail{}, err
	}
	sc.Name = req.Name
	sc.City = req.City
	sc.Description = req.Description
	sc.Stages = req.Stages
	if err := s.putScenario(ctx, sc); err != nil {
		return AdminScenarioDetail{}, err
	}
	return AdminScenarioDetail{
		ID:          id,
		Name:        req.Name,
		City:        req.City,
		Description: req.Description,
		Stages:      req.Stages,
		CreatedAt:   sc.CreatedAt,
	}, nil
}

func (s *DocStore) DeleteScenario(ctx context.Context, id string) error {
	return s.del(ctx, "scenarios", id)
}

func (s *DocStore) ScenarioHasGames(ctx context.Context, scenarioID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM games WHERE scenario_id = ?`, scenarioID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Admin games

func (s *DocStore) ListGames(ctx context.Context) ([]AdminGameSummary, error) {
	scenarios, err := s.allScenarios(ctx)
	if err != nil {
		return nil, err
	}
	scenarioNames := map[string]string{}
	for _, sc := range scenarios {
		scenarioNames[sc.ID] = sc.Name
	}

	allGames, err := s.allGames(ctx)
	if err != nil {
		return nil, err
	}

	var games []AdminGameSummary
	for _, g := range allGames {
		games = append(games, AdminGameSummary{
			ID:           g.ID,
			ScenarioID:   g.ScenarioID,
			ScenarioName: scenarioNames[g.ScenarioID],
			Status:       g.Status,
			TimerMinutes: g.TimerMinutes,
			TeamCount:    len(g.Teams),
			CreatedAt:    g.CreatedAt,
		})
	}
	// Sort newest first.
	for i, j := 0, len(games)-1; i < j; i, j = i+1, j-1 {
		games[i], games[j] = games[j], games[i]
	}
	return games, nil
}

func (s *DocStore) CreateGame(ctx context.Context, req AdminGameRequest) (AdminGameDetail, error) {
	id := newID()
	now := nowUTC()
	doc := gameDoc{
		ID:           id,
		ScenarioID:   req.ScenarioID,
		Status:       req.Status,
		TimerMinutes: req.TimerMinutes,
		CreatedAt:    now,
		Teams:        []teamDoc{},
	}
	if err := s.putGame(ctx, doc); err != nil {
		return AdminGameDetail{}, err
	}
	return AdminGameDetail{
		ID:           id,
		ScenarioID:   req.ScenarioID,
		Status:       req.Status,
		TimerMinutes: req.TimerMinutes,
		Teams:        []AdminTeamItem{},
		CreatedAt:    now,
	}, nil
}

func (s *DocStore) GetGame(ctx context.Context, id string) (AdminGameDetail, error) {
	g, err := s.getGame(ctx, id)
	if err != nil {
		return AdminGameDetail{}, err
	}

	var sc scenarioDoc
	if err := s.get(ctx, "scenarios", g.ScenarioID, &sc); err != nil && !errors.Is(err, ErrNotFound) {
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
		ID:           g.ID,
		ScenarioID:   g.ScenarioID,
		ScenarioName: sc.Name,
		Status:       g.Status,
		TimerMinutes: g.TimerMinutes,
		Teams:        teams,
		CreatedAt:    g.CreatedAt,
	}, nil
}

func (s *DocStore) UpdateGame(ctx context.Context, id string, req AdminGameRequest) (AdminGameDetail, error) {
	g, err := s.getGame(ctx, id)
	if err != nil {
		return AdminGameDetail{}, err
	}
	g.ScenarioID = req.ScenarioID
	g.Status = req.Status
	g.TimerMinutes = req.TimerMinutes
	if err := s.putGame(ctx, g); err != nil {
		return AdminGameDetail{}, err
	}
	return AdminGameDetail{
		ID:           id,
		ScenarioID:   req.ScenarioID,
		Status:       req.Status,
		TimerMinutes: req.TimerMinutes,
		CreatedAt:    g.CreatedAt,
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

func (s *DocStore) ScenarioName(ctx context.Context, scenarioID string) (string, error) {
	var name string
	err := s.db.QueryRowContext(ctx,
		`SELECT name FROM scenarios WHERE id = ?`, scenarioID,
	).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return name, err
}

// SeedDemo populates the store with demo data if empty.
func (s *DocStore) SeedDemo(ctx context.Context) error {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM scenarios`,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := nowUTC()

	scenario := scenarioDoc{
		ID:          "s0000000deadbeef",
		Name:        "Lima Centro Historico",
		City:        "Lima",
		Description: "Explore the historic center of Lima through four iconic landmarks.",
		CreatedAt:   now,
		Stages: []AdminStage{
			{StageNumber: 1, Location: "Plaza Mayor", Clue: "Head to the main square where Pizarro founded the city. Look for the bronze fountain in the center.", Question: "What year was the fountain in Plaza Mayor built?", CorrectAnswer: "1651", Lat: -12.0464, Lng: -77.0300},
			{StageNumber: 2, Location: "Iglesia de San Francisco", Clue: "Walk south to the yellow church with famous underground tunnels.", Question: "What are the underground tunnels beneath San Francisco called?", CorrectAnswer: "catacombs", Lat: -12.0463, Lng: -77.0275},
			{StageNumber: 3, Location: "Jiron de la Union", Clue: "Stroll down Limas most famous pedestrian street. Find the statue of the liberator.", Question: "Which liberator has a statue on Jiron de la Union?", CorrectAnswer: "San Martin", Lat: -12.0500, Lng: -77.0350},
			{StageNumber: 4, Location: "Parque de la Muralla", Clue: "Follow the old city wall to the park along the Rimac river.", Question: "What century were the original city walls built in?", CorrectAnswer: "17th", Lat: -12.0450, Lng: -77.0260},
		},
	}
	if err := s.putScenario(ctx, scenario); err != nil {
		return err
	}

	game := gameDoc{
		ID:           "g0000000deadbeef",
		ScenarioID:   "s0000000deadbeef",
		Status:       "active",
		TimerMinutes: 120,
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
	if err := s.putGame(ctx, game); err != nil {
		return err
	}

	return nil
}

// Ensure DocStore implements Store at compile time.
var _ Store = (*DocStore)(nil)
