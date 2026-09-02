// HTTP server bootstrap. Connects infrastructure (PostgreSQL, MongoDB, Redis,
// object storage), starts the Gin app, and wires graceful shutdown.
//
// Note: Asynq WORKERS are started separately (cmd/worker) so the API and the
// worker fleet can scale independently.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/audit"
	"github.com/raunak0400/metrology-backend/internal/auth"
	"github.com/raunak0400/metrology-backend/internal/authtoken"
	"github.com/raunak0400/metrology-backend/internal/compliance"
	"github.com/raunak0400/metrology-backend/internal/config"
	"github.com/raunak0400/metrology-backend/internal/dashboard"
	"github.com/raunak0400/metrology-backend/internal/datasets"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/declarations"
	"github.com/raunak0400/metrology-backend/internal/evidence"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/ml"
	"github.com/raunak0400/metrology-backend/internal/products"
	"github.com/raunak0400/metrology-backend/internal/reports"
	"github.com/raunak0400/metrology-backend/internal/response"
	"github.com/raunak0400/metrology-backend/internal/roles"
	"github.com/raunak0400/metrology-backend/internal/rules"
	"github.com/raunak0400/metrology-backend/internal/scans"
	"github.com/raunak0400/metrology-backend/internal/users"
	"github.com/raunak0400/metrology-backend/internal/violations"
	"github.com/raunak0400/metrology-backend/pkg/logger"
	"github.com/raunak0400/metrology-backend/pkg/mongodb"
	"github.com/raunak0400/metrology-backend/pkg/postgres"
	"github.com/raunak0400/metrology-backend/pkg/redisclient"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	logger.New(cfg.LogLevel, cfg.GoEnv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("PostgreSQL connected")

	mongoClient, mongoDB, err := mongodb.Connect(ctx, cfg.MongoDBURI, "legal_metrology_ml")
	if err != nil {
		slog.Error("failed to connect to mongodb", "err", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(ctx)
	slog.Info("MongoDB connected")

	mlRepo := ml.NewRepository(mongoDB)
	if err := mlRepo.EnsureIndexes(ctx); err != nil {
		slog.Error("failed to ensure mongodb indexes", "err", err)
		os.Exit(1)
	}

	redisClient, err := redisclient.Connect(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("Redis connected")

	objectStorage, err := storage.New(storage.Config{
		Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey, Bucket: cfg.S3Bucket, UseSSL: cfg.S3UseSSL, ForcePathStyle: cfg.S3ForcePathStyle,
	})
	if err != nil {
		slog.Error("failed to create storage client", "err", err)
		os.Exit(1)
	}
	if err := objectStorage.EnsureBucket(ctx); err != nil {
		slog.Warn("could not ensure storage bucket", "err", err)
	}

	asynqOpt, err := asynqRedisOpt(cfg.RedisURL)
	if err != nil {
		slog.Error("invalid redis url for asynq", "err", err)
		os.Exit(1)
	}
	asynqClient := asynq.NewClient(asynqOpt)
	defer asynqClient.Close()

	deps := &app.Deps{
		Config:    cfg,
		Pool:      pool,
		Queries:   db.New(pool),
		Mongo:     mongoDB,
		MLRepo:    mlRepo,
		Redis:     redisClient,
		Storage:   objectStorage,
		MLClient:  ml.NewClient(cfg.MLServiceURL, cfg.MLServiceAPIKey, cfg.MLServiceTimeout),
		Asynq:     asynqClient,
		Tokenizer: authtoken.NewTokenizer(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL),
	}

	router := buildRouter(cfg, deps)

	addr := ":" + strconv.Itoa(cfg.Port)
	server := &http.Server{Addr: addr, Handler: router}
	go func() {
		slog.Info("API listening", "addr", "http://localhost:"+strconv.Itoa(cfg.Port)+cfg.APIPrefix)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
}

func asynqRedisOpt(redisURL string) (asynq.RedisConnOpt, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return asynq.RedisClientOpt{Addr: opt.Addr, Password: opt.Password, DB: opt.DB}, nil
}

func buildRouter(cfg *config.Config, deps *app.Deps) *gin.Engine {
	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "x-request-id"},
		AllowCredentials: true,
	}))
	router.Use(middleware.RequestLogger())

	router.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})
	router.GET("/ready", func(c *gin.Context) {
		if err := deps.Pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": gin.H{"code": "SERVICE_UNAVAILABLE", "message": "postgres not ready"}})
			return
		}
		if err := deps.Redis.Ping(c.Request.Context()).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": gin.H{"code": "SERVICE_UNAVAILABLE", "message": "redis not ready"}})
			return
		}
		if err := deps.Mongo.Client().Ping(c.Request.Context(), nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": gin.H{"code": "SERVICE_UNAVAILABLE", "message": "mongodb not ready"}})
			return
		}
		if err := deps.Storage.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": gin.H{"code": "SERVICE_UNAVAILABLE", "message": "object storage not ready"}})
			return
		}
		response.OK(c, gin.H{"status": "ready"})
	})

	v1 := router.Group(cfg.APIPrefix)
	v1.Use(middleware.RateLimit(deps.Redis, middleware.RateLimiterConfig{Window: cfg.RateLimitWindow, Max: cfg.RateLimitMax, Prefix: "global"}))
	v1.Use(middleware.ErrorHandler(cfg.IsProd()))

	v1.GET("/", func(c *gin.Context) {
		response.OK(c, gin.H{"name": "Legal Metrology Compliance Checker API", "version": "v1"})
	})

	auth.RegisterRoutes(v1, deps)
	users.RegisterRoutes(v1, deps)
	roles.RegisterRoutes(v1, deps)
	products.RegisterRoutes(v1, deps)
	rules.RegisterRoutes(v1, deps)
	dashboard.RegisterRoutes(v1, deps)
	audit.RegisterRoutes(v1, deps)

	inspSvc := inspections.RegisterRoutes(v1, deps)
	scans.RegisterRoutes(v1, deps, inspSvc)
	violSvc := violations.RegisterRoutes(v1, deps, inspSvc)
	complianceSvc := compliance.RegisterRoutes(v1, deps, inspSvc, violSvc)
	evidence.RegisterRoutes(v1, deps, inspSvc)
	declarations.RegisterRoutes(v1, deps, inspSvc, complianceSvc)
	reports.RegisterRoutes(v1, deps, inspSvc)
	datasets.RegisterRoutes(v1, deps)

	router.NoRoute(middleware.NotFoundHandler)
	return router
}
