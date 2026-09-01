// Package constants holds shared role names, permission scopes and pagination
// defaults used across modules and middleware.
package constants

const (
	RoleAdmin      = "ADMIN"
	RoleInspector  = "INSPECTOR"
	RoleReviewer   = "REVIEWER"
	RoleSupervisor = "SUPERVISOR"
)

const (
	PermInspectionsRead     = "inspections:read"
	PermInspectionsWrite    = "inspections:write"
	PermInspectionsFinalize = "inspections:finalize"
	PermProductsRead        = "products:read"
	PermProductsWrite       = "products:write"
	PermScansWrite          = "scans:write"
	PermDeclarationsWrite   = "declarations:write"
	PermViolationsWrite     = "violations:write"
	PermRulesRead           = "rules:read"
	PermRulesWrite          = "rules:write"
	PermReportsRead         = "reports:read"
	PermDatasetsWrite       = "datasets:write"
	PermAuditRead           = "audit:read"
	PermDashboardRead       = "dashboard:read"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// ManualReviewConfidenceThreshold: below this confidence, an extracted
// declaration / check is routed to MANUAL_REVIEW rather than becoming an
// automatic legal violation.
const ManualReviewConfidenceThreshold = 0.7
