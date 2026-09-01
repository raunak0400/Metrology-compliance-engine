// Package app holds the shared dependency container injected into every
// module's route registration function. It exists so main.go wires
// infrastructure exactly once and every module receives the same, already
// health-checked clients.
package app

import (
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/raunak0400/metrology-backend/internal/authtoken"
	"github.com/raunak0400/metrology-backend/internal/config"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/ml"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

type Deps struct {
	Config    *config.Config
	Pool      *pgxpool.Pool
	Queries   *db.Queries
	Mongo     *mongo.Database
	MLRepo    *ml.Repository
	Redis     *redis.Client
	Storage   storage.ObjectStorage
	MLClient  *ml.Client
	Asynq     *asynq.Client
	Tokenizer *authtoken.Tokenizer
}
