package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nhooyr.io/websocket"
)

type WSEcho struct {
	logger *slog.Logger
}

func NewWSEcho(logger *slog.Logger) *WSEcho {
	return &WSEcho{logger: logger}
}

func (ws *WSEcho) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/echo", ws.echo)
	return r
}

func (ws *WSEcho) echo(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		ws.logger.Error("websocket accept failed", "error", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	for {
		typ, msg, err := conn.Read(ctx)
		if err != nil {
			ws.logger.Debug("websocket read ended", "error", err)
			return
		}

		if err := conn.Write(ctx, typ, msg); err != nil {
			ws.logger.Debug("websocket write failed", "error", err)
			return
		}
	}
}
