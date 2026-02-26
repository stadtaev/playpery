package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type AdminStore interface {
	AdminByEmail(ctx context.Context, email string) (adminID, passwordHash string, err error)
	CreateAdminSession(ctx context.Context, adminID string) (sessionID string, err error)
	DeleteAdminSession(ctx context.Context, sessionID string) error
	AdminFromSession(ctx context.Context, sessionID string) (adminSession, error)
	ListClients(ctx context.Context) ([]ClientInfo, error)
	CreateClient(ctx context.Context, slug, name string) error

	ListScenarios(ctx context.Context) ([]AdminScenarioSummary, error)
	CreateScenario(ctx context.Context, req AdminScenarioRequest) (AdminScenarioDetail, error)
	GetScenario(ctx context.Context, id string) (AdminScenarioDetail, error)
	UpdateScenario(ctx context.Context, id string, req AdminScenarioRequest) (AdminScenarioDetail, error)
	DeleteScenario(ctx context.Context, id string) error
	ScenarioHasGames(ctx context.Context, scenarioID string, clients *Registry) (bool, error)
}

type ClientInfo struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type adminDoc struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

type adminSessionDoc struct {
	ID      string `json:"id"`
	AdminID string `json:"adminId"`
	Email   string `json:"email"`
}

type AdminDocStore struct {
	db *sql.DB
}

func NewAdminDocStore(ctx context.Context, db *sql.DB) (*AdminDocStore, error) {
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS admins (
			id    TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			data  JSONB NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS admin_sessions (
			id   TEXT PRIMARY KEY,
			data JSONB NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS clients (
			slug TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS scenarios (
			id   TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			data JSONB NOT NULL
		)`,
	} {
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return nil, fmt.Errorf("creating table: %w", err)
		}
	}

	s := &AdminDocStore{db: db}
	if err := s.seedIfEmpty(ctx); err != nil {
		return nil, fmt.Errorf("seeding admin: %w", err)
	}
	return s, nil
}

func (s *AdminDocStore) seedIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admins`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	admin := adminDoc{
		ID:           newID(),
		Email:        "admin@playperu.com",
		PasswordHash: "$2a$10$trCdqP4npsbw0R1vQxVwXeT1HebzRmP01SXaNGPz1eSAZ7mpcL0Uu",
	}
	data, err := json.Marshal(admin)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO admins (id, email, data) VALUES (?, ?, jsonb(?))`,
		admin.ID, admin.Email, string(data),
	)
	return err
}

func (s *AdminDocStore) AdminByEmail(ctx context.Context, email string) (string, string, error) {
	var data string
	err := s.db.QueryRowContext(ctx,
		`SELECT json(data) FROM admins WHERE email = ?`, email,
	).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNotFound
	}
	if err != nil {
		return "", "", err
	}
	var a adminDoc
	if err := json.Unmarshal([]byte(data), &a); err != nil {
		return "", "", err
	}
	return a.ID, a.PasswordHash, nil
}

func (s *AdminDocStore) CreateAdminSession(ctx context.Context, adminID string) (string, error) {
	// Look up admin email.
	var data string
	err := s.db.QueryRowContext(ctx,
		`SELECT json(data) FROM admins WHERE id = ?`, adminID,
	).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	var a adminDoc
	if err := json.Unmarshal([]byte(data), &a); err != nil {
		return "", err
	}

	sessionID := newID()
	sessData, err := json.Marshal(adminSessionDoc{
		ID:      sessionID,
		AdminID: adminID,
		Email:   a.Email,
	})
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO admin_sessions (id, data) VALUES (?, jsonb(?))`,
		sessionID, string(sessData),
	)
	return sessionID, err
}

func (s *AdminDocStore) DeleteAdminSession(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM admin_sessions WHERE id = ?`, sessionID,
	)
	return err
}

func (s *AdminDocStore) AdminFromSession(ctx context.Context, sessionID string) (adminSession, error) {
	var data string
	err := s.db.QueryRowContext(ctx,
		`SELECT json(data) FROM admin_sessions WHERE id = ?`, sessionID,
	).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return adminSession{}, errNoAdminSession
	}
	if err != nil {
		return adminSession{}, err
	}
	var as adminSessionDoc
	if err := json.Unmarshal([]byte(data), &as); err != nil {
		return adminSession{}, err
	}
	return adminSession{AdminID: as.AdminID, Email: as.Email}, nil
}

func (s *AdminDocStore) ListClients(ctx context.Context) ([]ClientInfo, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT slug, name FROM clients ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []ClientInfo
	for rows.Next() {
		var c ClientInfo
		if err := rows.Scan(&c.Slug, &c.Name); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, nil
}

func (s *AdminDocStore) CreateClient(ctx context.Context, slug, name string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO clients (slug, name) VALUES (?, ?)`, slug, name,
	)
	return err
}

// Scenario CRUD — global, stored in admin DB.

func (s *AdminDocStore) ListScenarios(ctx context.Context) ([]AdminScenarioSummary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT json(data) FROM scenarios ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []AdminScenarioSummary
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var sc scenario
		if err := json.Unmarshal([]byte(data), &sc); err != nil {
			return nil, err
		}
		scenarios = append(scenarios, AdminScenarioSummary{
			ID:          sc.ID,
			Name:        sc.Name,
			City:        sc.City,
			Description: sc.Description,
			StageCount:  len(sc.Stages),
			CreatedAt:   sc.CreatedAt,
		})
	}
	// Newest first.
	for i, j := 0, len(scenarios)-1; i < j; i, j = i+1, j-1 {
		scenarios[i], scenarios[j] = scenarios[j], scenarios[i]
	}
	return scenarios, nil
}

func (s *AdminDocStore) CreateScenario(ctx context.Context, req AdminScenarioRequest) (AdminScenarioDetail, error) {
	id := newID()
	now := nowUTC()
	doc := scenario{
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

func (s *AdminDocStore) GetScenario(ctx context.Context, id string) (AdminScenarioDetail, error) {
	var sc scenario
	if err := s.getDoc(ctx, "scenarios", id, &sc); err != nil {
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

func (s *AdminDocStore) UpdateScenario(ctx context.Context, id string, req AdminScenarioRequest) (AdminScenarioDetail, error) {
	var sc scenario
	if err := s.getDoc(ctx, "scenarios", id, &sc); err != nil {
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

func (s *AdminDocStore) DeleteScenario(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM scenarios WHERE id = ?`, id,
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

func (s *AdminDocStore) ScenarioHasGames(ctx context.Context, scenarioID string, clients *Registry) (bool, error) {
	clients.mu.RLock()
	stores := make([]*DocStore, 0, len(clients.stores))
	for _, st := range clients.stores {
		stores = append(stores, st)
	}
	clients.mu.RUnlock()

	for _, st := range stores {
		var count int
		err := st.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM games WHERE scenario_id = ?`, scenarioID,
		).Scan(&count)
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

// Internal helpers for scenario storage.

func (s *AdminDocStore) getDoc(ctx context.Context, table, id string, dest any) error {
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

func (s *AdminDocStore) putScenario(ctx context.Context, sc scenario) error {
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

// SeedDemoScenario creates the demo scenario in the admin DB if none exist.
func (s *AdminDocStore) SeedDemoScenario(ctx context.Context) (*scenario, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM scenarios`).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, nil
	}

	now := nowUTC()
	sc := scenario{
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
	if err := s.putScenario(ctx, sc); err != nil {
		return nil, err
	}
	return &sc, nil
}

var _ AdminStore = (*AdminDocStore)(nil)
