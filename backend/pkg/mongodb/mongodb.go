// Package mongodb wires the MongoDB connection used to store flexible, raw
// OCR/CV/ML documents. MongoDB is NEVER the source of truth for compliance
// results — PostgreSQL is. See internal/ml for the collections stored here.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect dials MongoDB and verifies connectivity with a ping. The returned
// *mongo.Database is scoped to the database named in the connection URI
// (falling back to dbNameFallback when the URI has no path segment).
func Connect(ctx context.Context, uri string, dbNameFallback string) (*mongo.Client, *mongo.Database, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(uri).SetServerSelectionTimeout(10*time.Second))
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongodb: %w", err)
	}

	if err := client.Ping(connectCtx, nil); err != nil {
		return nil, nil, fmt.Errorf("ping mongodb: %w", err)
	}

	dbName := client.Database("").Name()
	if dbName == "" {
		dbName = dbNameFallback
	}

	return client, client.Database(dbNameFallback), nil
}
