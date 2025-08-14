package gopher_test

import (
	"context"
	"testing"
	"time"

	"github.com/andream16/gophercon-tutorial/tools/mockgen/complete/gopher"
)

type mockStore struct{}

func (m mockStore) Create(ctx context.Context, gopher gopher.Gopher) error {
	return nil
}

func (m mockStore) Get(ctx context.Context, name string) (gopher.Gopher, error) {
	return gopher.Gopher{}, nil
}

func TestManager(t *testing.T) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		g           = gopher.Gopher{
			Name:  "Ella",
			Color: gopher.BlueColor,
		}
		mockStorer = mockStore{}
	)
	defer cancel()
	_ = g

	manager, err := gopher.NewManager(mockStorer)
	if err != nil {
		t.Fatalf("could not create gopher manager: %v", err)
	}

	t.Run("it should return an error because the gopher is not valid", func(t *testing.T) {
		t.Run("when the name is invalid", func(t *testing.T) {
			if err := manager.Create(
				ctx,
				gopher.Gopher{
					Color: gopher.BlueColor,
				},
			); err == nil {
				t.Fatal("expected an error because the name is invalid, got none")
			}
		})
		t.Run("when the color is invalid", func(t *testing.T) {
			if err := manager.Create(
				ctx,
				gopher.Gopher{
					Name:  "steve",
					Color: "yellow",
				},
			); err == nil {
				t.Fatal("expected an error because the color is invalid, got none")
			}
		})
	})

	t.Run("it should return an error when something unexpected happens when getting a gopher", func(t *testing.T) {
		// TODO: implement
	})

	t.Run("it should return an error when something unexpected happens when creating a gopher", func(t *testing.T) {
		// TODO: implement
	})

	t.Run("it should successfully create a gopher", func(t *testing.T) {
		// // TODO: implement
	})
}
