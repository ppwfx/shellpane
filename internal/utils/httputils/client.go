package httputils

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-limiter"

	"github.com/ppwfx/shellpane/internal/utils/logutil"
)

func copyClient(c *http.Client) *http.Client {
	var t http.RoundTripper
	if c.Transport != nil {
		t = c.Transport

	} else {
		t = http.DefaultTransport
	}

	return &http.Client{
		Transport:     t,
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
		Timeout:       c.Timeout,
	}
}

func withRoundTripper(c *http.Client, m roundTripperMiddleware) *http.Client {
	nc := copyClient(c)

	m.SetNext(nc.Transport)

	nc.Transport = m

	return nc
}

type roundTripperMiddleware interface {
	http.RoundTripper
	SetNext(http.RoundTripper)
}

func WithLogging(c *http.Client, beforeLog func(logutil.LogHTTPResponse) logutil.LogHTTPResponse) *http.Client {
	return withRoundTripper(c, &loggingRoundTripper{BeforeLog: beforeLog})
}

type loggingRoundTripper struct {
	Next      http.RoundTripper
	BeforeLog func(logutil.LogHTTPResponse) logutil.LogHTTPResponse
}

func (t *loggingRoundTripper) SetNext(n http.RoundTripper) {
	t.Next = n
}

func (t loggingRoundTripper) RoundTrip(r *http.Request) (res *http.Response, err error) {
	start := time.Now()

	l, err := logutil.LoggerValue(r.Context())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get logger from context")
	}

	res, err = t.Next.RoundTrip(r)
	if err != nil {
		return res, errors.Wrapf(err, "failed to round trip")
	}

	l = logutil.WithHTTPResponse(l, r, res, start, t.BeforeLog)

	switch {
	case res.StatusCode >= 500:
		l.Error()
		break
	case res.StatusCode >= 400:
		l.Error()
		break
	default:
		l.Info()
		break
	}

	return
}

func WithRateLimiter(c *http.Client, store limiter.Store, checkInterval time.Duration) *http.Client {
	return withRoundTripper(c, &rateLimitRoundTripper{Limiter: store, CheckLimiterInterval: checkInterval})
}

type rateLimitRoundTripper struct {
	Next                 http.RoundTripper
	Limiter              limiter.Store
	CheckLimiterInterval time.Duration
}

func (t *rateLimitRoundTripper) SetNext(n http.RoundTripper) {
	t.Next = n
}

func (t rateLimitRoundTripper) RoundTrip(r *http.Request) (res *http.Response, err error) {
	err = waitForStore(r.Context(), t.Limiter, "default", t.CheckLimiterInterval)
	if err != nil {
		err = errors.Wrapf(err, "failed to apply rate limit")
		return
	}

	return t.Next.RoundTrip(r)
}

func waitForStore(ctx context.Context, store limiter.Store, key string, interval time.Duration) (err error) {
	timer := time.NewTicker(interval)
	for {
		var ok bool
		_, _, _, ok, err = store.Take(ctx, key)
		if err != nil {
			return errors.Wrapf(err, "failed to take from limiter store for key: %s", key)
		}

		if ok {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}

func WithHeader(c *http.Client, h http.Header) *http.Client {
	return withRoundTripper(c, &headerRoundTripper{Header: h})
}

type headerRoundTripper struct {
	Next   http.RoundTripper
	Header http.Header
}

func (t *headerRoundTripper) SetNext(n http.RoundTripper) {
	t.Next = n
}

func (t headerRoundTripper) RoundTrip(r *http.Request) (res *http.Response, err error) {
	for k, v := range t.Header {
		r.Header[k] = v
	}

	return t.Next.RoundTrip(r)
}

type BasicAuthConfig struct {
	Username string
	Password string
}

func WithBasicAuth(c *http.Client, config BasicAuthConfig) *http.Client {
	return withRoundTripper(c, &basicAuthRoundTripper{config: config})
}

type basicAuthRoundTripper struct {
	next   http.RoundTripper
	config BasicAuthConfig
}

func (t *basicAuthRoundTripper) SetNext(n http.RoundTripper) {
	t.next = n
}

func (t basicAuthRoundTripper) RoundTrip(r *http.Request) (res *http.Response, err error) {
	r.SetBasicAuth(t.config.Username, t.config.Password)

	return t.next.RoundTrip(r)
}
