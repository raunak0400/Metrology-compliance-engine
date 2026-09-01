package products

import (
	"context"

	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/pagination"
)

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service { return &Service{q: q} }

func (s *Service) Create(ctx context.Context, req CreateProductRequest, createdByID string) (db.GetProductByIDRow, error) {
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return db.GetProductByIDRow{}, apperror.Validation("Invalid categoryId")
	}

	product, err := s.q.CreateProduct(ctx, db.CreateProductParams{
		Name:                req.Name,
		CategoryID:          categoryID,
		Sku:                 dbutil.StrPtr(req.SKU),
		Brand:               dbutil.StrPtr(req.Brand),
		Barcode:             dbutil.StrPtr(req.Barcode),
		ManufacturerID:      dbutil.NullUUID(req.ManufacturerID),
		NetQuantityDeclared: dbutil.StrPtr(req.NetQuantityDeclared),
		MrpDeclared:         db.Float64ToNumeric(req.MRPDeclared),
		Attributes:          dbutil.JSONBytes(req.Attributes),
		CreatedByID:         dbutil.NullUUID(createdByID),
	})
	if err != nil {
		return db.GetProductByIDRow{}, err
	}
	return s.Get(ctx, product.ID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (db.GetProductByIDRow, error) {
	p, err := s.q.GetProductByID(ctx, id)
	if err != nil {
		return db.GetProductByIDRow{}, apperror.NotFound("Product")
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, q ListProductsQuery, page pagination.Query) ([]db.ListProductsRow, int64, error) {
	categoryID := dbutil.NullUUID(q.CategoryID)
	manufacturerID := dbutil.NullUUID(q.ManufacturerID)
	var search *string
	if q.Search != "" {
		search = &q.Search
	}

	items, err := s.q.ListProducts(ctx, db.ListProductsParams{
		Limit: page.LimitInt32(), Offset: page.Offset(),
		CategoryID: categoryID, ManufacturerID: manufacturerID, Search: search,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountProducts(ctx, db.CountProductsParams{
		CategoryID: categoryID, ManufacturerID: manufacturerID, Search: search,
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateProductRequest) (db.GetProductByIDRow, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return db.GetProductByIDRow{}, err
	}

	params := db.UpdateProductParams{
		ID:                  id,
		Sku:                 dbutil.StrPtr(req.SKU),
		Brand:               dbutil.StrPtr(req.Brand),
		Barcode:             dbutil.StrPtr(req.Barcode),
		ManufacturerID:      dbutil.NullUUID(req.ManufacturerID),
		NetQuantityDeclared: dbutil.StrPtr(req.NetQuantityDeclared),
		MrpDeclared:         db.Float64ToNumeric(req.MRPDeclared),
		Attributes:          dbutil.JSONBytes(req.Attributes),
	}
	if req.Name != "" {
		params.Name = &req.Name
	}
	if req.CategoryID != "" {
		params.CategoryID = dbutil.NullUUID(req.CategoryID)
	}

	if _, err := s.q.UpdateProduct(ctx, params); err != nil {
		return db.GetProductByIDRow{}, err
	}
	return s.Get(ctx, id)
}
