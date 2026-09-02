// Worker fleet bootstrap. Connects the infrastructure the workers need
// (PostgreSQL, MongoDB, object storage, the ML service client; Redis via
// Asynq) and starts the Asynq server processing the scan, ocr, compliance and
// report queues.
//
// Run separately from the API: `go run ./cmd/worker`.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/raunak0400/metrology-backend/internal/compliance"
	"github.com/raunak0400/metrology-backend/internal/config"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/declarations"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/ml"
	"github.com/raunak0400/metrology-backend/internal/reports"
	"github.com/raunak0400/metrology-backend/internal/scans"
	"github.com/raunak0400/metrology-backend/internal/taskqueue"
	"github.com/raunak0400/metrology-backend/internal/workers"
	"github.com/raunak0400/metrology-backend/pkg/logger"
	"github.com/raunak0400/metrology-backend/pkg/mongodb"
	"github.com/raunak0400/metrology-backend/pkg/postgres"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	logger.New(cfg.LogLevel, cfg.GoEnv)

	ctx := context.Background()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	mongoClient, mongoDB, err := mongodb.Connect(ctx, cfg.MongoDBURI, "legal_metrology_ml")
	if err != nil {
		slog.Error("failed to connect to mongodb", "err", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(ctx)

	objectStorage, err := storage.New(storage.Config{
		Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey, Bucket: cfg.S3Bucket, UseSSL: cfg.S3UseSSL, ForcePathStyle: cfg.S3ForcePathStyle,
	})
	if err != nil {
		slog.Error("failed to create storage client", "err", err)
		os.Exit(1)
	}

	queries := db.New(pool)
	mlClient := ml.NewClient(cfg.MLServiceURL, cfg.MLServiceAPIKey, cfg.MLServiceTimeout)
	mlRepo := ml.NewRepository(mongoDB)
	if err := mlRepo.EnsureIndexes(ctx); err != nil {
		slog.Error("failed to ensure mongodb indexes", "err", err)
		os.Exit(1)
	}

	inspSvc := inspections.NewService(queries)
	declSvc := declarations.NewService(queries, inspSvc)

	asynqOpt, err := asynqRedisOpt(cfg.RedisURL)
	if err != nil {
		slog.Error("invalid redis url for asynq", "err", err)
		os.Exit(1)
	}
	asynqClient := asynq.NewClient(asynqOpt)
	defer asynqClient.Close()

	scanPipeline := scans.NewPipeline(queries, objectStorage, mlClient, mlRepo, declSvc, asynqClient)
	complianceSvc := compliance.NewService(queries, asynqClient, inspSvc)
	reportSvc := reports.NewService(queries, objectStorage, asynqClient, inspSvc)

	mux := workers.NewMux(workers.Deps{
		Queries: queries, Storage: objectStorage, ScanPipeline: scanPipeline,
		ComplianceSvc: complianceSvc, ReportSvc: reportSvc, MLClient: mlClient, MLRepo: mlRepo,
	})

	server := asynq.NewServer(asynqOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			taskqueue.QueueCompliance: 6,
			taskqueue.QueueScan:       4,
			taskqueue.QueueOCR:        4,
			taskqueue.QueueReport:     2,
		},
	})

	slog.Info("starting worker server", "queues", []string{taskqueue.QueueScan, taskqueue.QueueOCR, taskqueue.QueueCompliance, taskqueue.QueueReport})
	if err := server.Run(mux); err != nil {
		slog.Error("worker server stopped", "err", err)
		os.Exit(1)
	}
}

func asynqRedisOpt(redisURL string) (asynq.RedisConnOpt, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return asynq.RedisClientOpt{Addr: opt.Addr, Password: opt.Password, DB: opt.DB}, nil
}
