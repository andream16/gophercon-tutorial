package gopher

import (
	"context"
	"errors"
	"fmt"
)

const (
	BlueColor  Color = "blue"
	PinkColor  Color = "pink"
	GreenColor Color = "green"
)

var (
	ErrGopherNotFound      = errors.New("gopher not found")
	ErrGopherAlreadyExists = errors.New("gopher already exists")
	validColors            = map[Color]struct{}{
		BlueColor:  {},
		PinkColor:  {},
		GreenColor: {},
	}
)

type (
	// Storer abstracts store interactions.
	Storer interface {
		Create(ctx context.Context, gopher Gopher) error
		Get(ctx context.Context, name string) (Gopher, error)
	}
	// Color is a gopher's color.
	Color string
	// Gopher contains gopher's info.
	Gopher struct {
		Name  string
		Color Color
	}
	// Manager allows to manage gophers.
	Manager struct {
		storer Storer
	}
)

// NewManager is the Manager's constructor.
func NewManager(storer Storer) (*Manager, error) {
	if storer == nil {
		return nil, errors.New("nil storer")
	}
	return &Manager{
		storer: storer,
	}, nil
}

// String stringer method.
func (c Color) String() string {
	return string(c)
}

func (g Gopher) validate() error {
	if g.Name == "" {
		return errors.New("invalid empty name")
	}

	if _, ok := validColors[g.Color]; !ok {
		return fmt.Errorf("invalid color: %s", g.Color)
	}

	return nil
}

// Create creates a gopher if it doesn't exist.
func (m *Manager) Create(ctx context.Context, g Gopher) error {
	if err := g.validate(); err != nil {
		return fmt.Errorf("invalid gopher: %w", err)
	}

	if _, err := m.Get(ctx, g.Name); err != nil {
		if errors.Is(err, ErrGopherNotFound) {
			return m.storer.Create(ctx, g)
		}
		return fmt.Errorf("could not create gopher %q: %w", g.Name, err)
	}

	return nil
}

// Get retrieves a gopher if it exists.
func (m *Manager) Get(ctx context.Context, name string) (Gopher, error) {
	return m.storer.Get(ctx, name)
}
