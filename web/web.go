package web

import (
	"cmp"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ostcar/bietrunde/config"
	"github.com/ostcar/bietrunde/model"
	"github.com/ostcar/bietrunde/user"
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
	router.Handle("/edit", handleError(s.handleEdit))

	router.Handle("/admin", handleError(s.handleAdmin))
	router.Handle("/admin/login", handleError(s.handleAdminLogin))
	router.Handle("/admin/delete/{id:[0-9]+}", handleError(s.handleAdminDelete))

	router.Handle("/register", handleError(s.handleRegister))
	router.Handle("/login", handleError(s.handleLogin))
	router.Handle("/logout", handleError(s.handleLogout))

	s.Handler = loggingMiddleware(router)
}

func (s server) handleHome(w http.ResponseWriter, r *http.Request) error {
	m, done := s.model.ForReading()
	defer done()

	// Ignore error and use anonymous instead.
	user, _ := user.FromRequest(r, []byte(s.cfg.Secred))

	bieter, ok := m.Bieter[user.BieterID]

	if user.IsAnonymous() || !ok {
		return template.LoginPage(user, m.State).Render(r.Context(), w)
	}

	return template.Bieter(user, bieter).Render(r.Context(), w)
}

func (s server) handleEdit(w http.ResponseWriter, r *http.Request) error {
	user, err := user.FromRequest(r, []byte(s.cfg.Secred))
	if err != nil {
		return err
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			return err
		}

		m, write := s.model.ForWriting()
		defer write()

		bieter, ok := m.Bieter[user.BieterID]

		if user.IsAnonymous() || !ok {
			return template.LoginPage(user, m.State).Render(r.Context(), w)
		}

		r.ParseForm()
		bieter.Vorname = strings.TrimSpace(r.Form.Get("vorname"))
		bieter.Nachname = strings.TrimSpace(r.Form.Get("nachname"))
		bieter.Mail = strings.TrimSpace(r.Form.Get("mail"))
		bieter.Adresse = strings.TrimSpace(r.Form.Get("adresse"))
		bieter.Mitglied = r.Form.Has("mitglied")
		bieter.Verteilstelle = model.VerteilstelleFromAttr(strings.TrimSpace(r.Form.Get("verteilstelle")))
		bieter.Teilpartner = strings.TrimSpace(r.Form.Get("teilpartner"))
		bieter.IBAN = strings.TrimSpace(r.Form.Get("iban"))
		bieter.Kontoinhaber = strings.TrimSpace(r.Form.Get("kontoinhaber"))
		bieter.Jaehrlich = strings.TrimSpace(r.Form.Get("abbuchung")) == "jaehrlich"
		if err := write(m.BieterUpdate(bieter)); err != nil {
			return err
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return nil
	}

	m, done := s.model.ForReading()
	defer done()

	bieter, ok := m.Bieter[user.BieterID]

	if user.IsAnonymous() || !ok {
		return template.LoginPage(user, m.State).Render(r.Context(), w)
	}

	return template.BieterEdit(user, bieter).Render(r.Context(), w)
}

func (s server) handleAdmin(w http.ResponseWriter, r *http.Request) error {
	user, err := user.FromRequest(r, []byte(s.cfg.Secred))
	if err != nil {
		return err
	}

	m, done := s.model.ForReading()
	defer done()

	bieter := make([]model.Bieter, 0, len(m.Bieter))
	for _, v := range m.Bieter {
		bieter = append(bieter, v)
	}
	slices.SortFunc(bieter, func(a, b model.Bieter) int {
		if c := cmp.Compare(a.Nachname, b.Nachname); c != 0 {
			return c
		}

		if c := cmp.Compare(a.Vorname, b.Vorname); c != 0 {
			return c
		}

		return cmp.Compare(a.ID, b.ID)
	})

	return template.Admin(user, bieter).Render(r.Context(), w)
}

func (s server) handleAdminLogin(w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()

	pw := r.Form.Get("password")
	if pw == s.cfg.AdminToken {
		user, err := user.FromRequest(r, []byte(s.cfg.Secred))
		if err != nil {
			return err
		}
		user.IsAdmin = true
		user.SetCookie(w, []byte(s.cfg.Secred))
	}

	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
	return nil
}

func (s server) handleAdminDelete(w http.ResponseWriter, r *http.Request) error {
	// TODO: Make me a post
	user, err := user.FromRequest(r, []byte(s.cfg.Secred))
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		return nil
	}

	m, write := s.model.ForWriting()
	defer write()

	bietID, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := write(m.BieterDelete(bietID)); err != nil {
		return err
	}

	http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
	return nil
}

func (s server) handleRegister(w http.ResponseWriter, r *http.Request) error {
	m, write := s.model.ForWriting()
	defer write()

	bieterID, event := m.BieterCreate()
	if err := write(event); err != nil {
		return err
	}

	user := user.FromID(bieterID)
	user.SetCookie(w, []byte(s.cfg.Secred))

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
}

func (s server) handleLogin(w http.ResponseWriter, r *http.Request) error {
	m, done := s.model.ForReading()
	defer done()

	if err := r.ParseForm(); err != nil {
		return err
	}

	bieterID, _ := strconv.Atoi(r.Form.Get("bietnumber"))
	_, ok := m.Bieter[bieterID]
	user := user.FromID(bieterID)
	if user.IsAnonymous() || !ok {
		return fmt.Errorf("invalid number")
	}

	user.SetCookie(w, []byte(s.cfg.Secred))

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
}

func (s server) handleLogout(w http.ResponseWriter, r *http.Request) error {
	user.Logout(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
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
