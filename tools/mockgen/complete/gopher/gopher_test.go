package gopher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/andream16/gophercon-tutorial/tools/mockgen/complete/gopher"
)

func TestManager(t *testing.T) {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		ctrl        = gomock.NewController(t)
		g           = gopher.Gopher{
			Name:  "Ella",
			Color: gopher.BlueColor,
		}
		mockStorer = NewMockStorer(ctrl)
	)
	defer cancel()

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
		mockStorer.
			EXPECT().
			Get(ctx, g.Name).
			Return(gopher.Gopher{}, errors.New("something bad happened"))

		if err := manager.Create(
			ctx,
			g,
		); err == nil {
			t.Fatal("expected an error on get, got none")
		}
	})

	t.Run("it should return an error when something unexpected happens when creating a gopher", func(t *testing.T) {
		gomock.InOrder(
			mockStorer.
				EXPECT().
				Get(ctx, g.Name).
				Return(gopher.Gopher{}, gopher.ErrGopherNotFound),
			mockStorer.
				EXPECT().
				Create(ctx, g).
				Return(errors.New("something bad happened")),
		)

		if err := manager.Create(
			ctx,
			g,
		); err == nil {
			t.Fatal("expected an error on create, got none")
		}
	})

	t.Run("it should successfully create a gopher", func(t *testing.T) {
		gomock.InOrder(
			mockStorer.
				EXPECT().
				Get(ctx, g.Name).
				Return(gopher.Gopher{}, gopher.ErrGopherNotFound),
			mockStorer.
				EXPECT().
				Create(ctx, g).
				Return(nil),
		)

		if err := manager.Create(
			ctx,
			g,
		); err != nil {
			t.Fatalf("expected no error on create, got %v", err)
		}
	})
}
