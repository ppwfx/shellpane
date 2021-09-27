package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ppwfx/shellpane/internal/business"
	"github.com/ppwfx/shellpane/internal/utils/errutil"

	"github.com/pkg/errors"
)

type ClientConfig struct {
	Host      string
	BasicAuth BasicAuthConfig
}

type ClientOpts struct {
	Config     ClientConfig
	HttpClient *http.Client
}

type Client struct {
	opts ClientOpts
}

func NewClient(opts ClientOpts) Client {
	return Client{
		opts: opts,
	}
}

func (c Client) GetViewOutput(ctx context.Context, req business.GetViewOutputRequest) (rsp business.GetViewOutputResponse, err error) {
	rawURL := c.opts.Config.Host + RouteGetViewOutput
	URL, err := url.Parse(rawURL)
	if err != nil {
		return business.GetViewOutputResponse{}, errors.Wrapf(err, "failed to parse url=%v", rawURL)
	}

	q := URL.Query()
	q.Set("name", req.Name)
	for i := range req.Env {
		q.Set("env" + req.Env[i].Name, req.Env[i].Value)
	}

	URL.RawQuery = q.Encode()

	err = c.doJsonRequest(ctx, URL.String(), http.MethodGet, nil, &rsp)
	if err != nil {
		return rsp, errors.Wrapf(err, "failed to do json request with url=%v", URL.String())
	}

	return
}

func (c Client) GetViewSpecs(ctx context.Context, req business.GetViewSpecsRequest) (rsp business.GetViewSpecsResponse, err error) {
	rawURL := c.opts.Config.Host + RouteGetViewSpecs
	URL, err := url.Parse(rawURL)
	if err != nil {
		return business.GetViewSpecsResponse{}, errors.Wrapf(err, "failed to parse url=%v", rawURL)
	}

	err = c.doJsonRequest(ctx, URL.String(), http.MethodGet, nil, &rsp)
	if err != nil {
		return rsp, errors.Wrapf(err, "failed to do json request with url=%v", URL.String())
	}

	return
}

func (c Client) doJsonRequest(ctx context.Context, u string, method string, req interface{}, rsp interface{}) error {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(req)
	if err != nil {
		return errors.Wrap(errutil.Encoding(err), "failed to json encode req")
	}

	r, err := http.NewRequestWithContext(ctx, method, u, &b)
	if err != nil {
		return errors.Wrapf(err, "failed to create request for url=%v", u)
	}

	resp, err := c.opts.HttpClient.Do(r)
	if err != nil {
		return errors.Wrapf(errutil.Unknown(err), "failed to do request for url=%v", u)
	}

	err = errutil.ExpectHTTPStatusCode(resp, http.StatusOK)
	if err != nil {
		return errors.Wrapf(err, "unexpected status code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "failed to read response body")
	}

	err = json.Unmarshal(body, &rsp)
	if err != nil {
		return errors.Wrapf(err, "failed to json unmarshal response body=%v", string(body))
	}

	return nil
}
