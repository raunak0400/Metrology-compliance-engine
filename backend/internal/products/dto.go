package products

type CreateProductRequest struct {
	Name                string                 `json:"name" binding:"required,min=1,max=200"`
	SKU                 string                 `json:"sku" binding:"omitempty,max=80"`
	Brand               string                 `json:"brand" binding:"omitempty,max=120"`
	Barcode             string                 `json:"barcode" binding:"omitempty,max=64"`
	CategoryID          string                 `json:"categoryId" binding:"required,uuid4"`
	ManufacturerID      string                 `json:"manufacturerId" binding:"omitempty,uuid4"`
	NetQuantityDeclared string                 `json:"netQuantityDeclared" binding:"omitempty,max=80"`
	MRPDeclared         *float64               `json:"mrpDeclared" binding:"omitempty,gte=0"`
	Attributes          map[string]interface{} `json:"attributes"`
}

type UpdateProductRequest struct {
	Name                string                 `json:"name" binding:"omitempty,min=1,max=200"`
	SKU                 string                 `json:"sku" binding:"omitempty,max=80"`
	Brand               string                 `json:"brand" binding:"omitempty,max=120"`
	Barcode             string                 `json:"barcode" binding:"omitempty,max=64"`
	CategoryID          string                 `json:"categoryId" binding:"omitempty,uuid4"`
	ManufacturerID      string                 `json:"manufacturerId" binding:"omitempty,uuid4"`
	NetQuantityDeclared string                 `json:"netQuantityDeclared" binding:"omitempty,max=80"`
	MRPDeclared         *float64               `json:"mrpDeclared" binding:"omitempty,gte=0"`
	Attributes          map[string]interface{} `json:"attributes"`
}

type ListProductsQuery struct {
	Page           int    `form:"page"`
	Limit          int    `form:"limit"`
	CategoryID     string `form:"categoryId" binding:"omitempty,uuid4"`
	ManufacturerID string `form:"manufacturerId" binding:"omitempty,uuid4"`
	Search         string `form:"search" binding:"omitempty,max=120"`
}
