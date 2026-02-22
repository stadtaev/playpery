package server

import (
	"fmt"
	"net/http"
	"time"
)

func handleEvents(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			writeError(w, http.StatusUnauthorized, "token query parameter required")
			return
		}

		store := clientStore(r)

		sess, err := store.PlayerFromToken(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid session token")
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, http.StatusInternalServerError, "streaming not supported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		flusher.Flush()

		ch := broker.Subscribe(sess.TeamID)
		defer broker.Unsubscribe(sess.TeamID, ch)

		ping := time.NewTicker(30 * time.Second)
		defer ping.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case data := <-ch:
				fmt.Fprintf(w, "event: state\ndata: %s\n\n", data)
				flusher.Flush()
			case <-ping.C:
				fmt.Fprintf(w, ": ping\n\n")
				flusher.Flush()
			}
		}
	}
}
