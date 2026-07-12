package app

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/domain"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/modules/environmental/carbon/port"
	"github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/events"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/page"
)

type RecordCommand struct {
	DepartmentID     id.ID
	Source           domain.Source
	Quantity         decimal.Decimal
	EmissionFactorID id.ID
	Unit             string
	TxnDate          time.Time
	EvidenceURL      string
}

type Service struct {
	repo  port.Repository
	flags port.Flags
	bus   events.Bus
	now   func() time.Time
}

func New(repo port.Repository, flags port.Flags, bus events.Bus) *Service {
	return &Service{repo: repo, flags: flags, bus: bus, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Record(ctx context.Context, cmd RecordCommand) (*domain.Transaction, error) {
	exists, err := s.repo.DepartmentExists(ctx, cmd.DepartmentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.Invalid("invalid_department", "Department does not exist", map[string]string{"departmentId": "Department not found"})
	}
	unit, factorValue, active, err := s.repo.Factor(ctx, cmd.EmissionFactorID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, errs.Invalid("inactive_emission_factor", "Emission factor is inactive", map[string]string{"emissionFactorId": "Select an active factor"})
	}
	if unit != cmd.Unit {
		return nil, errs.Invalid("factor_unit_mismatch", "Unit does not match the emission factor", map[string]string{"unit": "Expected " + unit})
	}
	txn, err := domain.New(cmd.DepartmentID, cmd.Source, cmd.Quantity, cmd.EmissionFactorID, factorValue, cmd.TxnDate, cmd.EvidenceURL, s.now())
	if err != nil {
		return nil, err
	}
	if err = s.repo.Create(ctx, txn); err != nil {
		return nil, err
	}
	return txn, nil
}

func (s *Service) Verify(ctx context.Context, transactionID, by id.ID) (*domain.Transaction, error) {
	txn, err := s.repo.ByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	allowed, err := s.repo.IsDepartmentHead(ctx, by, txn.DepartmentID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errs.Forbidden("not_dept_head", "Only the assigned department head can verify this transaction")
	}
	if err = txn.Verify(by, s.now()); err != nil {
		return nil, err
	}
	if err = s.repo.SaveVerified(ctx, txn); err != nil {
		return nil, err
	}
	if err = s.bus.Publish(ctx, events.EmissionRecorded{DepartmentID: txn.DepartmentID, Source: string(txn.Source), CO2: txn.ComputedCO2, At: txn.TxnDate}); err != nil {
		return txn, err
	}
	return txn, nil
}

func (s *Service) List(ctx context.Context, filter port.Filter) (page.Result[domain.Transaction], error) {
	return s.repo.List(ctx, filter)
}
func (s *Service) Summary(ctx context.Context, departmentID *id.ID, from, to time.Time) (port.Summary, error) {
	return s.repo.Summary(ctx, departmentID, from, to)
}
func (s *Service) AutoCalcEnabled(ctx context.Context) bool {
	return s.flags.IsEnabled(ctx, "auto_emission_calc")
}
