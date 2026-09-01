package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/pagination"
)

type Service struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewService(q *db.Queries, pool *pgxpool.Pool) *Service { return &Service{q: q, pool: pool} }

// View is the API-facing user shape (never exposes the password hash).
type View struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	FullName    string     `json:"fullName"`
	Phone       *string    `json:"phone"`
	IsActive    bool       `json:"isActive"`
	LastLoginAt *string    `json:"lastLoginAt,omitempty"`
	Roles       []RoleView `json:"roles"`
}

type RoleView struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (s *Service) toView(ctx context.Context, u db.User) (View, error) {
	roles, err := s.q.ListRolesForUser(ctx, u.ID)
	if err != nil {
		return View{}, err
	}
	roleViews := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		roleViews = append(roleViews, RoleView{ID: r.ID, Name: r.Name})
	}
	var lastLogin *string
	if u.LastLoginAt != nil {
		s := u.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		lastLogin = &s
	}
	return View{
		ID: u.ID, Email: u.Email, FullName: u.FullName, Phone: u.Phone,
		IsActive: u.IsActive, LastLoginAt: lastLogin, Roles: roleViews,
	}, nil
}

func (s *Service) List(ctx context.Context, q ListUsersQuery, page pagination.Query) ([]View, int64, error) {
	params := db.ListUsersParams{Limit: page.LimitInt32(), Offset: page.Offset()}
	countParams := db.CountUsersParams{}
	if q.IsActive != "" {
		v := q.IsActive == "true"
		params.IsActive = &v
		countParams.IsActive = &v
	}
	if q.Search != "" {
		params.Search = &q.Search
		countParams.Search = &q.Search
	}

	rows, err := s.q.ListUsers(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	views := make([]View, 0, len(rows))
	for _, u := range rows {
		v, err := s.toView(ctx, u)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, v)
	}
	total, err := s.q.CountUsers(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}
	return views, total, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (View, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return View{}, apperror.NotFound("User")
	}
	return s.toView(ctx, u)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (View, error) {
	if _, err := s.q.GetUserByID(ctx, id); err != nil {
		return View{}, apperror.NotFound("User")
	}

	if len(req.RoleIDs) > 0 {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return View{}, err
		}
		defer tx.Rollback(ctx)
		qtx := s.q.WithTx(tx)

		if err := qtx.DeleteUserRoles(ctx, id); err != nil {
			return View{}, err
		}
		for _, rid := range req.RoleIDs {
			roleID, err := uuid.Parse(rid)
			if err != nil {
				continue
			}
			if err := qtx.InsertUserRole(ctx, db.InsertUserRoleParams{UserID: id, RoleID: roleID}); err != nil {
				return View{}, err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return View{}, err
		}
	}

	params := db.UpdateUserParams{ID: id}
	if req.FullName != "" {
		params.FullName = &req.FullName
	}
	if req.Phone != "" {
		params.Phone = &req.Phone
	}
	if req.IsActive != nil {
		params.IsActive = req.IsActive
	}

	u, err := s.q.UpdateUser(ctx, params)
	if err != nil {
		return View{}, err
	}
	return s.toView(ctx, u)
}
