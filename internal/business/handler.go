package business

import (
	"context"

	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/persistence"
	"github.com/ppwfx/shellpane/internal/utils/errutil"
	"github.com/ppwfx/shellpane/internal/utils/logutil"
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
	log := logutil.MustLoggerValue(ctx)

	userID := UserID(ctx)
	log = log.With("userID", userID)

	var views []domain.ViewConfig
	switch {
	case userID == "":
		views = h.opts.Repository.GetViewConfigs()
	case userID != "":
		allowedViews := h.opts.Repository.GetUserAllowedViews()[userID]
		log = log.With("allowedViews", allowedViews)

		views = h.opts.Repository.GetViewConfigsIn(allowedViews)
	}

	log.With("views", views).
		Debug()

	return GetViewConfigsResponse{ViewConfigs: views}, nil
}

type GetCategoryConfigsRequest struct {
}

type GetCategoryConfigsResponse struct {
	errutil.Response
	CategoryConfigs []domain.CategoryConfig
}

func (h Handler) GetCategoryConfigs(ctx context.Context, req GetCategoryConfigsRequest) (GetCategoryConfigsResponse, error) {
	userID := UserID(ctx)
	var categories []domain.CategoryConfig
	switch {
	case userID == "":
		categories = h.opts.Repository.GetCategoryConfigs()
	case userID != "":
		allowedCategories := h.opts.Repository.GetUserAllowedCategories()[userID]

		categories = h.opts.Repository.GetCategoryConfigsIn(allowedCategories)
	}

	for i := range categories {
		categories[i].Views = nil
	}

	return GetCategoryConfigsResponse{CategoryConfigs: categories}, nil
}
