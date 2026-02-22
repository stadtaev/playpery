package server

import "errors"

type adminSession struct {
	AdminID string
	Email   string
}

var errNoAdminSession = errors.New("no valid admin session")

const adminCookieName = "admin_session"
