package dashboard

import (
	"context"
	"time"

	"github.com/raunak0400/metrology-backend/internal/db"
)

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service { return &Service{q: q} }

type Summary struct {
	TotalInspections     int64            `json:"totalInspections"`
	TotalProducts        int64            `json:"totalProducts"`
	ActiveRules          int64            `json:"activeRules"`
	OpenViolations       int64            `json:"openViolations"`
	InspectionsByOutcome map[string]int64 `json:"inspectionsByOutcome"`
	InspectionsByStatus  map[string]int64 `json:"inspectionsByStatus"`
}

// GetSummary returns high-level counts for the compliance dashboard.
func (s *Service) GetSummary(ctx context.Context) (Summary, error) {
	total, err := s.q.CountAllInspections(ctx)
	if err != nil {
		return Summary{}, err
	}
	byOutcome, err := s.q.InspectionsGroupByOutcome(ctx)
	if err != nil {
		return Summary{}, err
	}
	byStatus, err := s.q.InspectionsGroupByStatus(ctx)
	if err != nil {
		return Summary{}, err
	}
	openViolations, err := s.q.CountOpenViolations(ctx)
	if err != nil {
		return Summary{}, err
	}
	totalProducts, err := s.q.CountAllProducts(ctx)
	if err != nil {
		return Summary{}, err
	}
	activeRules, err := s.q.CountActiveRules(ctx)
	if err != nil {
		return Summary{}, err
	}

	outcomeMap := make(map[string]int64, len(byOutcome))
	for _, r := range byOutcome {
		outcomeMap[string(r.Outcome)] = r.Total
	}
	statusMap := make(map[string]int64, len(byStatus))
	for _, r := range byStatus {
		statusMap[string(r.Status)] = r.Total
	}

	return Summary{
		TotalInspections: total, TotalProducts: totalProducts, ActiveRules: activeRules,
		OpenViolations: openViolations, InspectionsByOutcome: outcomeMap, InspectionsByStatus: statusMap,
	}, nil
}

type ViolationStats struct {
	BySeverity map[string]int64         `json:"bySeverity"`
	ByStatus   map[string]int64         `json:"byStatus"`
	TopRules   []db.TopViolatedRulesRow `json:"topRules"`
}

// GetViolationStats returns a violation breakdown by severity + status, plus
// the most-triggered rules.
func (s *Service) GetViolationStats(ctx context.Context) (ViolationStats, error) {
	bySeverity, err := s.q.ViolationsGroupBySeverity(ctx)
	if err != nil {
		return ViolationStats{}, err
	}
	byStatus, err := s.q.ViolationsGroupByStatus(ctx)
	if err != nil {
		return ViolationStats{}, err
	}
	topRules, err := s.q.TopViolatedRules(ctx)
	if err != nil {
		return ViolationStats{}, err
	}

	severityMap := make(map[string]int64, len(bySeverity))
	for _, r := range bySeverity {
		severityMap[string(r.Severity)] = r.Total
	}
	statusMap := make(map[string]int64, len(byStatus))
	for _, r := range byStatus {
		statusMap[string(r.Status)] = r.Total
	}

	return ViolationStats{BySeverity: severityMap, ByStatus: statusMap, TopRules: topRules}, nil
}

type InspectionStats struct {
	Recent          []db.RecentInspectionsRow `json:"recent"`
	Last14DaysCount int64                     `json:"last14DaysCount"`
}

// GetInspectionStats returns recent inspections plus a simple 14-day created count.
func (s *Service) GetInspectionStats(ctx context.Context) (InspectionStats, error) {
	recent, err := s.q.RecentInspections(ctx)
	if err != nil {
		return InspectionStats{}, err
	}
	since := time.Now().Add(-14 * 24 * time.Hour)
	count, err := s.q.CountInspectionsSince(ctx, since)
	if err != nil {
		return InspectionStats{}, err
	}
	return InspectionStats{Recent: recent, Last14DaysCount: count}, nil
}
