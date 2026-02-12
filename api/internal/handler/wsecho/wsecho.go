package wsecho

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"nhooyr.io/websocket"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/echo", h.echo)
	return r
}

func (h *Handler) echo(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.logger.Error("websocket accept failed", "error", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	for {
		typ, msg, err := conn.Read(ctx)
		if err != nil {
			h.logger.Debug("websocket read ended", "error", err)
			return
		}

		if err := conn.Write(ctx, typ, msg); err != nil {
			h.logger.Debug("websocket write failed", "error", err)
			return
		}
	}
}
