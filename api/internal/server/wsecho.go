package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

func handleWSEcho(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			logger.Error("websocket accept failed", "error", err)
			return
		}
		defer conn.CloseNow()

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
		defer cancel()

		for {
			typ, msg, err := conn.Read(ctx)
			if err != nil {
				logger.Debug("websocket read ended", "error", err)
				return
			}

			if err := conn.Write(ctx, typ, msg); err != nil {
				logger.Debug("websocket write failed", "error", err)
				return
			}
		}
	}
}
