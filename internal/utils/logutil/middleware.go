package logutil

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WithLoggerValueMiddleware returns a http middleware that injects a *zap.SugaredLogger into the context
func WithLoggerValueMiddleware(l *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := l.With(
				"requestId", uuid.New().String(),
			)

			r = r.WithContext(WithLoggerValue(r.Context(), l))

			next.ServeHTTP(w, r)
		})
	}
}

// LogRequestMiddleware defines a http middleware logs every requests
func LogRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := MustLoggerValue(r.Context())

		iw := &interceptingWriter{0, http.StatusOK, w}

		start := time.Now()

		next.ServeHTTP(iw, r)

		l = l.With(zap.Object("context.httpRequest", &logHTTPRequest{
			Method:             r.Method,
			URL:                r.URL.String(),
			UserAgent:          r.UserAgent(),
			Referrer:           r.Referer(),
			RemoteIP:           r.RemoteAddr,
			RequestSize:        r.ContentLength,
			ResponseSize:       iw.count,
			ResponseStatusCode: iw.code,
			Latency:            fmt.Sprintf("%.6fs", time.Since(start).Seconds()),
		}))

		switch {
		case iw.code >= 500:
			l.Error()
			break
		case iw.code >= 400:
			l.Warn()
			break
		default:
			l.Info()
			break
		}
	})
}

type logHTTPRequest struct {
	Method             string
	URL                string
	UserAgent          string
	Referrer           string
	RemoteIP           string
	RequestSize        int64
	ResponseSize       int64
	ResponseStatusCode int
	Latency            string
}

func (r *logHTTPRequest) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("method", r.Method)
	e.AddString("url", r.URL)
	e.AddString("userAgent", r.UserAgent)
	e.AddString("referrer", r.Referrer)
	e.AddInt("responseStatusCode", r.ResponseStatusCode)
	e.AddString("remoteIp", r.RemoteIP)
	e.AddInt64("requestSize", r.RequestSize)
	e.AddInt64("responseSize", r.ResponseSize)
	e.AddString("latency", r.Latency)

	return nil
}

type interceptingWriter struct {
	count int64
	code  int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}

func (iw *interceptingWriter) Write(p []byte) (int, error) {
	iw.count += int64(len(p))
	return iw.ResponseWriter.Write(p)
}

type LogHTTPResponse struct {
	Method             string
	URL                string
	UserAgent          string
	Referrer           string
	RequestSize        int64
	ResponseSize       int64
	ResponseStatusCode int
	Latency            string
}

func (r *LogHTTPResponse) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("method", r.Method)
	e.AddString("url", r.URL)
	e.AddString("userAgent", r.UserAgent)
	e.AddString("referrer", r.Referrer)
	e.AddInt("responseStatusCode", r.ResponseStatusCode)
	e.AddInt64("requestSize", r.RequestSize)
	e.AddInt64("responseSize", r.ResponseSize)
	e.AddString("latency", r.Latency)

	return nil
}

func WithHTTPResponse(l *zap.SugaredLogger, r *http.Request, resp *http.Response, start time.Time, beforeLog func(LogHTTPResponse) LogHTTPResponse) *zap.SugaredLogger {
	entry := LogHTTPResponse{
		Method:             r.Method,
		URL:                r.URL.String(),
		UserAgent:          r.UserAgent(),
		Referrer:           r.Referer(),
		RequestSize:        r.ContentLength,
		ResponseSize:       resp.ContentLength,
		ResponseStatusCode: resp.StatusCode,
		Latency:            fmt.Sprintf("%.6fs", time.Since(start).Seconds()),
	}

	entry = beforeLog(entry)

	return l.With("context.httpResponse", entry)
}
