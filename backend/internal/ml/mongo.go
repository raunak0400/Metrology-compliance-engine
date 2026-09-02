// Mongo persistence for raw ML/OCR/CV documents. Intentionally schema-loose:
// it captures the raw, evolving output of the ML models (OCR blocks, bounding
// boxes, confidences, detected regions, preprocessing metadata, field
// extractions). PostgreSQL stores the normalized, authoritative data; these
// documents are the evidence trail behind it.
package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ocrCollectionName    = "ocr_documents"
	resultCollectionName = "ml_results"
)

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

// EnsureIndexes creates the indexes both collections are queried by
// (inspection_id, image_id, created_at). Safe to call on every process start —
// index creation is idempotent (a no-op if the index already exists with the
// same spec).
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	byInspectionImageCreated := []mongo.IndexModel{
		{Keys: bson.D{{Key: "inspectionId", Value: 1}}},
		{Keys: bson.D{{Key: "imageId", Value: 1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
	}

	for _, coll := range []string{ocrCollectionName, resultCollectionName} {
		if _, err := r.db.Collection(coll).Indexes().CreateMany(ctx, byInspectionImageCreated, options.CreateIndexes()); err != nil {
			return fmt.Errorf("ensure indexes on %s: %w", coll, err)
		}
	}
	return nil
}

type ocrDocument struct {
	InspectionID  string                 `bson:"inspectionId"`
	ImageID       string                 `bson:"imageId"`
	ModelVersion  string                 `bson:"modelVersion"`
	Blocks        []OcrBlock             `bson:"blocks"`
	Preprocessing map[string]interface{} `bson:"preprocessing,omitempty"`
	CreatedAt     time.Time              `bson:"createdAt"`
}

// SaveOcrDocument persists raw OCR output and returns the Mongo document id
// (stored back on InspectionImage.mongo_ocr_doc_id in PostgreSQL).
func (r *Repository) SaveOcrDocument(ctx context.Context, result OcrResult) (string, error) {
	createdAt, err := time.Parse(time.RFC3339, result.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}
	doc := ocrDocument{
		InspectionID:  result.InspectionID,
		ImageID:       result.ImageID,
		ModelVersion:  result.ModelVersion,
		Blocks:        result.Blocks,
		Preprocessing: result.Preprocessing,
		CreatedAt:     createdAt,
	}
	res, err := r.db.Collection(ocrCollectionName).InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

type mlResultDocument struct {
	InspectionID string      `bson:"inspectionId"`
	ImageID      string      `bson:"imageId"`
	JobType      string      `bson:"jobType"`
	ModelVersion string      `bson:"modelVersion"`
	Payload      interface{} `bson:"payload"`
	CreatedAt    time.Time   `bson:"createdAt"`
}

// SaveMLResult persists the raw, unmodified ML service payload for a job type
// (OCR | LABEL_DETECTION | FIELD_EXTRACTION | ANALYZE) and returns its id.
func (r *Repository) SaveMLResult(ctx context.Context, jobType, inspectionID, imageID, modelVersion string, payload interface{}) (string, error) {
	// Round-trip through JSON so the stored document preserves the exact
	// camelCase field names of the raw ML response, rather than bson's
	// default lowercased-struct-field naming.
	var asMap map[string]interface{}
	if raw, err := json.Marshal(payload); err == nil {
		_ = json.Unmarshal(raw, &asMap)
	}

	doc := mlResultDocument{
		InspectionID: inspectionID,
		ImageID:      imageID,
		JobType:      jobType,
		ModelVersion: modelVersion,
		Payload:      asMap,
		CreatedAt:    time.Now().UTC(),
	}
	res, err := r.db.Collection(resultCollectionName).InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}
