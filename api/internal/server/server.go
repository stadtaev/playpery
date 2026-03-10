package server

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/quic-go/quic-go/http3"
)

type Server struct {
	tcpSrv *http.Server
	h3Srv  *http3.Server // nil when TLS not configured
	logger *slog.Logger
}

func New(addr string, logger *slog.Logger, admin AdminStore, clients *Registry, adminDB *sql.DB, spaDir, dataDir string, tlsCert, tlsKey string) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(newStructuredLogger(logger))
	r.Use(middleware.Recoverer)

	addRoutes(r, logger, admin, clients, adminDB, spaDir, dataDir)

	s := &Server{
		tcpSrv: &http.Server{
			Addr:              addr,
			Handler:           r,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		logger: logger,
	}

	if tlsCert != "" && tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			logger.Error("failed to load TLS cert, falling back to plain HTTP", "error", err)
			return s
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
		}

		s.tcpSrv.TLSConfig = tlsConfig

		s.h3Srv = &http3.Server{
			Addr:      addr,
			Handler:   r,
			TLSConfig: http3.ConfigureTLSConfig(tlsConfig.Clone()),
		}
	}

	return s
}

func (s *Server) Run(_ context.Context) error {
	ln, err := net.Listen("tcp", s.tcpSrv.Addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.tcpSrv.Addr, err)
	}

	if s.h3Srv != nil {
		// Start HTTP/3 (UDP) in background.
		go func() {
			s.logger.Info("starting http/3 server (udp)", "addr", s.tcpSrv.Addr)
			if err := s.h3Srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error("http/3 server error", "error", err)
			}
		}()

		// Wrap TCP handler to set Alt-Svc header advertising HTTP/3.
		origHandler := s.tcpSrv.Handler
		s.tcpSrv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := s.h3Srv.SetQUICHeaders(w.Header()); err != nil {
				s.logger.Debug("failed to set Alt-Svc header", "error", err)
			}
			origHandler.ServeHTTP(w, r)
		})

		// Serve TCP with TLS (HTTP/1.1 + HTTP/2).
		s.logger.Info("starting https server (tcp)", "addr", s.tcpSrv.Addr)
		err = s.tcpSrv.ServeTLS(ln, "", "")
	} else {
		// Plain HTTP mode (no TLS configured).
		s.logger.Info("starting http server (tcp, no tls)", "addr", s.tcpSrv.Addr)
		err = s.tcpSrv.Serve(ln)
	}

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var h3Err error
	if s.h3Srv != nil {
		h3Err = s.h3Srv.Shutdown(ctx)
	}

	tcpErr := s.tcpSrv.Shutdown(ctx)

	if h3Err != nil {
		return h3Err
	}
	return tcpErr
}

func newStructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.Info("http request",
					"method", r.Method,
					"path", r.URL.Path,
					"proto", r.Proto,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration_ms", time.Since(start).Milliseconds(),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
