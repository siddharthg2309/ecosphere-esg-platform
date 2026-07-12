package id

import (
	"fmt"

	"github.com/google/uuid"
)

type ID string

func New() ID {
	v, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Errorf("create uuid v7: %w", err))
	}
	return ID(v.String())
}

func Parse(value string) (ID, error) {
	if _, err := uuid.Parse(value); err != nil {
		return "", fmt.Errorf("invalid id: %w", err)
	}
	return ID(value), nil
}

func (v ID) String() string { return string(v) }
