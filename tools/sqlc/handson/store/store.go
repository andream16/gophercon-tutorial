package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type (
	// FindingsStore implements a findings storage.
	FindingsStore struct {
		db *sql.DB
	}

	// Finding is the DTO for findings.
	Finding struct {
		ID         int64
		InstanceID string
		Details    json.RawMessage
	}
)

// NewFindingsStore returns a new findings store.
func NewFindingsStore() (FindingsStore, error) {
	db, err := sql.Open("sqlite3", "demo.db")
	if err != nil {
		return FindingsStore{}, fmt.Errorf("failed to open database: %w", err)
	}

	return FindingsStore{
		db: db,
	}, nil
}

// ReadFindings returns a slice of findings by instance id.
func (f FindingsStore) ReadFindings(ctx context.Context, instanceID string) ([]Finding, error) {
	return nil, nil
}

// CreateFinding creates a finding.
func (f FindingsStore) CreateFinding(ctx context.Context, finding Finding) error {
	return nil
}
