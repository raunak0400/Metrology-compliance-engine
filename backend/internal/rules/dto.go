package rules

type RequirementInput struct {
	Key             string                 `json:"key" binding:"required,min=1,max=120"`
	DeclarationType string                 `json:"declarationType" binding:"omitempty,oneof=MRP NET_QUANTITY MANUFACTURER PACKER IMPORTER MANUFACTURE_DATE PACKING_DATE EXPIRY_DATE CONSUMER_CARE COUNTRY_OF_ORIGIN DIMENSIONS UNIT_SALE_PRICE GENERIC_NAME OTHER"`
	Operator        string                 `json:"operator" binding:"required,oneof=EXISTS NOT_EMPTY REGEX_MATCH NUMERIC_MIN NUMERIC_MAX NUMERIC_RANGE MIN_TEXT_HEIGHT_MM ENUM_IN DATE_VALID DATE_NOT_FUTURE REQUIRED_IF_IMPORTED CUSTOM"`
	Expected        map[string]interface{} `json:"expected"`
	Unit            string                 `json:"unit" binding:"omitempty,max=20"`
	Severity        string                 `json:"severity" binding:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	Message         string                 `json:"message" binding:"omitempty,max=500"`
	IsMandatory     *bool                  `json:"isMandatory"`
	OrderIndex      *int                   `json:"orderIndex" binding:"omitempty,gte=0"`
}

type CreateRuleRequest struct {
	Code            string                 `json:"code" binding:"required,min=2,max=80"`
	Title           string                 `json:"title" binding:"required,min=2,max=200"`
	Description     string                 `json:"description" binding:"omitempty,max=2000"`
	ActReference    string                 `json:"actReference" binding:"omitempty,max=300"`
	Section         string                 `json:"section" binding:"omitempty,max=80"`
	SeverityDefault string                 `json:"severityDefault" binding:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	EffectiveFrom   string                 `json:"effectiveFrom" binding:"omitempty"`
	EffectiveTo     string                 `json:"effectiveTo" binding:"omitempty"`
	IsActive        *bool                  `json:"isActive"`
	Params          map[string]interface{} `json:"params"`
	CategoryIDs     []string               `json:"categoryIds" binding:"omitempty,dive,uuid4"`
	Requirements    []RequirementInput     `json:"requirements"`
}

type UpdateRuleRequest struct {
	Title           string                 `json:"title" binding:"omitempty,min=2,max=200"`
	Description     string                 `json:"description" binding:"omitempty,max=2000"`
	ActReference    string                 `json:"actReference" binding:"omitempty,max=300"`
	Section         string                 `json:"section" binding:"omitempty,max=80"`
	SeverityDefault string                 `json:"severityDefault" binding:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	EffectiveFrom   string                 `json:"effectiveFrom" binding:"omitempty"`
	EffectiveTo     string                 `json:"effectiveTo" binding:"omitempty"`
	IsActive        *bool                  `json:"isActive"`
	Params          map[string]interface{} `json:"params"`
	CategoryIDs     []string               `json:"categoryIds" binding:"omitempty,dive,uuid4"`
	Requirements    []RequirementInput     `json:"requirements"`
}

type ListRulesQuery struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	IsActive string `form:"isActive" binding:"omitempty,oneof=true false"`
	Search   string `form:"search" binding:"omitempty,max=120"`
}
