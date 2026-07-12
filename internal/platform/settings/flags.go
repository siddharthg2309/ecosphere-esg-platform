package settings

import (
	"context"
	config "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/esgconfig/domain"
	masterport "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"sync"
	"time"
)

type Flags interface {
	IsEnabled(context.Context, string) bool
	Weights(context.Context) (int, int, int)
}
type Service struct {
	repo    masterport.Repository
	mu      sync.RWMutex
	cached  config.Config
	expires time.Time
	ttl     time.Duration
}

func New(repo masterport.Repository, bus events.Bus) *Service {
	s := &Service{repo: repo, ttl: 60 * time.Second}
	bus.Subscribe("ESGConfigChanged", func(context.Context, events.Event) error { s.Invalidate(); return nil })
	return s
}
func (s *Service) load(ctx context.Context) config.Config {
	s.mu.RLock()
	if time.Now().Before(s.expires) {
		v := s.cached
		s.mu.RUnlock()
		return v
	}
	s.mu.RUnlock()
	v, err := s.repo.GetConfig(ctx)
	if err != nil {
		return config.Config{WeightEnv: 40, WeightSocial: 30, WeightGov: 30}
	}
	s.mu.Lock()
	s.cached = v
	s.expires = time.Now().Add(s.ttl)
	s.mu.Unlock()
	return v
}
func (s *Service) IsEnabled(ctx context.Context, key string) bool {
	v := s.load(ctx)
	switch key {
	case "auto_emission_calc":
		return v.AutoEmissionCalc
	case "require_csr_evidence":
		return v.RequireCSREvidence
	case "auto_award_badges":
		return v.AutoAwardBadges
	case "notify_compliance_email":
		return v.NotifyComplianceEmail
	default:
		return false
	}
}
func (s *Service) Weights(ctx context.Context) (int, int, int) {
	v := s.load(ctx)
	return v.WeightEnv, v.WeightSocial, v.WeightGov
}
func (s *Service) Invalidate() { s.mu.Lock(); s.expires = time.Time{}; s.mu.Unlock() }
