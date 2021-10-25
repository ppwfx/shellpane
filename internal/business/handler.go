package business

import (
	"context"

	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/persistence"
	"github.com/ppwfx/shellpane/internal/utils/errutil"
)

type HandlerConfig struct {
}

type HandlerOpts struct {
	Config     HandlerConfig
	Repository persistence.Repository
}

type Handler struct {
	opts HandlerOpts
}

func NewHandler(opts HandlerOpts) Handler {
	return Handler{
		opts: opts,
	}
}

type GetViewConfigsRequest struct {
}

type GetViewConfigsResponse struct {
	errutil.Response
	ViewConfigs []domain.ViewConfig
}

func (h Handler) GetViewConfigs(ctx context.Context, req GetViewConfigsRequest) (GetViewConfigsResponse, error) {
	return GetViewConfigsResponse{ViewConfigs: h.opts.Repository.GetViewConfigs()}, nil
}

type GetCategoryConfigsRequest struct {
}

type GetCategoryConfigsResponse struct {
	errutil.Response
	CategoryConfigs []domain.CategoryConfig
}

func (h Handler) GetCategoryConfigs(ctx context.Context, req GetCategoryConfigsRequest) (GetCategoryConfigsResponse, error) {
	return GetCategoryConfigsResponse{CategoryConfigs: h.opts.Repository.GetCategoryConfigs()}, nil
}
