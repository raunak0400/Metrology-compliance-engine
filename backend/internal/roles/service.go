package roles

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/db"
)

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service { return &Service{q: q} }

func (s *Service) List(ctx context.Context) ([]db.Role, error) {
	return s.q.ListRoles(ctx)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (db.Role, error) {
	role, err := s.q.GetRoleByID(ctx, id)
	if err != nil {
		return db.Role{}, apperror.NotFound("Role")
	}
	return role, nil
}

func (s *Service) Create(ctx context.Context, req CreateRoleRequest) (db.Role, error) {
	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}
	return s.q.CreateRole(ctx, db.CreateRoleParams{
		Name:        strings.ToUpper(req.Name),
		Description: desc,
		Permissions: perms,
	})
}
