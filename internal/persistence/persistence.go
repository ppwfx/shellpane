package persistence

import "github.com/ppwfx/shellpane/internal/domain"

type Config struct {
	ViewSpecsYAMLPath string
	ViewSpecs         []domain.ViewSpec
	FS                string
	Repository        RepositoryConfig
}

