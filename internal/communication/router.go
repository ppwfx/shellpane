package communication

import (
	"net/http"
	"strings"

	"github.com/ppwfx/shellpane/internal/business"
	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/utils/errutil"
)

type RouterConfig struct {
	BasicAuth BasicAuthConfig
}

type RouterOpts struct {
	Config          RouterConfig
	Handler         business.Handler
	CategoryConfigs []domain.CategoryConfig
}

const (
	RouteExecuteCommand      = "/executeCommand"
	RouteGetViewConfigs      = "/getViewConfigs"
	RouteGetCategoryConfigs  = "/getCategoryConfigs"
	RouteStaticCategoriesCSS = "/static/categories.css"
)

func NewRouter(opts RouterOpts) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", webHandler)

	mux.HandleFunc(RouteExecuteCommand, func(w http.ResponseWriter, r *http.Request) {
		var req business.ExecuteCommandRequest
		req.Slug = r.URL.Query().Get("slug")
		req.Format = r.URL.Query().Get("format")

		for k, v := range r.URL.Query() {
			if !strings.HasPrefix(k, "input_") {
				continue
			}

			req.Inputs = append(req.Inputs, business.InputValue{
				Name:  strings.TrimPrefix(k, "input_"),
				Value: strings.Join(v, ""),
			})
		}

		rsp, err := opts.Handler.ExecuteCommand(r.Context(), req)

		switch {
		case err == nil && req.Format == business.FormatRaw:
			_, _ = w.Write([]byte(rsp.Output.Stdout))
		default:
			errutil.HandleJSONResponse(w, r, rsp, err)
		}

		return
	})

	mux.HandleFunc(RouteGetViewConfigs, errutil.HandlerFuncJSON(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		var req business.GetViewConfigsRequest

		return opts.Handler.GetViewConfigs(r.Context(), req)
	}))

	mux.HandleFunc(RouteGetCategoryConfigs, errutil.HandlerFuncJSON(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		var req business.GetCategoryConfigsRequest

		return opts.Handler.GetCategoryConfigs(r.Context(), req)
	}))

	mux.HandleFunc(RouteStaticCategoriesCSS, getCategoriesCSSHandler(opts.CategoryConfigs))

	return mux
}
