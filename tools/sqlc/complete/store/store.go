package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	sqlc "github.com/andream16/gophercon-tutorial/tools/sqlc/complete/store/sqlc/gen"
)

type (
	// FindingsStore implements a findings storage.
	FindingsStore struct {
		db      *sql.DB
		queries *sqlc.Queries
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
		db:      db,
		queries: sqlc.New(db),
	}, nil
}

// ReadFindings returns a slice of findings by instance id.
func (f FindingsStore) ReadFindings(ctx context.Context, instanceID string) ([]Finding, error) {
	rows, err := f.queries.FindingsByID(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("could not find findings by ID %s: %w", instanceID, err)
	}

	var findings = make([]Finding, 0, len(rows))
	for _, row := range rows {
		findings = append(
			findings,
			Finding{
				ID:         row.ID,
				InstanceID: instanceID,
				Details:    json.RawMessage(row.Details),
			},
		)
	}

	return findings, nil
}

// CreateFinding creates a finding.
func (f FindingsStore) CreateFinding(ctx context.Context, finding Finding) error {
	bb, err := finding.Details.MarshalJSON()
	if err != nil {
		return fmt.Errorf("could not marshal finding details: %w", err)
	}

	if err := f.queries.CreateFinding(ctx, sqlc.CreateFindingParams{
		InstanceID: finding.InstanceID,
		Details:    string(bb),
	}); err != nil {
		return fmt.Errorf("could not create finding: %w", err)
	}

	return nil
}
