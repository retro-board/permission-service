package service

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugMiddleware "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog"
	"github.com/retro-board/permission-service/internal/config"
	"github.com/retro-board/permission-service/internal/permissions"
)

type Service struct {
	Config *config.Config
}

func CheckAPIKey(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != os.Getenv("KEY_SERVICE_KEY") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (s *Service) Start() error {
	bugLog.Local().Info("Starting Key")

	logger := httplog.NewLogger("permission-service", httplog.Options{
		JSON: true,
	})

	allowedOrigins := []string{
		"http://localhost:8080",
		"https://retro-board.it",
		"https://*.retro-board.it",
	}
	if s.Config.Development {
		allowedOrigins = append(allowedOrigins, "http://*")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(c.Handler)
		r.Use(bugMiddleware.BugFixes)
		r.Use(httplog.RequestLogger(logger))

		if !s.Config.Development {
			r.Use(CheckAPIKey)
		}

		r.Post("/", permissions.NewPermissions(s.Config).CreateHandler)
		r.Get("/", permissions.NewPermissions(s.Config).GetHandler)
		r.Put("/", permissions.NewPermissions(s.Config).UpdateHandler)
		r.Get("/{permission}", permissions.NewPermissions(s.Config).ValidateHandler)
	})

	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)

	bugLog.Local().Infof("listening on %d\n", s.Config.Local.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Local.Port), r); err != nil {
		return bugLog.Errorf("port failed: %+v", err)
	}

	return nil
}
