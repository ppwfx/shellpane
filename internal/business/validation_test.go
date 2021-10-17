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


func Test_ValidateGetViewOutputRequest(t *testing.T) {
	t.Run("validate", func(t *testing.T) {
		viewSpecs := []domain.ViewSpec{
			{
				Name:    "Hello",
				Command: "echo hello $HELLO",
				Env: []domain.EnvSpec{
					{
						Name: "HELLO",
						Validator: `^[A-Za-z0-9_./-]{1,15}$`,
					},
				},
			},
		}

		getViewOutputRequest := GetViewOutputRequest{
			Name: "Hello",
			Format: FormatRaw,
			Env: []EnvValue{
				{
					Name: "HELLO",
					Value: "Herbert",
				},
			},
		}

		for i := range viewSpecs {
			err := validateGetViewOutputRequest(viewSpecs[i], getViewOutputRequest)
			require.NoError(t, err)
		}
	})

	t.Run("multiple env specs", func(t *testing.T) {
		viewSpecs := []domain.ViewSpec{
			{
				Name:    "Hello",
				Command: "echo hello $HELLO $BYE",
				Env: []domain.EnvSpec{
					{
						Name: "HELLO",
						Validator: `^[A-Za-z0-9_./-]{1,15}$`,
					},
					{
						Name: "SIR",
						Validator: `^[A-Za-z0-9_./-]{1,15}$`,
					},
				},
			},
		}

		getViewOutputRequest := GetViewOutputRequest{
			Name: "Hello",
			Format: FormatRaw,
			Env: []EnvValue{
				{
					Name: "HELLO",
					Value: "Hello",
				},
				{
					Name: "SIR",
					Value: "Sir.",
				},
			},
		}

		for i := range viewSpecs {
			err := validateGetViewOutputRequest(viewSpecs[i], getViewOutputRequest)
			require.NoError(t, err)
		}
	})

	t.Run("to short env spec value", func(t *testing.T) {
		viewSpecs := []domain.ViewSpec{
			{
				Name:    "Hello",
				Command: "echo hello $B",
				Env: []domain.EnvSpec{
					{
						Name:      "B",
						Validator: `^[A-Za-z0-9_./-]{2,15}$`,
					},
				},
			},
		}

		getViewOutputRequest := GetViewOutputRequest{
			Name:   "Hello",
			Format: FormatRaw,
			Env: []EnvValue{
				{
					Name:  "B",
					Value: "g",
				},
			},
		}

		for i := range viewSpecs {
			err := validateGetViewOutputRequest(viewSpecs[i], getViewOutputRequest)
			require.Error(t, err)
		}
	})

	t.Run("number input", func(t *testing.T) {
		viewSpecs := []domain.ViewSpec{
			{
				Name:    "TIME",
				Command: "echo $DATE",
				Env: []domain.EnvSpec{
					{
						Name:      "DATE",
						Validator: `^[A-Za-z0-9_./-]{2,15}$`,
					},
				},
			},
		}

		getViewOutputRequest := GetViewOutputRequest{
			Name:   "TIME",
			Format: FormatRaw,
			Env: []EnvValue{
				{
					Name:  "DATE",
					Value: "07.07.1988",
				},
			},
		}

		for i := range viewSpecs {
			err := validateGetViewOutputRequest(viewSpecs[i], getViewOutputRequest)
			require.NoError(t, err)
		}
	})

	t.Run("denied characters", func(t *testing.T) {
		viewSpecs := []domain.ViewSpec{
			{
				Name:    "Hello",
				Command: "echo hello $He $Be",
				Env: []domain.EnvSpec{
					{
						Name:      "He",
						Validator: `^[A-Za-z]?([a-z]|\S){2,10}$`,
					},
					{
						Name:      "B",
						Validator: `^[A-Za-z]?([a-z]|\S){2,10}$`,
					},
				},
			},
		}

		getViewOutputRequest := GetViewOutputRequest{
			Name:   "Hello",
			Format: FormatRaw,
			Env: []EnvValue{
				{
					Name:  "Colon",
					Value: "dei-er",
				},
				{
					Name:  "Bye",
					Value: "Ein;sdf",
				},
			},
		}

		for i := range viewSpecs {
			err := validateGetViewOutputRequest(viewSpecs[i], getViewOutputRequest)
			require.Error(t, err)
		}
	})

	t.Run("missing env value", func(t *testing.T) {
		viewSpecs := []domain.ViewSpec{
			{
				Name:    "Hello",
				Command: "echo hello $HELLO $BYE",
				Env: []domain.EnvSpec{
					{
						Name:      "He",
						Validator: `^[A-Za-z]?([a-z]|\S){2,10}$`,
					},
					{
						Name:      "B",  // ^[A-Za-z0-9_.-]{1,15}$
						Validator: `^[A-Za-z]?([a-z]|\S){2,10}$`,
					},
				},
			},
		}

		getViewOutputRequest := GetViewOutputRequest{
			Name:   "Hello",
			Format: FormatRaw,
			Env: []EnvValue{
				{
					Name:  "Hey",
					Value: "",
				},
				{
					Name:  "Bye",
					Value: "gogogo",
				},
			},
		}

		for i := range viewSpecs {
			err := validateGetViewOutputRequest(viewSpecs[i], getViewOutputRequest)
			require.Error(t, err)
		}
	})

	t.Run("test regular expressions from config", func(t *testing.T) {
		viewSpecs := []domain.ViewSpec{
			{
				Name:    "Hello",
				Command: "echo hello $He $Be",
				Env: []domain.EnvSpec{
					{
						Name:      "He",
						Validator: `^[A-Za-z]?([a-z]|\S){2,10}$`,
					},
					{
						Name:      "B",
						Validator: `^[A-Za-z]?([a-z]|\S){2,10}$`,
					},
				},
			},
		}

		getViewOutputRequest := GetViewOutputRequest{
			Name:   "Hello",
			Format: FormatRaw,
			Env: []EnvValue{
				{
					Name:  "Colon",
					Value: "dei,er",
				},
				{
					Name:  "Bye",
					Value: "Ein;sdf",
				},
			},
		}

		for i := range viewSpecs {
			err := validateGetViewOutputRequest(viewSpecs[i], getViewOutputRequest)
			require.Error(t, err)
		}
	})
}