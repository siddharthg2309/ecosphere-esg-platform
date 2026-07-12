package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	identityhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/adapter/http"
	identitypg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/adapter/postgres"
	identityapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	departmenthttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/adapter/http"
	departmentpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/adapter/postgres"
	departmentapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/app"
	masterhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/adapter/http"
	masterpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/adapter/postgres"
	masterapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/app"
	platformauth "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/config"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db/sqlc"
	platformemail "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/email"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	identityRepo := identitypg.New(sqlc.New(pool))
	identityService := identityapp.New(identityRepo, []byte(cfg.JWTSecret), cfg.AccessTTL, cfg.RefreshTTL)
	identityHandler := identityhttp.New(identityService)
	departmentService := departmentapp.New(departmentpg.New(sqlc.New(pool)))
	departmentHandler := departmenthttp.New(departmentService)
	bus := events.NewInProcess()
	masterService := masterapp.New(masterpg.New(pool), platformemail.New(cfg.SMTPAddr), bus)
	masterHandler := masterhttp.New(masterService)

	router := chi.NewRouter()
	router.Use(httpserver.Recover, httpserver.RequestID, httpserver.Logger, httpserver.CORS(cfg.CORSOrigin), middleware.StripSlashes)
	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	for _, entityName := range []string{"categories", "emission-factors", "products", "policies", "badges", "rewards", "employees"} {
		entityName := entityName
		router.Route("/"+entityName, func(r chi.Router) {
			r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
			r.Get("/", masterHandler.List(entityName))
			r.Get("/{id}", masterHandler.Get(entityName))
			r.Group(func(r chi.Router) {
				r.Use(platformauth.RequireRole(domain.RoleAdmin))
				r.Post("/", masterHandler.Create(entityName))
				r.Put("/{id}", masterHandler.Update(entityName))
				r.Delete("/{id}", masterHandler.Delete(entityName))
			})
		})
	}
	router.Route("/settings", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/esg-config", masterHandler.GetConfig)
		r.Get("/notification-preferences", masterHandler.GetPreferences)
		r.Group(func(r chi.Router) {
			r.Use(platformauth.RequireRole(domain.RoleAdmin))
			r.Put("/esg-config", masterHandler.SaveConfig)
			r.Put("/notification-preferences", masterHandler.SavePreferences)
		})
	})
	router.Post("/auth/login", identityHandler.Login)
	router.Post("/auth/refresh", identityHandler.Refresh)
	router.With(platformauth.Authenticate([]byte(cfg.JWTSecret))).Get("/me", identityHandler.Me)
	router.Route("/departments", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", departmentHandler.List)
		r.Get("/{id}", departmentHandler.Get)
		r.Group(func(r chi.Router) {
			r.Use(platformauth.RequireRole(domain.RoleAdmin))
			r.Post("/", departmentHandler.Create)
			r.Put("/{id}", departmentHandler.Update)
			r.Delete("/{id}", departmentHandler.Deactivate)
		})
	})

	server := &http.Server{Addr: cfg.Addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("EcoSphere API listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
