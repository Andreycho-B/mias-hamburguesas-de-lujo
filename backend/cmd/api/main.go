package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"mias-backend/internal/config"
	"mias-backend/internal/database/db"
	"mias-backend/internal/handler"
	"mias-backend/internal/middleware"
	"mias-backend/internal/service"
)

func main() {
	// 1. Logger Estructurado con log/slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Iniciando Mías Hamburguesas de Lujo API...")

	// 2. Cargar Configuración
	cfg := config.Load()

	// 3. Pool de Conexiones PostgreSQL con pgxpool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		slog.Warn("No se pudo parsear DATABASE_URL, iniciando con fallback local", slog.Any("error", err))
	}

	var pool *pgxpool.Pool
	if poolConfig != nil {
		poolConfig.MaxConns = 25
		poolConfig.MinConns = 5
		poolConfig.MaxConnLifetime = 1 * time.Hour
		poolConfig.MaxConnIdleTime = 30 * time.Minute

		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			slog.Warn("No se pudo conectar al pool de Postgres, operando en modo degradado", slog.Any("error", err))
		} else {
			slog.Info("Conexión a PostgreSQL 16 establecida exitosamente")
			defer pool.Close()
		}
	}

	var queries db.Querier
	if pool != nil {
		queries = db.New(pool)
	}

	// 4. Inicializar Servicios y Hubs
	sseHub := service.NewSSEHub()
	authService := service.NewAuthService(queries, cfg.JWTSecret)
	catalogService := service.NewCatalogService(queries)
	orderService := service.NewOrderService(queries, sseHub)
	stampService := service.NewStampService(queries, sseHub)

	// 5. Inicializar Handlers HTTP
	authHandler := handler.NewAuthHandler(authService)
	catalogHandler := handler.NewCatalogHandler(catalogService)
	orderHandler := handler.NewOrderHandler(orderService, queries)
	stampHandler := handler.NewStampHandler(stampService, queries)
	adminHandler := handler.NewAdminHandler(queries)
	sseHandler := handler.NewSSEHandler(sseHub)

	// 6. Router Principal Chi
	r := chi.NewRouter()

	// Middlewares Globales
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.StructuredLogger)

	// CORS de Grado Bancario
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL, "http://localhost:5173", "http://localhost:4173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Idempotency-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Endpoints de Liveness y Readiness
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"UP"}`))
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			http.Error(w, `{"status":"DEGRADED","database":"DISCONNECTED"}`, http.StatusServiceUnavailable)
			return
		}
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, `{"status":"DOWN","database":"PING_FAILED"}`, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"READY","database":"CONNECTED"}`))
	})

	// API v1
	r.Route("/api/v1", func(api chi.Router) {
		// Catálogo Gastronómico (Público)
		api.Get("/catalog", catalogHandler.GetPublicMenu)

		// Autenticación
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/staff-login", authHandler.StaffLogin)
			auth.Post("/oauth", authHandler.OAuthLogin)
			auth.Post("/refresh", authHandler.RefreshSession)
			auth.Post("/logout", authHandler.Logout)
			auth.With(middleware.RequireAuth).Get("/me", authHandler.GetMe)
		})

		// Pedidos y Comandas
		api.Route("/orders", func(orders chi.Router) {
			orders.Post("/", orderHandler.CreateOrder)
			orders.Get("/{code}", orderHandler.GetOrderByCode)

			// Sub-rutas exclusivas de Personal (Caja / Meseros)
			orders.Group(func(staff chi.Router) {
				staff.Use(middleware.RequireAuth)
				staff.Use(middleware.RequireRole("waiter", "cashier", "admin"))

				staff.Get("/staff/pending-in-store", orderHandler.ListPendingInStore)
				staff.Get("/staff/active-delivery", orderHandler.ListActiveDelivery)
				staff.Patch("/{id}/pay", orderHandler.MarkAsPaid)
				staff.Patch("/{id}/delivery-status", orderHandler.UpdateDeliveryStatus)
			})
		})

		// Fidelización y Sellos
		api.Route("/stamps", func(stamps chi.Router) {
			stamps.Group(func(client chi.Router) {
				client.Use(middleware.RequireAuth)
				client.Get("/balance", stampHandler.GetBalance)
				client.Post("/request", stampHandler.RequestStamp)
				client.Post("/redeem", stampHandler.RequestRedemption)
			})

			// Sub-rutas exclusivas de Personal
			stamps.Group(func(staff chi.Router) {
				staff.Use(middleware.RequireAuth)
				staff.Use(middleware.RequireRole("waiter", "cashier", "admin"))

				staff.Get("/staff/pending", stampHandler.ListPendingStaff)
				staff.Patch("/staff/{id}/review", stampHandler.ReviewStaff)
				staff.Get("/staff/pending-redemptions", stampHandler.ListPendingRedemptions)
				staff.Patch("/staff/redemptions/{id}/approve", stampHandler.ApproveRedemption)
			})
		})

		// Server-Sent Events para Personal
		api.With(middleware.RequireAuth).Get("/events/staff", sseHandler.StreamStaffEvents)

		// Panel Ejecutivo de Administración
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.RequireAuth)
			admin.Use(middleware.RequireRole("admin"))

			admin.Get("/metrics", adminHandler.GetMetrics)
			admin.Get("/audit-logs", adminHandler.ListAuditLogs)
			admin.Post("/staff", adminHandler.CreateStaffUser)
			admin.Patch("/catalog/products/{id}/stock", catalogHandler.ToggleStock)
		})
	})

	// 7. Worker Ligero de Fondo (Expiración de Solicitudes cada 2 minutos)
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if queries != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := queries.ExpirePendingStampRequests(ctx); err != nil {
					slog.Debug("Evaluación periódica de expiración", slog.Any("error", err))
				}
				cancel()
			}
		}
	}()

	// 8. Servidor HTTP y Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Servidor HTTP escuchando", slog.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error en servidor HTTP", slog.Any("error", err))
		}
	}()

	// Captura de Señales de Parada
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Apagando servidor de forma ordenada (Graceful Shutdown)...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Forzando apagado del servidor", slog.Any("error", err))
	}
	slog.Info("Servidor finalizado limpiamente")
}
