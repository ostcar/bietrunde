package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ostcar/bietrunde/config"
	"github.com/ostcar/bietrunde/model"
	"github.com/ostcar/timer/sticky"
	"golang.org/x/exp/slog"
)

//go:embed files
var publicFiles embed.FS

//go:generate templ generate -path template

// Run starts the server.
func Run(ctx context.Context, s *sticky.Sticky[model.Model], cfg config.Config) error {
	handler := newServer(cfg, s)

	httpSRV := &http.Server{
		Addr:        cfg.WebListenAddr,
		Handler:     handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	// Shutdown logic in separate goroutine.
	wait := make(chan error)
	go func() {
		// Wait for the context to be closed.
		<-ctx.Done()

		if err := httpSRV.Shutdown(context.WithoutCancel(ctx)); err != nil {
			wait <- fmt.Errorf("HTTP server shutdown: %w", err)
			return
		}
		wait <- nil
	}()

	fmt.Printf("Listen webserver on: %s\n", cfg.WebListenAddr)
	if err := httpSRV.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP Server failed: %v", err)
	}

	return <-wait
}

type server struct {
	http.Handler
	cfg   config.Config
	model *sticky.Sticky[model.Model]
}

func newServer(cfg config.Config, s *sticky.Sticky[model.Model]) server {
	srv := server{
		cfg:   cfg,
		model: s,
	}
	srv.registerHandlers()

	return srv
}

func (s *server) registerHandlers() {
	router := mux.NewRouter()

	router.PathPrefix("/assets").Handler(handleStatic())
	// router.Handle("/", handleError(s.campainList))
	// router.Handle("/create_campain", handleError(s.campainCreate))
	// router.Handle("/campain/{campain_id:[0-9]+}", handleError(s.campainDetail))
	// router.Handle("/campain/{campain_id:[0-9]+}/create_day", handleError(s.dayCreate))

	s.Handler = loggingMiddleware(router)
}

func handleStatic() http.Handler {
	files, err := fs.Sub(publicFiles, "files")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(files))
}

func handleError(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			// TODO
			http.Error(w, err.Error(), 500)
		}
	}
}

type responselogger struct {
	http.ResponseWriter
	code int
}

func (r *responselogger) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &responselogger{w, 200}
		next.ServeHTTP(writer, r)
		slog.Info("Got request", "method", r.Method, "status", writer.code, "uri", r.RequestURI)
	})
}
