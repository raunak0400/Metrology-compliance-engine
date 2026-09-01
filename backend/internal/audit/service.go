// Package audit provides audit logging. Important mutations across every
// module call Record to persist who did what, when, from where, and the
// before/after state. Failures to write an audit record are logged but never
// break the primary request.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/middleware"
	"github.com/raunak0400/metrology-backend/internal/pagination"
)

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service { return &Service{q: q} }

type RecordParams struct {
	Action     db.AuditAction
	EntityType string
	EntityID   string
	OldState   interface{}
	NewState   interface{}
}

func marshalOrNil(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// Record persists one audit log entry, deriving actor/IP/user-agent/request-id
// from the current Gin request context.
func (s *Service) Record(c *gin.Context, p RecordParams) {
	var actorID uuid.NullUUID
	if u, ok := middleware.CurrentUser(c); ok {
		if id, err := uuid.Parse(u.ID); err == nil {
			actorID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var entityID *string
	if p.EntityID != "" {
		entityID = &p.EntityID
	}
	requestID := c.GetString("requestId")
	var requestIDPtr *string
	if requestID != "" {
		requestIDPtr = &requestID
	}
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err := s.q.CreateAuditLog(context.Background(), db.CreateAuditLogParams{
		Action:     p.Action,
		EntityType: p.EntityType,
		ActorID:    actorID,
		EntityID:   entityID,
		OldState:   marshalOrNil(p.OldState),
		NewState:   marshalOrNil(p.NewState),
		Ip:         &ip,
		UserAgent:  &userAgent,
		RequestID:  requestIDPtr,
	})
	if err != nil {
		slog.Error("failed to write audit log", "err", err, "action", p.Action, "entityType", p.EntityType)
	}
}

type ListQuery struct {
	EntityType string
	EntityID   string
	ActorID    string
}

func (s *Service) List(ctx context.Context, q ListQuery, page pagination.Query) ([]db.AuditLog, int64, error) {
	params := db.ListAuditLogsParams{Limit: page.LimitInt32(), Offset: page.Offset()}
	countParams := db.CountAuditLogsParams{}
	if q.EntityType != "" {
		params.EntityType = &q.EntityType
		countParams.EntityType = &q.EntityType
	}
	if q.EntityID != "" {
		params.EntityID = &q.EntityID
		countParams.EntityID = &q.EntityID
	}
	if q.ActorID != "" {
		if id, err := uuid.Parse(q.ActorID); err == nil {
			params.ActorID = uuid.NullUUID{UUID: id, Valid: true}
			countParams.ActorID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	items, err := s.q.ListAuditLogs(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountAuditLogs(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
