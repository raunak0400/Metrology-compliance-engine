// Database seed — development bootstrap data.
//
// Seeds:
//   - system roles (ADMIN, INSPECTOR, REVIEWER, SUPERVISOR)
//   - example product categories
//   - example legal rules + rule requirements (TEMPLATES — see note below)
//   - a development admin user
//
// IMPORTANT: The legal rules seeded here are illustrative EXAMPLES for local
// development only. They are NOT an authoritative statement of the law. The
// legal-rule repository must be populated/verified from the authoritative
// Legal Metrology Act, 2009 and Legal Metrology (Packaged Commodities)
// Rules, 2011 (and amendments) before any real compliance use.
//
// Idempotent: safe to run multiple times (uses upserts on natural keys).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/raunak0400/metrology-backend/internal/auth"
	"github.com/raunak0400/metrology-backend/internal/config"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/pkg/postgres"
)

type roleSeed struct {
	name        string
	description string
	permissions []string
}

var systemRoles = []roleSeed{
	{name: "ADMIN", description: "Full system administrator", permissions: []string{"*"}},
	{
		name:        "INSPECTOR",
		description: "Creates inspections, uploads images, triggers processing",
		permissions: []string{"inspections:read", "inspections:write", "products:read", "products:write", "scans:write", "reports:read"},
	},
	{
		name:        "REVIEWER",
		description: "Reviews low-confidence / flagged inspections and edits declarations",
		permissions: []string{"inspections:read", "declarations:write", "violations:write", "inspections:finalize"},
	},
	{
		name:        "SUPERVISOR",
		description: "Oversight, dashboard, rules and audit access",
		permissions: []string{"inspections:read", "rules:read", "rules:write", "audit:read", "dashboard:read"},
	},
}

type categorySeed struct{ name, code, description string }

var categories = []categorySeed{
	{"General Packaged Commodity", "GENERAL", "Default catch-all category"},
	{"Packaged Food", "FOOD", "Pre-packaged food commodities"},
	{"Cosmetics & Toiletries", "COSMETICS", "Cosmetic packaged goods"},
	{"Household & Cleaning", "HOUSEHOLD", "Cleaning agents, detergents"},
	{"Consumer Electronics", "ELECTRONICS", "Packaged electronic goods"},
}

type requirementSeed struct {
	key             string
	declarationType string
	operator        string
	expected        map[string]interface{}
	unit            string
	message         string
}

type ruleSeed struct {
	code            string
	title           string
	description     string
	actReference    string
	severityDefault string
	requirements    []requirementSeed
}

// Example rule templates. `requirements` describe how the deterministic
// engine evaluates the rule. Replace/verify against authoritative legal text.
var ruleTemplates = []ruleSeed{
	{
		code: "PCR2011-MRP-PRESENCE", title: "Retail sale price (MRP) declaration present",
		description:     "Every package must declare the retail sale price (MRP inclusive of all taxes).",
		actReference:    "LM (Packaged Commodities) Rules, 2011 — Rule 6 / 18 (EXAMPLE)",
		severityDefault: "HIGH",
		requirements: []requirementSeed{
			{key: "declaration.MRP.present", declarationType: "MRP", operator: "EXISTS", message: "MRP declaration is missing on the package."},
			{
				key: "declaration.MRP.format", declarationType: "MRP", operator: "REGEX_MATCH",
				expected: map[string]interface{}{"pattern": `(?i)(m\.?r\.?p\.?|maximum retail price).*?\d`},
				message:  `MRP is present but not in an expected "MRP Rs. <amount>" format.`,
			},
		},
	},
	{
		code: "PCR2011-NETQTY-PRESENCE", title: "Net quantity declaration present",
		description:  "Every package must declare the net quantity in standard units.",
		actReference: "LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)", severityDefault: "HIGH",
		requirements: []requirementSeed{
			{key: "declaration.NET_QUANTITY.present", declarationType: "NET_QUANTITY", operator: "EXISTS", message: "Net quantity declaration is missing."},
		},
	},
	{
		code: "PCR2011-MANUFACTURER-DETAILS", title: "Manufacturer / packer / importer name and address",
		description:  "The name and complete address of the manufacturer, packer or importer must be declared.",
		actReference: "LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)", severityDefault: "MEDIUM",
		requirements: []requirementSeed{
			{key: "declaration.MANUFACTURER.present", declarationType: "MANUFACTURER", operator: "NOT_EMPTY", message: "Manufacturer/packer/importer name and address not found."},
		},
	},
	{
		code: "PCR2011-CONSUMER-CARE", title: "Consumer care details present",
		description:  "Consumer care contact details (name, address, phone/email) must be declared.",
		actReference: "LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)", severityDefault: "MEDIUM",
		requirements: []requirementSeed{
			{key: "declaration.CONSUMER_CARE.present", declarationType: "CONSUMER_CARE", operator: "NOT_EMPTY", message: "Consumer care details not found."},
		},
	},
	{
		code: "PCR2011-COUNTRY-OF-ORIGIN", title: "Country of origin (imported packages)",
		description:  "Imported packages must declare the country of origin.",
		actReference: "LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)", severityDefault: "MEDIUM",
		requirements: []requirementSeed{
			{key: "declaration.COUNTRY_OF_ORIGIN.requiredIfImported", declarationType: "COUNTRY_OF_ORIGIN", operator: "REQUIRED_IF_IMPORTED", message: "Country of origin not declared for an imported package."},
		},
	},
	{
		code: "PCR2011-MFG-DATE", title: "Date of manufacture / packing / import",
		description:  "The month and year of manufacture, pre-packing or import must be declared.",
		actReference: "LM (Packaged Commodities) Rules, 2011 — Rule 6 (EXAMPLE)", severityDefault: "MEDIUM",
		requirements: []requirementSeed{
			{key: "declaration.MANUFACTURE_DATE.valid", declarationType: "MANUFACTURE_DATE", operator: "DATE_VALID", message: "Date of manufacture/packing missing or not a valid date."},
		},
	},
	{
		code: "PCR2011-MIN-TEXT-HEIGHT", title: "Minimum height of declarations (measurement)",
		description:  "Mandatory declarations must meet the minimum letter/numeral height based on the principal display area. Requires a valid scale reference; otherwise flagged for manual review.",
		actReference: "LM (Packaged Commodities) Rules, 2011 — Rule 8 (EXAMPLE)", severityDefault: "LOW",
		requirements: []requirementSeed{
			{
				key: "measurement.declaration.minHeight", declarationType: "NET_QUANTITY", operator: "MIN_TEXT_HEIGHT_MM",
				expected: map[string]interface{}{"minMm": 1.0}, unit: "mm",
				message: "Declaration text height appears below the minimum required height.",
			},
		},
	},
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect postgres:", err)
		os.Exit(1)
	}
	defer pool.Close()

	q := db.New(pool)

	fmt.Println("Seeding database...")
	seedRoles(ctx, q)
	seedCategories(ctx, q)
	if os.Getenv("SEED_EXAMPLE_RULES") == "true" {
		seedRules(ctx, q)
	} else {
		fmt.Println("Skipping example legal rule templates (set SEED_EXAMPLE_RULES=true to seed illustrative examples for local dev; " +
			"the legal-rule repository must otherwise be populated from authoritative Legal Metrology (Packaged Commodities) Rules, 2011 text).")
	}
	seedAdmin(ctx, q)
	fmt.Println("Seeding complete.")
}

func seedRoles(ctx context.Context, q *db.Queries) {
	for _, r := range systemRoles {
		if existing, err := q.GetRoleByName(ctx, r.name); err == nil {
			_, _ = q.UpdateRole(ctx, db.UpdateRoleParams{ID: existing.ID, Description: &r.description, Permissions: r.permissions})
			continue
		}
		if _, err := q.CreateRole(ctx, db.CreateRoleParams{Name: r.name, Description: &r.description, Permissions: r.permissions}); err != nil {
			fmt.Fprintln(os.Stderr, "seed role", r.name, err)
		}
	}
	fmt.Printf("Seeded %d roles.\n", len(systemRoles))
}

func seedCategories(ctx context.Context, q *db.Queries) {
	for _, c := range categories {
		if _, err := q.UpsertCategory(ctx, db.UpsertCategoryParams{Name: c.name, Code: c.code, Description: &c.description}); err != nil {
			fmt.Fprintln(os.Stderr, "seed category", c.code, err)
		}
	}
	fmt.Printf("Seeded %d product categories.\n", len(categories))
}

func seedRules(ctx context.Context, q *db.Queries) {
	general, err := q.GetCategoryByCode(ctx, "GENERAL")
	hasGeneral := err == nil

	for _, t := range ruleTemplates {
		severity := db.ViolationSeverity(t.severityDefault)
		var rule db.LegalRule
		if existing, err := findRuleByCode(ctx, q, t.code); err == nil {
			rule, err = q.UpdateLegalRule(ctx, db.UpdateLegalRuleParams{
				ID: existing.ID, Title: &t.title, Description: &t.description, ActReference: &t.actReference,
				SeverityDefault: &severity, IsActive: boolPtr(true),
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, "update rule", t.code, err)
				continue
			}
		} else {
			rule, err = q.CreateLegalRule(ctx, db.CreateLegalRuleParams{
				Code: t.code, Title: t.title, Description: &t.description, ActReference: &t.actReference,
				SeverityDefault: &severity, IsActive: boolPtr(true),
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, "create rule", t.code, err)
				continue
			}
		}

		// Replace requirements deterministically for idempotency.
		_ = q.DeleteRuleRequirements(ctx, rule.ID)
		for i, r := range t.requirements {
			declType := db.DeclarationType(r.declarationType)
			_, err := q.CreateRuleRequirement(ctx, db.CreateRuleRequirementParams{
				RuleID: rule.ID, Key: r.key, DeclarationType: &declType, Operator: db.RequirementOperator(r.operator),
				Expected: dbutil.JSONBytes(r.expected), Unit: dbutil.StrPtr(r.unit), Message: &r.message,
				IsMandatory: true, OrderIndex: int32(i),
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, "create requirement", r.key, err)
			}
		}

		// Map every example rule to the GENERAL category.
		if hasGeneral {
			_ = q.CreateRuleCategory(ctx, db.CreateRuleCategoryParams{RuleID: rule.ID, CategoryID: general.ID})
		}
	}
	fmt.Printf("Seeded %d example legal rules with requirements.\n", len(ruleTemplates))
}

func findRuleByCode(ctx context.Context, q *db.Queries, code string) (db.LegalRule, error) {
	rows, err := q.ListLegalRules(ctx, db.ListLegalRulesParams{Limit: 1, Offset: 0, Search: &code})
	if err != nil {
		return db.LegalRule{}, err
	}
	for _, r := range rows {
		if r.Code == code {
			return q.GetLegalRule(ctx, r.ID)
		}
	}
	return db.LegalRule{}, fmt.Errorf("rule %s not found", code)
}

func seedAdmin(ctx context.Context, q *db.Queries) {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@legalmetrology.local"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "ChangeMe123!"
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hash admin password:", err)
		return
	}

	admin, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		admin, err = q.CreateUser(ctx, db.CreateUserParams{Email: email, PasswordHash: passwordHash, FullName: "Development Admin"})
		if err != nil {
			fmt.Fprintln(os.Stderr, "create admin:", err)
			return
		}
	}

	if adminRole, err := q.GetRoleByName(ctx, "ADMIN"); err == nil {
		_ = q.InsertUserRole(ctx, db.InsertUserRoleParams{UserID: admin.ID, RoleID: adminRole.ID})
	}

	fmt.Printf("Seeded admin user: %s (password: %s)\n", email, password)
}

func boolPtr(b bool) *bool { return &b }
