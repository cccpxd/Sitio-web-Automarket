package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"automarket/internal/config"
	"automarket/internal/httpapi"
	"automarket/internal/httpapi/router"
	"automarket/internal/platform/clock"
	"automarket/internal/repository/postgres"
	"automarket/internal/security"
	"automarket/internal/service"
	"automarket/internal/webui"
)

// App is the fully wired application. main and the serverless runtime only
// need the Handler and Close function.
type App struct {
	Handler http.Handler
	Close   func() error
}

// New is the single composition root: it creates adapters, injects them into
// services, and builds the HTTP router.
func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	db, err := postgres.Open(ctx, postgres.Options{
		URL:             cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
		ConnectTimeout:  cfg.DBConnectTimeout,
	})
	if err != nil {
		return nil, err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = db.Close()
		}
	}()

	if err := migrate(ctx, db); err != nil {
		return nil, err
	}

	authService, err := buildAuthService(db, cfg)
	if err != nil {
		return nil, err
	}

	carService, err := buildCarService(db)
	if err != nil {
		return nil, err
	}

	if cfg.AdminEmail != "" {
		created, err := authService.EnsureAdmin(ctx, cfg.AdminName, cfg.AdminEmail, cfg.AdminPassword)
		if err != nil {
			return nil, fmt.Errorf("bootstrap admin: %w", err)
		}
		if created {
			logger.InfoContext(ctx, "bootstrap_admin_created")
		}
	}

	renderer, err := webui.NewRenderer()
	if err != nil {
		return nil, err
	}
	staticFS, err := webui.StaticFS()
	if err != nil {
		return nil, err
	}

	if err := seedCars(ctx, carService); err != nil {
		return nil, fmt.Errorf("seed cars: %w", err)
	}

	engine, err := router.New(router.Dependencies{
		Auth:     authService,
		Admin:    authService,
		Cars:     carService,
		Renderer: renderer,
		StaticFS: staticFS,
		Logger:   logger,
		Cookie: httpapi.CookieConfig{
			Name:     cfg.CookieName,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			TTL:      cfg.SessionTTL,
		},
		HealthCheck: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			return db.PingContext(pingCtx)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("build router: %w", err)
	}

	cleanup = false
	return &App{
		Handler: engine,
		Close:   db.Close,
	}, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := postgres.Migrate(migrationCtx, db); err != nil {
		return err
	}
	return nil
}

func seedCars(ctx context.Context, cars *service.CarService) error {
	existing, err := cars.ListCars(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	demo := []service.CreateCarInput{
		{Brand: "Toyota", Model: "Corolla", Year: 2020, Description: "Sedan confiable, un solo dueno, mantenimiento al dia.", Price: "18,500"},
		{Brand: "Honda", Model: "Civic", Year: 2019, Description: "Excelente rendimiento de combustible y equipo completo.", Price: "17,200"},
		{Brand: "Mazda", Model: "CX-5", Year: 2021, Description: "SUV familiar con camara de reversa y sensores.", Price: "24,900"},
		{Brand: "Nissan", Model: "Versa", Year: 2018, Description: "Auto economico ideal para la ciudad.", Price: "12,800"},
		{Brand: "Ford", Model: "Ranger", Year: 2022, Description: "Pickup 4x4 lista para trabajo y aventura.", Price: "32,500"},
		{Brand: "Chevrolet", Model: "Spark", Year: 2017, Description: "Compacto practico, bajo consumo y facil de estacionar.", Price: "9,900"},
	}

	for _, input := range demo {
		if _, err := cars.CreateCar(ctx, "", input); err != nil {
			return err
		}
	}
	return nil
}

func buildCarService(db *sql.DB) (*service.CarService, error) {
	idGen := security.NewIDGenerator()
	carService, err := service.NewCarService(service.CarServiceDeps{
		Cars:  postgres.NewCarRepository(db),
		IDs:   idGen,
		Clock: clock.New(),
	})
	if err != nil {
		return nil, fmt.Errorf("build car service: %w", err)
	}
	return carService, nil
}

func buildAuthService(db *sql.DB, cfg config.Config) (*service.AuthService, error) {
	hasher, err := security.NewPBKDF2Hasher(security.DefaultIterations)
	if err != nil {
		return nil, err
	}

	authService, err := service.NewAuthService(service.AuthServiceDeps{
		Users:      postgres.NewUserRepository(db),
		Sessions:   postgres.NewSessionRepository(db),
		Hasher:     hasher,
		Tokens:     security.NewTokenIssuer(),
		IDs:        security.NewIDGenerator(),
		Clock:      clock.New(),
		SessionTTL: cfg.SessionTTL,
		DummyHash:  security.DummyPasswordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("build auth service: %w", err)
	}
	return authService, nil
}
