package business

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ppwfx/shellpane/internal/domain"
)

func Test_ValidateViewSpecs(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		specs := []domain.ViewSpec{
			{
				Name: "a",
				Command: "a",
				Env: []domain.EnvSpec{
					{
						Name: "a",
					},
				},
			},
			{
				Name: "b",
				Command: "b",
			},
		}

		err := ValidateViewSpecs(specs)
		require.NoError(t, err)
	})

	t.Run("duplicate spec names", func(t *testing.T) {
		specs := []domain.ViewSpec{
			{
				Name: "a",
				Command: "a",
			},
			{
				Name: "a",
				Command: "b",
			},
		}

		err := ValidateViewSpecs(specs)
		require.Error(t, err)
	})

	t.Run("missing spec name", func(t *testing.T) {
		specs := []domain.ViewSpec{
			{
				Name: "",
				Command: "a",
			},
		}

		err := ValidateViewSpecs(specs)
		require.Error(t, err)
	})

	t.Run("missing spec command", func(t *testing.T) {
		specs := []domain.ViewSpec{
			{
				Name: "a",
				Command: "",
			},
		}

		err := ValidateViewSpecs(specs)
		require.Error(t, err)
	})

	t.Run("duplicate env names", func(t *testing.T) {
		specs := []domain.ViewSpec{
			{
				Name: "a",
				Command: "a",
				Env: []domain.EnvSpec{
					{
						Name: "a",
					},
					{
						Name: "a",
					},
				},
			},
		}

		err := ValidateViewSpecs(specs)
		require.Error(t, err)
	})

	t.Run("missing env name", func(t *testing.T) {
		specs := []domain.ViewSpec{
			{
				Name: "a",
				Command: "a",
				Env: []domain.EnvSpec{
					{
						Name: "",
					},
				},
			},
		}

		err := ValidateViewSpecs(specs)
		require.Error(t, err)
	})
}