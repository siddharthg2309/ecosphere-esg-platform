package settings

import (
	"context"
	config "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/esgconfig/domain"
	masterapp "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/app"
	masterport "github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/settings/masterdata/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"testing"
)

type flagsRepo struct {
	masterport.Repository
	value config.Config
	reads int
}

func (r *flagsRepo) GetConfig(context.Context) (config.Config, error) { r.reads++; return r.value, nil }

func TestConfigEventInvalidatesCache(t *testing.T) {
	repo := &flagsRepo{value: config.Config{AutoEmissionCalc: false, WeightEnv: 40, WeightSocial: 30, WeightGov: 30}}
	bus := events.NewInProcess()
	flags := New(repo, bus)
	if flags.IsEnabled(context.Background(), "auto_emission_calc") {
		t.Fatal("expected disabled")
	}
	repo.value.AutoEmissionCalc = true
	if flags.IsEnabled(context.Background(), "auto_emission_calc") {
		t.Fatal("cache should retain old value")
	}
	if err := bus.Publish(context.Background(), masterapp.ESGConfigChanged{}); err != nil {
		t.Fatal(err)
	}
	if !flags.IsEnabled(context.Background(), "auto_emission_calc") {
		t.Fatal("expected refreshed value")
	}
	if repo.reads != 2 {
		t.Fatalf("reads=%d", repo.reads)
	}
}
