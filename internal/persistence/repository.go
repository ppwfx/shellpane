package persistence

import "github.com/ppwfx/shellpane/internal/domain"

type RepositoryConfig struct {
}

type RepositoryOpts struct {
	Config    RepositoryConfig
	ViewSpecs []domain.ViewSpec
}

type Repository struct {
	opts RepositoryOpts
}

func NewRepository(opts RepositoryOpts) Repository {
	return Repository{
		opts: opts,
	}
}

func (r Repository) GetViewSpecs() []domain.ViewSpec {
	return r.opts.ViewSpecs
}

func (r Repository) GetViewSpecByName(name string) (domain.ViewSpec, bool) {
	for i := range r.opts.ViewSpecs {
		if r.opts.ViewSpecs[i].Name == name {
			return r.opts.ViewSpecs[i], true
		}
	}

	return domain.ViewSpec{}, false
}