package persistence

import (
	"github.com/ppwfx/shellpane/internal/domain"
)

type RepositoryOpts struct {
	ViewConfigs     []domain.ViewConfig
	CommandConfigs  map[string]domain.CommandConfig
	CategoryConfigs []domain.CategoryConfig
}

type Repository struct {
	opts RepositoryOpts
}

func NewRepository(opts RepositoryOpts) Repository {
	return Repository{
		opts: opts,
	}
}

func (r Repository) GetViewConfigs() []domain.ViewConfig {
	return r.opts.ViewConfigs
}

func (r Repository) GetCategoryConfigs() []domain.CategoryConfig {
	return r.opts.CategoryConfigs
}

func (r Repository) GetViewConfig(name string) (domain.ViewConfig, bool) {
	for i := range r.opts.ViewConfigs {
		if r.opts.ViewConfigs[i].Name == name {
			return r.opts.ViewConfigs[i], true
		}
	}

	return domain.ViewConfig{}, false
}

func (r Repository) GetCommandConfig(slug string) (domain.CommandConfig, bool) {
	command, ok := r.opts.CommandConfigs[slug]

	return command, ok
}
