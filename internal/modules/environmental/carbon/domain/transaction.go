package domain

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/errs"
	"github.com/siddharthg2309/ecosphere-esg-platform/pkg/id"
)

type Source string

const (
	SourcePurchase      Source = "purchase"
	SourceManufacturing Source = "manufacturing"
	SourceExpense       Source = "expense"
	SourceFleet         Source = "fleet"
)

func (s Source) Valid() bool {
	return s == SourcePurchase || s == SourceManufacturing || s == SourceExpense || s == SourceFleet
}

type Status string

const (
	StatusDraft    Status = "draft"
	StatusVerified Status = "verified"
)

type Transaction struct {
	ID               id.ID           `json:"id"`
	DepartmentID     id.ID           `json:"departmentId"`
	Source           Source          `json:"source"`
	Quantity         decimal.Decimal `json:"quantity"`
	EmissionFactorID id.ID           `json:"emissionFactorId"`
	FactorValue      decimal.Decimal `json:"factorValue"`
	ComputedCO2      decimal.Decimal `json:"computedCo2"`
	TxnDate          time.Time       `json:"txnDate"`
	EvidenceURL      string          `json:"evidenceUrl,omitempty"`
	Status           Status          `json:"status"`
	VerifiedBy       *id.ID          `json:"verifiedBy,omitempty"`
	VerifiedAt       *time.Time      `json:"verifiedAt,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
}

func New(departmentID id.ID, source Source, quantity decimal.Decimal, factorID id.ID, factorValue decimal.Decimal, txnDate time.Time, evidenceURL string, now time.Time) (*Transaction, error) {
	t := &Transaction{ID: id.New(), DepartmentID: departmentID, Source: source, Quantity: quantity, EmissionFactorID: factorID, FactorValue: factorValue, ComputedCO2: decimal.Zero, TxnDate: dateOnly(txnDate), EvidenceURL: strings.TrimSpace(evidenceURL), Status: StatusDraft, CreatedAt: now.UTC()}
	return t, t.ValidateDraft()
}

func (t *Transaction) ValidateDraft() error {
	fields := map[string]string{}
	if t.DepartmentID == "" {
		fields["departmentId"] = "Department is required"
	}
	if !t.Source.Valid() {
		fields["source"] = "Source is invalid"
	}
	if !t.Quantity.GreaterThan(decimal.Zero) {
		fields["quantity"] = "Quantity must be greater than zero"
	}
	if t.EmissionFactorID == "" {
		fields["emissionFactorId"] = "Emission factor is required"
	}
	if !t.FactorValue.GreaterThan(decimal.Zero) {
		fields["factorValue"] = "Emission factor must be greater than zero"
	}
	if t.TxnDate.IsZero() {
		fields["txnDate"] = "Transaction date is required"
	}
	if len(fields) > 0 {
		return errs.Invalid("invalid_carbon_transaction", "Carbon transaction details are invalid", fields)
	}
	return nil
}

func (t *Transaction) Verify(by id.ID, now time.Time) error {
	if t.Status == StatusVerified {
		return errs.Conflict("already_verified", "Verified carbon transactions are immutable")
	}
	if by == "" {
		return errs.Invalid("invalid_verifier", "Verifier is required", nil)
	}
	t.ComputedCO2 = t.Quantity.Mul(t.FactorValue).Round(3)
	verifiedAt := now.UTC()
	t.Status, t.VerifiedBy, t.VerifiedAt = StatusVerified, &by, &verifiedAt
	return nil
}

func dateOnly(value time.Time) time.Time {
	y, m, d := value.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
