package communication

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"net/http"
	"net/url"

	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/utils/logutil"
)

var (
	//go:embed web/build/*
	webFS embed.FS
)

var webHandler = AddPrefix("/web/build", http.FileServer(http.FS(webFS)))

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

var cssTemplate = template.Must(template.New("").Parse(`
{{range $category := . }}
.background--{{$category.Slug}} {
	background-color: {{$category.Color}}0C;
}

.input-background--{{$category.Slug}} input[type="text"] {
	background-color: {{$category.Color}}4C;
	outline-color: {{$category.Color}};
}

.a-color--{{$category.Slug}} a, .a-color--{{$category.Slug}} a:hover {
	color: {{$category.Color}};
}

.scrollbar-color--{{$category.Slug}}::-webkit-scrollbar-thumb {
  background: {{$category.Color}};
}

.flash-border--{{$category.Slug}}0 {
  animation-name: flash-border--{{$category.Slug}}0;
  animation-duration: 0.5s;
  animation-timing-function: linear;
  animation-iteration-count: 1;
}

.flash-border--{{$category.Slug}}1 {
  animation-name: flash-border--{{$category.Slug}}1;
  animation-duration: 0.5s;
  animation-timing-function: linear;
  animation-iteration-count: 1;
}

@keyframes flash-border--{{$category.Slug}}0 {
  0% {
    border-color: {{$category.Color}};
  }
  99% {
    border-color: #f3f3f3;
  }
}

@keyframes flash-border--{{$category.Slug}}1 {
  0% {
    border-color: {{$category.Color}};
  }
  100% {
    border-color: #f3f3f3;
  }
}

.loader--{{$category.Slug}} {
  width: 14px;
  height: 14px;
  border: 2px solid {{$category.Color}};
  border-right-color: transparent;
  border-radius: 50%;
  animation: loader-rotate 1s linear infinite;
}

{{end}}
`))

func getCategoriesCSSHandler(cs []domain.CategoryConfig) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logutil.MustLoggerValue(r.Context())

		var b bytes.Buffer
		err := cssTemplate.Execute(&b, cs)
		if err != nil {
			log.Errorf("failed to execute template: %v", err.Error())

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.Header().Set("content-type", "text/css; charset=utf-8")

		_, err = io.Copy(w, &b)
		if err != nil {
			log.Errorf("failed copy io: %v", err.Error())

			return
		}
	}
}
