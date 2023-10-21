package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ostcar/bietrunde/config"
	"github.com/ostcar/bietrunde/model"
	"github.com/ostcar/bietrunde/web/template"
	"github.com/ostcar/sticky"
	"golang.org/x/exp/slog"
)

const cookieName = "bietrunde"

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
	router.Handle("/", handleError(s.handleHome))
	router.Handle("/register", handleError(s.handleRegister))
	router.Handle("/edit", handleError(s.handleEdit))

	s.Handler = loggingMiddleware(router)
}

func (s server) handleHome(w http.ResponseWriter, r *http.Request) error {
	m, done := s.model.ForReading()
	defer done()

	bieterID := readAuthCookie(r)
	bieter, ok := m.Bieter[bieterID]

	if bieterID == 0 || !ok {
		return template.LoginPage(m.State).Render(r.Context(), w)
	}

	return template.Bieter(bieter).Render(r.Context(), w)
}

func (s server) handleRegister(w http.ResponseWriter, r *http.Request) error {
	m, write := s.model.ForWriting()

	id, event := m.BieterCreate()
	write(event)

	log.Printf("set id %d", id)
	setAuthCookie(w, id)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
}

func (s server) handleEdit(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			return err
		}

		m, write := s.model.ForWriting()
		bieterID := readAuthCookie(r)
		bieter, ok := m.Bieter[bieterID]

		if bieterID == 0 || !ok {
			return template.LoginPage(m.State).Render(r.Context(), w)
		}

		r.ParseForm()
		bieter.Vorname = r.Form.Get("vorname")
		bieter.Nachname = r.Form.Get("nachname")
		bieter.Mail = r.Form.Get("mail")
		bieter.Adresse = r.Form.Get("adresse")
		bieter.Mitglied = r.Form.Has("mitglied")
		bieter.Verteilstelle = model.VerteilstelleFromAttr(r.Form.Get("verteilstelle"))
		bieter.Teilpartner = r.Form.Get("teilpartner")
		bieter.IBAN = r.Form.Get("iban")
		bieter.Kontoinhaber = r.Form.Get("kontoinhaber")
		bieter.Jaehrlich = r.Form.Get("abbuchung") == "jaehrlich"
		if err := write(m.BieterUpdate(bieter)); err != nil {
			return err
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return nil
	}

	m, done := s.model.ForReading()
	defer done()

	bieterID := readAuthCookie(r)
	bieter, ok := m.Bieter[bieterID]

	if bieterID == 0 || !ok {
		return template.LoginPage(m.State).Render(r.Context(), w)
	}

	return template.BieterEdit(bieter).Render(r.Context(), w)
}

func setAuthCookie(w http.ResponseWriter, id int) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    strconv.Itoa(id),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func readAuthCookie(r *http.Request) int {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return 0
	}

	// Ignore the error, since we want to return 0 on error anyway.
	id, _ := strconv.Atoi(cookie.Value)
	return id
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
