package domain

import "testing"

func TestWeightsMustTotalOneHundred(t *testing.T) {
	if err := (Config{WeightEnv: 40, WeightSocial: 30, WeightGov: 40}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
	if err := (Config{WeightEnv: 40, WeightSocial: 30, WeightGov: 30}).Validate(); err != nil {
		t.Fatal(err)
	}
}
