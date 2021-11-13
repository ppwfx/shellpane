package persistence

import (
	"github.com/ppwfx/shellpane/internal/domain"
)

type RepositoryOpts struct {
	ViewConfigs           []domain.ViewConfig
	UserConfigs           map[string]domain.UserConfig
	UserAllowedViews      map[string]map[string]struct{}
	UserAllowedCategories map[string]map[string]struct{}
	UserAllowedCommands   map[string]map[string]struct{}
	CommandConfigs        map[string]domain.CommandConfig
	CategoryConfigs       []domain.CategoryConfig
}

type Repository struct {
	opts RepositoryOpts
}

func NewRepository(opts RepositoryOpts) Repository {
	return Repository{
		opts: opts,
	}
}

func (r Repository) GetUserAllowedViews() map[string]map[string]struct{} {
	return r.opts.UserAllowedViews
}

func (r Repository) GetUserAllowedCategories() map[string]map[string]struct{} {
	return r.opts.UserAllowedCategories
}

func (r Repository) GetUserAllowedCommands() map[string]map[string]struct{} {
	return r.opts.UserAllowedCommands
}

func (r Repository) GetViewConfigs() []domain.ViewConfig {
	return r.opts.ViewConfigs
}

func (r Repository) GetViewConfigsIn(slugs map[string]struct{}) []domain.ViewConfig {
	var views []domain.ViewConfig
	for i := range r.GetViewConfigs() {
		_, ok := slugs[r.opts.ViewConfigs[i].Slug]
		if !ok {
			continue
		}

		views = append(views, r.opts.ViewConfigs[i])
	}

	return views
}

func (r Repository) GetCategoryConfigs() []domain.CategoryConfig {
	return r.opts.CategoryConfigs
}

func (r Repository) GetCategoryConfigsIn(slugs map[string]struct{}) []domain.CategoryConfig {
	categories := []domain.CategoryConfig{}
	for i := range r.GetCategoryConfigs() {
		_, ok := slugs[r.opts.CategoryConfigs[i].Slug]
		if !ok {
			continue
		}

		categories = append(categories, r.opts.CategoryConfigs[i])
	}

	return categories
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
