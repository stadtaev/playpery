package handler_test

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"nhooyr.io/websocket"

	"github.com/playperu/cityquiz/internal/handler"
)

func TestWSEcho(t *testing.T) {
	h := handler.NewWSEcho(slog.Default())
	srv := httptest.NewServer(h.Routes())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + srv.URL[len("http"):] + "/echo"

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	messages := []string{"hello cityquiz", "¡hola lima!", "🎯"}

	for _, want := range messages {
		if err := conn.Write(ctx, websocket.MessageText, []byte(want)); err != nil {
			t.Fatalf("write %q: %v", want, err)
		}

		_, got, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("read: %v", err)
		}

		if string(got) != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	conn.Close(websocket.StatusNormalClosure, "done")
}
