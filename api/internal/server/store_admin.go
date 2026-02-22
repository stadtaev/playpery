package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type AdminAuth interface {
	AdminByEmail(ctx context.Context, email string) (adminID, passwordHash string, err error)
	CreateAdminSession(ctx context.Context, adminID string) (sessionID string, err error)
	DeleteAdminSession(ctx context.Context, sessionID string) error
	AdminFromSession(ctx context.Context, sessionID string) (adminSession, error)
	ListClients(ctx context.Context) ([]ClientInfo, error)
	CreateClient(ctx context.Context, slug, name string) error
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

type AdminStore struct {
	db *sql.DB
}

func NewAdminStore(ctx context.Context, db *sql.DB) (*AdminStore, error) {
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
	} {
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return nil, fmt.Errorf("creating table: %w", err)
		}
	}

	s := &AdminStore{db: db}
	if err := s.seedIfEmpty(ctx); err != nil {
		return nil, fmt.Errorf("seeding admin: %w", err)
	}
	return s, nil
}

func (s *AdminStore) seedIfEmpty(ctx context.Context) error {
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

func (s *AdminStore) AdminByEmail(ctx context.Context, email string) (string, string, error) {
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

func (s *AdminStore) CreateAdminSession(ctx context.Context, adminID string) (string, error) {
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

func (s *AdminStore) DeleteAdminSession(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM admin_sessions WHERE id = ?`, sessionID,
	)
	return err
}

func (s *AdminStore) AdminFromSession(ctx context.Context, sessionID string) (adminSession, error) {
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

func (s *AdminStore) ListClients(ctx context.Context) ([]ClientInfo, error) {
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

func (s *AdminStore) CreateClient(ctx context.Context, slug, name string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO clients (slug, name) VALUES (?, ?)`, slug, name,
	)
	return err
}

var _ AdminAuth = (*AdminStore)(nil)
