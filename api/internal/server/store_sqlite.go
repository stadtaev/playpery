package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

const demoClientID = "c0000000deadbeef"

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) PlayerFromToken(ctx context.Context, token string) (playerSession, error) {
	var sess playerSession
	err := s.db.QueryRowContext(ctx, `
		SELECT p.id, p.team_id, t.game_id
		FROM players p
		JOIN teams t ON t.id = p.team_id
		WHERE p.session_id = ?
	`, token).Scan(&sess.PlayerID, &sess.TeamID, &sess.GameID)
	if errors.Is(err, sql.ErrNoRows) {
		return sess, errNoSession
	}
	return sess, err
}

func (s *SQLiteStore) AdminFromSession(ctx context.Context, sessionID string) (adminSession, error) {
	var sess adminSession
	err := s.db.QueryRowContext(ctx, `
		SELECT a.id, a.email
		FROM admin_sessions s
		JOIN admins a ON a.id = s.admin_id
		WHERE s.id = ?
	`, sessionID).Scan(&sess.AdminID, &sess.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return adminSession{}, errNoAdminSession
	}
	return sess, err
}

func (s *SQLiteStore) TeamLookup(ctx context.Context, joinToken string) (TeamLookupResponse, error) {
	var resp TeamLookupResponse
	err := s.db.QueryRowContext(ctx, `
		SELECT t.id, t.name, s.name
		FROM teams t
		JOIN games g ON g.id = t.game_id
		JOIN scenarios s ON s.id = g.scenario_id
		WHERE t.join_token = ? AND g.status = 'active'
	`, joinToken).Scan(&resp.ID, &resp.Name, &resp.GameName)
	if errors.Is(err, sql.ErrNoRows) {
		return resp, ErrNotFound
	}
	return resp, err
}

func (s *SQLiteStore) JoinTeam(ctx context.Context, teamID, playerName string) (string, string, error) {
	var playerID, sessionID string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO players (team_id, name, session_id)
		VALUES (?, ?, lower(hex(randomblob(16))))
		RETURNING id, session_id
	`, teamID, playerName).Scan(&playerID, &sessionID)
	return playerID, sessionID, err
}

func (s *SQLiteStore) GameState(ctx context.Context, gameID, teamID string) (gameStateData, error) {
	var d gameStateData
	var startedAt sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT g.status, g.timer_minutes, g.started_at, s.stages, t.name
		FROM games g
		JOIN scenarios s ON s.id = g.scenario_id
		JOIN teams t ON t.id = ?
		WHERE g.id = ?
	`, teamID, gameID).Scan(&d.Status, &d.TimerMinutes, &startedAt, &d.StagesJSON, &d.TeamName)
	if startedAt.Valid {
		d.StartedAt = &startedAt.String
	}
	if errors.Is(err, sql.ErrNoRows) {
		return d, ErrNotFound
	}
	return d, err
}

func (s *SQLiteStore) ExpireGame(ctx context.Context, gameID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE games SET status = 'ended', ended_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ? AND status = 'active'
	`, gameID)
	return err
}

func (s *SQLiteStore) CountCorrectAnswers(ctx context.Context, gameID, teamID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM stage_results
		WHERE game_id = ? AND team_id = ? AND is_correct = 1
	`, gameID, teamID).Scan(&count)
	return count, err
}

func (s *SQLiteStore) RecordAnswer(ctx context.Context, gameID, teamID string, stageNumber int, answer string, isCorrect bool) error {
	isCorrectInt := 0
	if isCorrect {
		isCorrectInt = 1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO stage_results (game_id, team_id, stage_number, answer, is_correct, answered_at)
		VALUES (?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	`, gameID, teamID, stageNumber, answer, isCorrectInt)
	return err
}

func (s *SQLiteStore) ListPlayers(ctx context.Context, teamID string) ([]PlayerInfo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name FROM players WHERE team_id = ? ORDER BY joined_at
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []PlayerInfo
	for rows.Next() {
		var p PlayerInfo
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, nil
}

func (s *SQLiteStore) ListCompletedStages(ctx context.Context, gameID, teamID string) ([]CompletedStage, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT stage_number, answered_at
		FROM stage_results
		WHERE game_id = ? AND team_id = ? AND is_correct = 1
		ORDER BY stage_number
	`, gameID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var completed []CompletedStage
	for rows.Next() {
		var cs CompletedStage
		if err := rows.Scan(&cs.StageNumber, &cs.AnsweredAt); err != nil {
			return nil, err
		}
		cs.IsCorrect = true
		completed = append(completed, cs)
	}
	return completed, nil
}

func (s *SQLiteStore) AdminByEmail(ctx context.Context, email string) (string, string, error) {
	var adminID, passwordHash string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, password_hash FROM admins WHERE email = ?
	`, email).Scan(&adminID, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return adminID, passwordHash, err
}

func (s *SQLiteStore) CreateAdminSession(ctx context.Context, adminID string) (string, error) {
	var sessionID string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO admin_sessions (admin_id)
		VALUES (?)
		RETURNING id
	`, adminID).Scan(&sessionID)
	return sessionID, err
}

func (s *SQLiteStore) DeleteAdminSession(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE id = ?`, sessionID)
	return err
}

func (s *SQLiteStore) ListScenarios(ctx context.Context) ([]AdminScenarioSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, city, COALESCE(description, ''), stages, created_at
		FROM scenarios
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []AdminScenarioSummary
	for rows.Next() {
		var sc AdminScenarioSummary
		var stagesJSON string
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.City, &sc.Description, &stagesJSON, &sc.CreatedAt); err != nil {
			return nil, err
		}
		var stages []json.RawMessage
		json.Unmarshal([]byte(stagesJSON), &stages)
		sc.StageCount = len(stages)
		scenarios = append(scenarios, sc)
	}
	return scenarios, nil
}

func (s *SQLiteStore) CreateScenario(ctx context.Context, req AdminScenarioRequest) (AdminScenarioDetail, error) {
	stagesJSON, _ := json.Marshal(req.Stages)

	var id, createdAt string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO scenarios (id, name, city, description, stages)
		VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?)
		RETURNING id, created_at
	`, req.Name, req.City, req.Description, string(stagesJSON)).Scan(&id, &createdAt)
	if err != nil {
		return AdminScenarioDetail{}, err
	}

	return AdminScenarioDetail{
		ID:          id,
		Name:        req.Name,
		City:        req.City,
		Description: req.Description,
		Stages:      req.Stages,
		CreatedAt:   createdAt,
	}, nil
}

func (s *SQLiteStore) GetScenario(ctx context.Context, id string) (AdminScenarioDetail, error) {
	var d AdminScenarioDetail
	var stagesJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, city, COALESCE(description, ''), stages, created_at
		FROM scenarios WHERE id = ?
	`, id).Scan(&d.ID, &d.Name, &d.City, &d.Description, &stagesJSON, &d.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return d, ErrNotFound
	}
	if err != nil {
		return d, err
	}
	json.Unmarshal([]byte(stagesJSON), &d.Stages)
	if d.Stages == nil {
		d.Stages = []AdminStage{}
	}
	return d, nil
}

func (s *SQLiteStore) UpdateScenario(ctx context.Context, id string, req AdminScenarioRequest) (AdminScenarioDetail, error) {
	stagesJSON, _ := json.Marshal(req.Stages)

	var createdAt string
	err := s.db.QueryRowContext(ctx, `
		UPDATE scenarios SET name = ?, city = ?, description = ?, stages = ?
		WHERE id = ?
		RETURNING created_at
	`, req.Name, req.City, req.Description, string(stagesJSON), id).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return AdminScenarioDetail{}, ErrNotFound
	}
	if err != nil {
		return AdminScenarioDetail{}, err
	}

	return AdminScenarioDetail{
		ID:          id,
		Name:        req.Name,
		City:        req.City,
		Description: req.Description,
		Stages:      req.Stages,
		CreatedAt:   createdAt,
	}, nil
}

func (s *SQLiteStore) DeleteScenario(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM scenarios WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) ScenarioHasGames(ctx context.Context, scenarioID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM games WHERE scenario_id = ?
	`, scenarioID).Scan(&count)
	return count > 0, err
}

func (s *SQLiteStore) ListGames(ctx context.Context) ([]AdminGameSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT g.id, g.scenario_id, s.name, g.status, g.timer_minutes,
			(SELECT COUNT(*) FROM teams t WHERE t.game_id = g.id),
			g.created_at
		FROM games g
		JOIN scenarios s ON s.id = g.scenario_id
		ORDER BY g.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []AdminGameSummary
	for rows.Next() {
		var g AdminGameSummary
		if err := rows.Scan(&g.ID, &g.ScenarioID, &g.ScenarioName, &g.Status, &g.TimerMinutes, &g.TeamCount, &g.CreatedAt); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func (s *SQLiteStore) CreateGame(ctx context.Context, req AdminGameRequest) (AdminGameDetail, error) {
	var id, createdAt string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO games (id, scenario_id, client_id, status, timer_minutes)
		VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?)
		RETURNING id, created_at
	`, req.ScenarioID, demoClientID, req.Status, req.TimerMinutes).Scan(&id, &createdAt)
	if err != nil {
		return AdminGameDetail{}, err
	}

	return AdminGameDetail{
		ID:           id,
		ScenarioID:   req.ScenarioID,
		Status:       req.Status,
		TimerMinutes: req.TimerMinutes,
		Teams:        []AdminTeamItem{},
		CreatedAt:    createdAt,
	}, nil
}

func (s *SQLiteStore) GetGame(ctx context.Context, id string) (AdminGameDetail, error) {
	var g AdminGameDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT g.id, g.scenario_id, s.name, g.status, g.timer_minutes, g.created_at
		FROM games g
		JOIN scenarios s ON s.id = g.scenario_id
		WHERE g.id = ?
	`, id).Scan(&g.ID, &g.ScenarioID, &g.ScenarioName, &g.Status, &g.TimerMinutes, &g.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return g, ErrNotFound
	}
	if err != nil {
		return g, err
	}

	teams, err := s.ListTeams(ctx, id)
	if err != nil {
		return g, err
	}
	g.Teams = teams
	return g, nil
}

func (s *SQLiteStore) UpdateGame(ctx context.Context, id string, req AdminGameRequest) (AdminGameDetail, error) {
	var createdAt string
	err := s.db.QueryRowContext(ctx, `
		UPDATE games SET scenario_id = ?, status = ?, timer_minutes = ?
		WHERE id = ?
		RETURNING created_at
	`, req.ScenarioID, req.Status, req.TimerMinutes, id).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return AdminGameDetail{}, ErrNotFound
	}
	if err != nil {
		return AdminGameDetail{}, err
	}

	return AdminGameDetail{
		ID:           id,
		ScenarioID:   req.ScenarioID,
		Status:       req.Status,
		TimerMinutes: req.TimerMinutes,
		CreatedAt:    createdAt,
	}, nil
}

func (s *SQLiteStore) DeleteGame(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM games WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) GameHasPlayers(ctx context.Context, gameID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM players p
		JOIN teams t ON t.id = p.team_id
		WHERE t.game_id = ?
	`, gameID).Scan(&count)
	return count > 0, err
}

func (s *SQLiteStore) DeleteTeamsByGame(ctx context.Context, gameID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM teams WHERE game_id = ?`, gameID)
	return err
}

func (s *SQLiteStore) ListTeams(ctx context.Context, gameID string) ([]AdminTeamItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.join_token, COALESCE(t.guide_name, ''),
			(SELECT COUNT(*) FROM players p WHERE p.team_id = t.id),
			t.created_at
		FROM teams t
		WHERE t.game_id = ?
		ORDER BY t.created_at
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := []AdminTeamItem{}
	for rows.Next() {
		var t AdminTeamItem
		if err := rows.Scan(&t.ID, &t.Name, &t.JoinToken, &t.GuideName, &t.PlayerCount, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (s *SQLiteStore) CreateTeam(ctx context.Context, gameID string, req AdminTeamRequest, token string) (AdminTeamItem, error) {
	var id, createdAt string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO teams (id, game_id, name, join_token, guide_name)
		VALUES (lower(hex(randomblob(16))), ?, ?, ?, NULLIF(?, ''))
		RETURNING id, created_at
	`, gameID, req.Name, token, req.GuideName).Scan(&id, &createdAt)
	if err != nil {
		return AdminTeamItem{}, err
	}

	return AdminTeamItem{
		ID:          id,
		Name:        req.Name,
		JoinToken:   token,
		GuideName:   req.GuideName,
		PlayerCount: 0,
		CreatedAt:   createdAt,
	}, nil
}

func (s *SQLiteStore) UpdateTeam(ctx context.Context, gameID, teamID string, req AdminTeamRequest) (AdminTeamItem, error) {
	var joinToken, createdAt string
	var playerCount int
	err := s.db.QueryRowContext(ctx, `
		UPDATE teams SET name = ?, guide_name = NULLIF(?, '')
		WHERE id = ? AND game_id = ?
		RETURNING join_token, created_at, (SELECT COUNT(*) FROM players WHERE team_id = teams.id)
	`, req.Name, req.GuideName, teamID, gameID).Scan(&joinToken, &createdAt, &playerCount)
	if errors.Is(err, sql.ErrNoRows) {
		return AdminTeamItem{}, ErrNotFound
	}
	if err != nil {
		return AdminTeamItem{}, err
	}

	return AdminTeamItem{
		ID:          teamID,
		Name:        req.Name,
		JoinToken:   joinToken,
		GuideName:   req.GuideName,
		PlayerCount: playerCount,
		CreatedAt:   createdAt,
	}, nil
}

func (s *SQLiteStore) DeleteTeam(ctx context.Context, gameID, teamID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM teams WHERE id = ? AND game_id = ?`, teamID, gameID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) TeamHasPlayers(ctx context.Context, gameID, teamID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM players p
		JOIN teams t ON t.id = p.team_id
		WHERE p.team_id = ? AND t.game_id = ?
	`, teamID, gameID).Scan(&count)
	return count > 0, err
}

func (s *SQLiteStore) GameExists(ctx context.Context, gameID string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM games WHERE id = ?`, gameID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *SQLiteStore) ScenarioName(ctx context.Context, scenarioID string) (string, error) {
	var name string
	err := s.db.QueryRowContext(ctx, `SELECT name FROM scenarios WHERE id = ?`, scenarioID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return name, err
}
