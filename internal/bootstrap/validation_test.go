package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ValidateInputConfigs(t *testing.T) {
	tcs := []struct {
		name      string
		inputs    []InputConfig
		expectErr bool
	}{
		{
			name: "valid",
			inputs: []InputConfig{
				{
					Slug: "A",
				},
				{
					Slug: "B",
				},
			},
			expectErr: false,
		},
		{
			name: "missing slug",
			inputs: []InputConfig{
				{
					Slug: "",
				},
				{
					Slug: "B",
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate slug",
			inputs: []InputConfig{
				{
					Slug: "A",
				},
				{
					Slug: "A",
				},
			},
			expectErr: true,
		},
	}

	for i := range tcs {
		i := i

		t.Run(tcs[i].name, func(t *testing.T) {
			t.Parallel()

			err := validateInputs(tcs[i].inputs)
			if tcs[i].expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_ValidateCommandConfigs(t *testing.T) {
	tcs := []struct {
		name          string
		definedInputs map[string]struct{}
		commands      []CommandConfig
		expectErr     bool
	}{
		{
			name: "valid",
			definedInputs: map[string]struct{}{
				"A": {},
			},
			commands: []CommandConfig{
				{
					Slug:    "A",
					Command: "A",
				},
				{
					Slug:    "B",
					Command: "B",
					Inputs: []CommandInputConfig{
						{
							InputSlug: "A",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "missing slug",
			commands: []CommandConfig{
				{
					Slug:    "",
					Command: "A",
				},
				{
					Slug:    "B",
					Command: "B",
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate slug",
			commands: []CommandConfig{
				{
					Slug:    "A",
					Command: "A",
				},
				{
					Slug:    "A",
					Command: "B",
				},
			},
			expectErr: true,
		},
		{
			name: "missing input slug",
			definedInputs: map[string]struct{}{
				"A": {},
			},
			commands: []CommandConfig{
				{
					Slug:    "A",
					Command: "A",
					Inputs: []CommandInputConfig{
						{
							InputSlug: "",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate input slug",
			definedInputs: map[string]struct{}{
				"A": {},
			},
			commands: []CommandConfig{
				{
					Slug:    "A",
					Command: "A",
					Inputs: []CommandInputConfig{
						{
							InputSlug: "A",
						},
						{
							InputSlug: "A",
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "undefined input slug",
			commands: []CommandConfig{
				{
					Slug:    "A",
					Command: "A",
					Inputs: []CommandInputConfig{
						{
							InputSlug: "B",
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for i := range tcs {
		i := i

		t.Run(tcs[i].name, func(t *testing.T) {
			t.Parallel()

			err := validateCommands(tcs[i].definedInputs, tcs[i].commands)
			if tcs[i].expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_ValidateShellpaneConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		config := ShellpaneConfig{
			Inputs: []InputConfig{
				{
					Slug: "input-a",
				},
			},
			Commands: []CommandConfig{
				{
					Slug:    "command-a",
					Command: "A",
					Inputs: []CommandInputConfig{
						{
							InputSlug: "input-a",
						},
					},
				},
				{
					Slug:    "command-b",
					Command: "B",
				},
			},
			Categories: []CategoryConfig{
				{
					Slug:  "category-a",
					Name:  "A",
					Color: "green",
				},
			},
			Sequences: []SequenceConfig{
				{
					Slug: "sequence-a",
					Steps: []StepConfig{
						{
							Name:        "A",
							CommandSlug: "command-a",
						},
						{
							Name:        "B",
							CommandSlug: "command-b",
						},
					},
				},
			},
			Views: []ViewConfig{
				{
					Slug:         "view-a",
					Name:         "A",
					CategorySlug: "category-a",
					CommandSlug:  "command-a",
				},
				{
					Slug:         "view-b",
					Name:         "B",
					CategorySlug: "category-a",
					SequenceSlug: "sequence-a",
				},
			},
		}

		err := ValidateShellpaneConfig(config)
		require.NoError(t, err)
	})

	t.Run("duplicate command slugs", func(t *testing.T) {
		config := ShellpaneConfig{
			Inputs: []InputConfig{
				{
					Slug: "A",
				},
			},
			Commands: []CommandConfig{
				{
					Slug:    "A",
					Command: "A",
					Inputs: []CommandInputConfig{
						{
							InputSlug: "A",
						},
					},
				},
				{
					Slug:    "A",
					Command: "A",
				},
			},
		}

		err := ValidateShellpaneConfig(config)
		require.Error(t, err)
	})

	t.Run("duplicate input slugs", func(t *testing.T) {
		config := ShellpaneConfig{
			Inputs: []InputConfig{
				{
					Slug: "A",
				},
				{
					Slug: "A",
				},
			},
		}

		err := ValidateShellpaneConfig(config)
		require.Error(t, err)
	})
}
