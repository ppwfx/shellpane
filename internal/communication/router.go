package communication

import (
	"embed"
	"net/http"
	"net/url"
	"strings"

	"github.com/ppwfx/shellpane/internal/business"
	"github.com/ppwfx/shellpane/internal/utils/errutil"
)

var (
	//go:embed web/build/*
	webFS embed.FS
)

type Config struct {
	HttpAddr string
	Listener string
	Router   RouterConfig
	Client   ClientConfig
}

type RouterConfig struct {
	BasicAuth BasicAuthConfig
}

type RouterOpts struct {
	Config  RouterConfig
	Handler business.Handler
}

const (
	RouteGetViewOutput = "/getViewOutput"
	RouteGetViewSpecs  = "/getViewSpecs"
)

func AddPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := prefix + r.URL.Path
		rp := prefix + r.URL.RawPath
		if len(p) > len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) > len(r.URL.RawPath)) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp

			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	})
}

func NewRouter(opts RouterOpts) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", AddPrefix("/web/build", http.FileServer(http.FS(webFS))))

	mux.HandleFunc(RouteGetViewOutput, func(w http.ResponseWriter, r *http.Request) {
		var req business.GetViewOutputRequest
		req.Name = r.URL.Query().Get("name")
		req.Format = r.URL.Query().Get("format")

		for k, v := range r.URL.Query() {
			if !strings.HasPrefix(k, "env") {
				continue
			}

			req.Env = append(req.Env, business.EnvValue{
				Name:  strings.TrimPrefix(k, "env"),
				Value: strings.Join(v, ""),
			})
		}

		rsp, err := opts.Handler.GetViewOutput(r.Context(), req)

		switch {
		case err == nil && req.Format == business.FormatRaw:
			_, _ = w.Write([]byte(rsp.Output.Stdout))
		default:
			errutil.HandleJSONResponse(w, r, rsp, err)
		}

		return
	})

	mux.HandleFunc(RouteGetViewSpecs, errutil.HandlerFuncJSON(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		var req business.GetViewSpecsRequest

		return opts.Handler.GetViewSpecs(r.Context(), req)
	}))

	return mux
}
