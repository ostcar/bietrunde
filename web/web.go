package web

import (
	"cmp"
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ostcar/bietrunde/config"
	"github.com/ostcar/bietrunde/model"
	"github.com/ostcar/bietrunde/pdf"
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
	router.Handle("/logout", handleError(s.handleLogout))

	router.Handle("/", handleError(s.handleHome))
	router.Handle("/edit", handleError(s.handleEdit))
	router.Handle("/vertrag", handleError(s.handleVertrag))

	router.Handle("/admin", handleError(s.adminPage(s.handleAdmin)))
	router.Handle("/admin/new", handleError(s.adminPage(s.handleAdminNew)))
	router.Handle("/admin/edit/{id:[0-9]+}", handleError(s.adminPage(s.handleAdminEdit)))
	router.Handle("/admin/delete/{id:[0-9]+}", handleError(s.adminPage(s.handleAdminDelete)))
	router.Handle("/admin/state", handleError(s.adminPage(s.handleAdminState)))
	router.Handle("/admin/reset-gebot", handleError(s.adminPage(s.handleAdminResetGebot)))

	s.Handler = loggingMiddleware(router)
}

func (s server) handleLogout(w http.ResponseWriter, r *http.Request) error {
	// TODO: Should this be only POST?
	user.Logout(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return nil
}

func (s server) handleHome(w http.ResponseWriter, r *http.Request) error {
	u, _ := user.FromRequest(r, []byte(s.cfg.Secred))
	if bietID := r.URL.Query().Get("biet-id"); bietID != "" {
		m, done := s.model.ForReading()
		defer done()
		state := m.State
		id, _ := strconv.Atoi(bietID)
		u.BieterID = id
		_, ok := m.Bieter[u.BieterID]
		if u.IsAnonymous() || !ok {
			return s.showLoginPage(r.Context(), w, state, "", "")
		}

		u.SetCookie(w, []byte(s.cfg.Secred))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return nil
	}

	m, done := s.model.ForReading()

	bieter, ok := m.Bieter[u.BieterID]
	if u.IsAnonymous() || !ok {
		done()
		return s.handleLoginPage(w, r)
	}

	state := m.State
	done()
	return s.handleBieterDetail(w, r, state, bieter)
}

func (s server) handleLoginPage(w http.ResponseWriter, r *http.Request) error {
	m, done := s.model.ForReading()
	state := m.State
	done()

	switch r.Method {
	case http.MethodGet:

		return s.showLoginPage(r.Context(), w, state, "", "")

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			return s.showLoginPage(r.Context(), w, state, "Formular kann nicht gelesen werden. Bitte versuche es erneut.", "Formular kann nicht gelesen werden. Bitte versuche es erneut.")
		}

		switch r.Form.Get("form") {
		case "login":
			return s.handleLoginPost(w, r)

		case "register":
			return s.handleRegisterPost(w, r)

		default:
			return s.showLoginPage(r.Context(), w, state, "Formular kann nicht gelesen werden. Bitte versuche es erneut.", "Formular kann nicht gelesen werden. Bitte versuche es erneut.")
		}

	default:
		http.Error(w, "Fehler", http.StatusMethodNotAllowed)
		return nil
	}
}

func (s server) showLoginPage(ctx context.Context, w http.ResponseWriter, state model.ServiceState, loginFormError, registerFormError string) error {
	return template.LoginPage(state, "", loginFormError, registerFormError).Render(ctx, w)
}

func (s server) handleLoginPost(w http.ResponseWriter, r *http.Request) error {
	m, done := s.model.ForReading()
	bieterID, _ := strconv.Atoi(r.Form.Get("bietnumber"))
	bieter, ok := m.Bieter[bieterID]
	state := m.State
	done()
	if !ok {
		return template.LoginPage(state, strconv.Itoa(bieterID), "Bietnummer nicht bekannt. ", "").Render(r.Context(), w)
	}

	user := user.FromID(bieterID)
	user.SetCookie(w, []byte(s.cfg.Secred))
	return s.handleBieterDetail(w, r, state, bieter)
}

func (s server) handleRegisterPost(w http.ResponseWriter, r *http.Request) error {
	m, write, done := s.model.ForWriting()

	state := m.State
	if state != model.StateRegistration {
		done()
		return template.LoginPage(m.State, "", "", "Registrierung nicht möglich").Render(r.Context(), w)
	}

	bieterID, event := m.BieterCreate()
	if err := write(event); err != nil {
		done()
		return template.LoginPage(m.State, "", "", userError(err)).Render(r.Context(), w)
	}

	bieter := m.Bieter[bieterID]

	done()

	user := user.FromID(bieterID)
	user.SetCookie(w, []byte(s.cfg.Secred))
	return s.handleBieterDetail(w, r, state, bieter)
}

func (s server) handleBieterDetail(w http.ResponseWriter, r *http.Request, state model.ServiceState, bieter model.Bieter) error {
	invalidFields := bieter.InvalidFields()
	switch r.Method {
	case http.MethodGet:
		return template.Bieter(state, bieter, "", invalidFields).Render(r.Context(), w)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			return template.Bieter(state, bieter, "Formular kann nicht gelesen werden. Bitte versuche es erneut.", invalidFields).Render(r.Context(), w)
		}

		if r.Form.Get("form") != "gebot" {
			return template.Bieter(state, bieter, "", invalidFields).Render(r.Context(), w)
		}

		gebot, err := model.GebotFromString(r.Form.Get("gebot"))
		if err != nil {
			return template.Bieter(state, bieter, "Dein Gebot muss eine Zahl sein.", invalidFields).Render(r.Context(), w)
		}

		m, write, done := s.model.ForWriting()
		defer done()

		if err := write(m.SetGebot(bieter.ID, gebot)); err != nil {
			return template.Bieter(state, bieter, userError(err), invalidFields).Render(r.Context(), w)
		}

		bieter = m.Bieter[bieter.ID]

		return template.Bieter(state, bieter, "", invalidFields).Render(r.Context(), w)

	default:
		http.Error(w, "Fehler", http.StatusMethodNotAllowed)
		return nil

	}
}

func (s server) handleEdit(w http.ResponseWriter, r *http.Request) error {
	user, err := user.FromRequest(r, []byte(s.cfg.Secred))
	if err != nil || user.BieterID == 0 {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return nil
	}

	switch r.Method {
	case http.MethodGet:
		m, done := s.model.ForReading()
		defer done()

		if m.State != model.StateRegistration {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return nil
		}

		bieter, ok := m.Bieter[user.BieterID]

		if user.IsAnonymous() || !ok {
			return template.LoginPage(m.State, "", "", "").Render(r.Context(), w)
		}

		return template.BieterEdit(bieter, bieter.InvalidFields()).Render(r.Context(), w)

	case http.MethodPost:
		m, write, done := s.model.ForWriting()
		defer done()

		if m.State != model.StateRegistration {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return nil
		}

		bieter, ok := m.Bieter[user.BieterID]
		if !ok {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return nil
		}

		bieter, errMsg := parseBieterEdit(r, bieter)
		if errMsg != "" {
			invalidFields := bieter.InvalidFields()
			invalidFields["form"] = "Formular kann nicht gelesen werden. Versuche es erneut"
			return template.BieterEdit(bieter, invalidFields).Render(r.Context(), w)
		}

		if err := write(m.BieterUpdate(bieter)); err != nil {
			invalidFields := bieter.InvalidFields()
			invalidFields["form"] = userError(err)
			return template.BieterEdit(bieter, invalidFields).Render(r.Context(), w)
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return nil

	default:
		http.Error(w, "Fehler", http.StatusMethodNotAllowed)
		return nil
	}
}

func (s server) handleVertrag(w http.ResponseWriter, r *http.Request) error {
	user, err := user.FromRequest(r, []byte(s.cfg.Secred))
	if err != nil || user.BieterID == 0 {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return nil
	}

	m, done := s.model.ForReading()
	defer done()

	bieter, ok := m.Bieter[user.BieterID]

	if user.IsAnonymous() || !ok {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return nil
	}

	vertrag, err := pdf.Bietervertrag(s.cfg.BaseURL, bieter)
	if err != nil {
		return err
	}

	_, err = w.Write(vertrag)
	return err
}

func parseBieterEdit(r *http.Request, bieter model.Bieter) (model.Bieter, string) {
	if err := r.ParseForm(); err != nil {
		return bieter, "Formular kann nicht gelesen werden. Versuche es erneut"
	}

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
	return bieter, ""
}

func (s server) adminPage(next func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, err := user.FromRequest(r, []byte(s.cfg.Secred))
		if err == nil && user.IsAdmin {
			return next(w, r)
		}

		// TODO: Make a redirect on htmx requests!

		switch r.Method {
		case http.MethodGet:
			return template.AdminLogin("").Render(r.Context(), w)

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				return template.AdminLogin("Formular kann nicht gelesen werden. Bitte versuche es erneut.").Render(r.Context(), w)
			}

			pw := r.Form.Get("password")
			if pw != s.cfg.AdminToken {
				return template.AdminLogin("Password ist falsch.").Render(r.Context(), w)
			}

			user.IsAdmin = true
			user.SetCookie(w, []byte(s.cfg.Secred))

			return next(w, r)

		default:
			http.Error(w, "Fehler", http.StatusMethodNotAllowed)
			return nil
		}
	}
}

func (s server) handleAdmin(w http.ResponseWriter, r *http.Request) error {
	m, done := s.model.ForReading()
	defer done()

	bieter := adminBieterList(m)
	return template.Admin(m.State, bieter).Render(r.Context(), w)
}

func adminBieterList(m model.Model) []model.Bieter {
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
	return bieter
}

func (s server) handleAdminNew(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.Error(w, "Hier wird nur geupdated", http.StatusMethodNotAllowed)
		return nil
	}

	m, write, done := s.model.ForWriting()
	defer done()

	bieterID, event := m.BieterCreate()
	if err := write(event); err != nil {
		// TODO
		return err
	}

	bieter := m.Bieter[bieterID]
	template.AdminUserTable(adminBieterList(m)).Render(r.Context(), w)
	template.AdminBieterEdit(bieter, bieter.InvalidFields()).Render(r.Context(), w)

	return nil
}

func (s server) handleAdminEdit(w http.ResponseWriter, r *http.Request) error {
	bietID, _ := strconv.Atoi(mux.Vars(r)["id"])

	switch r.Method {
	case http.MethodGet:
		m, done := s.model.ForReading()
		defer done()

		bieter, ok := m.Bieter[bietID]
		if !ok {
			// TODO: fehlermeldung senden
			return nil
		}

		return template.AdminBieterEdit(bieter, bieter.InvalidFields()).Render(r.Context(), w)
	case http.MethodPost:
		m, write, done := s.model.ForWriting()
		defer done()
		bieter, ok := m.Bieter[bietID]
		if !ok {
			// TODO: fehlermeldung senden
			return nil
		}

		bieter, errMsg := parseBieterEdit(r, bieter)
		if errMsg != "" {
			// TODO: fehlermeldung senden
			return nil
			// return template.BieterEdit(bieter, "Formular kann nicht gelesen werden. Versuche es erneut").Render(r.Context(), w)
		}

		if err := write(m.BieterUpdate(bieter)); err != nil {
			// TODO: fehlermeldung senden
			return nil
		}

		template.AdminModalEmpty().Render(r.Context(), w)
		return template.AdminUserTable(adminBieterList(m)).Render(r.Context(), w)

	default:
		http.Error(w, "Fehler", http.StatusMethodNotAllowed)
		return nil
	}
}

func (s server) handleAdminDelete(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		http.Error(w, "Hier wird nur gelöscht", http.StatusMethodNotAllowed)
		return nil
	}

	m, write, done := s.model.ForWriting()
	defer done()

	bietID, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := write(m.BieterDelete(bietID)); err != nil {
		return err
	}

	bieter := adminBieterList(m)
	return template.AdminUserTable(bieter).Render(r.Context(), w)
}

func (s server) handleAdminResetGebot(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		http.Error(w, "Hier wird nur gelöscht", http.StatusMethodNotAllowed)
		return nil
	}

	m, write, done := s.model.ForWriting()
	defer done()

	if err := write(m.ResetGebot()); err != nil {
		return err
	}

	bieter := adminBieterList(m)
	return template.AdminUserTable(bieter).Render(r.Context(), w)
}

func (s server) handleAdminState(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.Error(w, "Hier wird nur geupdated", http.StatusMethodNotAllowed)
		return nil
	}

	m, write, done := s.model.ForWriting()
	defer done()

	if err := r.ParseForm(); err != nil {
		// TODO: show to
		return err
	}
	state := model.StateFromAttr(r.Form.Get("state"))
	if err := write(m.SetState(state)); err != nil {
		// TODO
		return err
	}

	return template.AdminStateSelect(state).Render(r.Context(), w)
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
			log.Printf("Error: %v", err)
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

func userError(err error) string {
	var errValidation sticky.ValidationError
	if errors.As(err, &errValidation) {
		return errValidation.String()
	}

	return "Unbekannter Fehler"
}
