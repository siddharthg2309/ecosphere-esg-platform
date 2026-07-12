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
	environmentalhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/adapter/http"
	environmentalpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/adapter/postgres"
	carbonapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/app"
	carbonport "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/port"
	goalapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/goal/app"
	gamehttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/adapter/http"
	gamepg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/adapter/postgres"
	gameapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/gamification/app"
	govhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/adapter/http"
	govpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/adapter/postgres"
	govapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/governance/app"
	identityhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/adapter/http"
	identitypg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/adapter/postgres"
	identityapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/app"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/identity/domain"
	notifhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/adapter/http"
	notifpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/adapter/postgres"
	notifapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/notification/app"
	departmenthttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/adapter/http"
	departmentpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/adapter/postgres"
	departmentapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/department/app"
	masterhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/adapter/http"
	masterpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/adapter/postgres"
	masterapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/app"
	reporthttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/adapter/http"
	reportexport "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/adapter/export"
	reportpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/adapter/postgres"
	reportapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/reporting/app"
	scorehttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/scoring/adapter/http"
	scorepg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/scoring/adapter/postgres"
	scoreapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/scoring/app"
	socialhttp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/adapter/http"
	socialpg "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/adapter/postgres"
	socialapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/social/app"
	platformai "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/ai"
	platformauth "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/config"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/db/sqlc"
	platformemail "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/email"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/httpserver"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/scheduler"
	platformsettings "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/settings"
	platformstorage "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/storage"
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
	masterRepo := masterpg.New(pool)
	masterService := masterapp.New(masterRepo, platformemail.New(cfg.SMTPAddr), bus)
	masterHandler := masterhttp.New(masterService)
	flags := platformsettings.New(masterRepo, bus)

	// Phase 2 — Environmental
	environmentalRepo := environmentalpg.New(pool)
	goalRepo := environmentalpg.NewGoals(pool)
	carbonService := carbonapp.New(environmentalRepo, flags, bus)
	goalService := goalapp.New(goalRepo, bus)
	objectStorage, err := platformstorage.NewMinIO(ctx, cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL)
	if err != nil {
		log.Fatal(err)
	}
	aiGW := platformai.NewGateway(cfg.OpenRouterAPIKey, cfg.OpenRouterModel, cfg.AIFixtureMode)
	var aiGateway carbonport.AIGateway
	if aiGW.UseFixture() {
		aiGateway = platformai.Fixture{}
	} else {
		aiGateway = aiGW.Live()
	}
	ingestService := carbonapp.NewIngest(objectStorage, aiGateway, flags, cfg.AIConfidence)
	environmentalHandler := environmentalhttp.New(carbonService, goalService, ingestService)
	aiHandler := platformai.NewHandler(aiGW)

	// Phase 3 — Social
	socialRepo := socialpg.New(pool)
	socialService := socialapp.New(
		socialpg.ActivityRepo{Repository: socialRepo},
		socialpg.ParticipationRepo{Repository: socialRepo},
		socialpg.TrainingRepo{Repository: socialRepo},
		socialRepo,
		flags,
		socialRepo,
		bus,
	)
	socialHandler := socialhttp.New(socialService)

	// Phase 3 — Gamification
	gameRepo := gamepg.New(pool)
	gameService := gameapp.New(
		gamepg.ChallengeRepo{Repository: gameRepo},
		gamepg.ChallengeParticipationRepo{Repository: gameRepo},
		gamepg.RewardRepo{Repository: gameRepo},
		gamepg.BadgeRepo{Repository: gameRepo},
		gamepg.UserBalanceRepo{Repository: gameRepo},
		gamepg.LeaderboardRepo{Repository: gameRepo},
		flags,
		bus,
	)
	gameHandler := gamehttp.New(gameService)

	// Phase 4 — Governance
	govRepo := govpg.New(pool)
	govService := govapp.New(
		govpg.AuditRepo{Repository: govRepo},
		govpg.IssueRepo{Repository: govRepo},
		govpg.AckRepo{Repository: govRepo},
		govpg.PolicyRepo{Repository: govRepo},
		govpg.BundleRepo{Repository: govRepo},
		bus,
	)
	govHandler := govhttp.New(govService)

	// Phase 4 — Notifications
	emailSender := platformemail.New(cfg.SMTPAddr)
	notifService := notifapp.New(
		notifpg.NewStore(pool),
		notifpg.NewPrefs(pool),
		notifpg.NewUserEmail(pool),
		platformemail.MailAdapter{Sender: emailSender},
		platformemail.TemplateAdapter{Templates: platformemail.NewTemplates()},
	)
	notifService.Wire(bus)
	notifHandler := notifhttp.New(notifService)
	sched := scheduler.New(pool, govService, notifService, bus)
	sched.Start(ctx)

	// Phase 5 — Scoring & Reports
	scoreService := scoreapp.New(scorepg.NewMetrics(pool), scorepg.NewStore(pool), bus)
	scoreHandler := scorehttp.New(scoreService)
	reportService := reportapp.New(
		reportpg.NewStore(pool),
		reportpg.NewData(pool),
		scoreService,
		aiGW,
		reportexport.New(),
	)
	reportHandler := reporthttp.New(reportService)

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
	router.Route("/carbon", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Post("/ingest", environmentalHandler.Ingest)
		r.Post("/transactions", environmentalHandler.CreateTransaction)
		r.Get("/transactions", environmentalHandler.ListTransactions)
		r.Get("/summary", environmentalHandler.Summary)
		r.With(platformauth.RequireRole(domain.RoleDeptHead)).Post("/transactions/{id}/verify", environmentalHandler.Verify)
	})
	router.Route("/goals", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", environmentalHandler.ListGoals)
		r.With(platformauth.RequireRole(domain.RoleDeptHead, domain.RoleAdmin)).Post("/", environmentalHandler.CreateGoal)
		r.With(platformauth.RequireRole(domain.RoleDeptHead, domain.RoleAdmin)).Put("/{id}", environmentalHandler.UpdateGoal)
	})

	// Phase 3 — Social
	router.Route("/csr", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/activities", socialHandler.ListActivities)
		r.Get("/activities/{id}", socialHandler.GetActivity)
		r.With(platformauth.RequireRole(domain.RoleAdmin, domain.RoleDeptHead)).Post("/activities", socialHandler.CreateActivity)
		r.Post("/participations", socialHandler.JoinActivity)
		r.Get("/participations", socialHandler.ListParticipations)
		r.With(platformauth.RequireRole(domain.RoleAdmin, domain.RoleDeptHead)).Post("/participations/{id}/approve", socialHandler.ApproveParticipation)
		r.With(platformauth.RequireRole(domain.RoleAdmin, domain.RoleDeptHead)).Post("/participations/{id}/reject", socialHandler.RejectParticipation)
	})
	router.With(platformauth.Authenticate([]byte(cfg.JWTSecret))).Get("/diversity", socialHandler.Diversity)
	router.Route("/trainings", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", socialHandler.ListTrainings)
		r.With(platformauth.RequireRole(domain.RoleAdmin)).Post("/", socialHandler.CreateTraining)
		r.Post("/{id}/complete", socialHandler.CompleteTraining)
	})

	// Phase 3 — Gamification
	router.Route("/challenges", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", gameHandler.ListChallenges)
		r.Get("/status-counts", gameHandler.StatusCounts)
		r.Get("/transitions", gameHandler.Transitions)
		r.With(platformauth.RequireRole(domain.RoleAdmin, domain.RoleDeptHead)).Post("/", gameHandler.CreateChallenge)
		r.With(platformauth.RequireRole(domain.RoleAdmin, domain.RoleDeptHead)).Put("/{id}/transition", gameHandler.Transition)
		r.Post("/{id}/participate", gameHandler.Participate)
	})
	router.Route("/challenge-participations", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", gameHandler.ListParticipations)
		r.With(platformauth.RequireRole(domain.RoleAdmin, domain.RoleDeptHead)).Post("/{id}/approve", gameHandler.ApproveParticipation)
		r.With(platformauth.RequireRole(domain.RoleAdmin, domain.RoleDeptHead)).Post("/{id}/reject", gameHandler.RejectParticipation)
	})
	router.With(platformauth.Authenticate([]byte(cfg.JWTSecret))).Get("/leaderboard", gameHandler.Leaderboard)
	router.With(platformauth.Authenticate([]byte(cfg.JWTSecret))).Get("/me/balance", gameHandler.Balance)
	router.Route("/game-rewards", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", gameHandler.ListRewards)
		r.Post("/{id}/redeem", gameHandler.Redeem)
	})
	router.With(platformauth.Authenticate([]byte(cfg.JWTSecret))).Get("/game-badges", gameHandler.ListBadges)

	// Phase 4 — Governance routes (avoid clashing with master CRUD /policies/{id})
	router.Route("/governance", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/stats", govHandler.Stats)
		r.Get("/policies", govHandler.ListGovernancePolicies)
		r.Get("/acknowledgements", govHandler.ListAcknowledgements)
		r.Get("/unacknowledged", govHandler.Unacknowledged)
		r.Post("/policies/{id}/acknowledge", govHandler.Acknowledge)
	})
	router.Route("/audits", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", govHandler.ListAudits)
		r.With(platformauth.RequireRole(domain.RoleAuditor, domain.RoleAdmin)).Post("/", govHandler.CreateAudit)
		r.With(platformauth.RequireRole(domain.RoleAuditor, domain.RoleAdmin)).Get("/department/{id}/bundle", govHandler.DepartmentBundle)
		r.Get("/{id}", govHandler.GetAudit)
	})
	router.Route("/compliance-issues", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", govHandler.ListIssues)
		r.Get("/{id}", govHandler.GetIssue)
		r.With(platformauth.RequireRole(domain.RoleAuditor, domain.RoleDeptHead, domain.RoleAdmin)).Post("/", govHandler.RaiseIssue)
		r.With(platformauth.RequireRole(domain.RoleAuditor, domain.RoleDeptHead, domain.RoleAdmin)).Put("/{id}", govHandler.UpdateIssue)
	})
	router.Route("/notifications", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/", notifHandler.List)
		r.Post("/{id}/read", notifHandler.MarkRead)
	})

	// Phase 5 — Scores, reports, AI evidence assist
	router.Route("/scores", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Get("/departments", scoreHandler.Departments)
		r.Get("/overall", scoreHandler.Overall)
		r.With(platformauth.RequireRole(domain.RoleAdmin)).Post("/recompute", scoreHandler.Recompute)
	})
	router.Route("/reports", func(r chi.Router) {
		r.Use(platformauth.Authenticate([]byte(cfg.JWTSecret)))
		r.Post("/generate", reportHandler.Generate)
		r.Get("/{id}", reportHandler.Get)
		r.Get("/{id}/export", reportHandler.Export)
	})
	router.With(platformauth.Authenticate([]byte(cfg.JWTSecret))).Post("/ai/evidence-review", aiHandler.ReviewEvidence)

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
