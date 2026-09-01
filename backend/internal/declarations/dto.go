package declarations

type UpdateDeclarationRequest struct {
	NormalizedValue string   `json:"normalizedValue" binding:"omitempty,max=1000"`
	NumericValue    *float64 `json:"numericValue"`
	Unit            string   `json:"unit" binding:"omitempty,max=20"`
	Type            string   `json:"type" binding:"omitempty,oneof=MRP NET_QUANTITY MANUFACTURER PACKER IMPORTER MANUFACTURE_DATE PACKING_DATE EXPIRY_DATE CONSUMER_CARE COUNTRY_OF_ORIGIN DIMENSIONS UNIT_SALE_PRICE GENERIC_NAME OTHER"`
}
