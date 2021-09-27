package internal

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ppwfx/shellpane/internal/bootstrap"
	"github.com/ppwfx/shellpane/internal/business"
	"github.com/ppwfx/shellpane/internal/communication"
	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/persistence"
	"github.com/ppwfx/shellpane/internal/utils/logutil"
)

const (
	NamePrintHello = "print hello"
	NameFailing    = "failing"
	NameWithEnv    = "withenv"
)

var testConfig = bootstrap.ContainerConfig{
	Logger: logutil.LoggerConfig{
		MinLevel:     logutil.LevelDebug,
		ReportCaller: true,
	},
	Persistence: persistence.Config{
		ViewSpecs: []domain.ViewSpec{
			{
				Name:    NamePrintHello,
				Command: `echo hello`,
			},
			{
				Name:    NameFailing,
				Command: `exit 1`,
			},
			{
				Name:    NameWithEnv,
				Command: "echo $FOO",
				Env: []domain.EnvSpec{
					{
						Name: "FOO",
					},
				},
			},
		},
	},
	Communication: communication.Config{
		Client: communication.ClientConfig{
			Host: "http://bufconn",
			BasicAuth: communication.BasicAuthConfig{
				Username: "username",
				Password: "password",
			},
		},
		Router: communication.RouterConfig{
			BasicAuth: communication.BasicAuthConfig{
				Username: "username",
				Password: "password",
			},
		},
		HttpAddr: "bufconn",
		Listener: bootstrap.ListenerBufconn,
	},
}

func Test_Internal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := testConfig

	c := bootstrap.NewContainer(bootstrap.ContainerOpts{
		Config: config,
	})

	go func() {
		srv, err := c.GetHTTPServer(ctx)
		require.NoError(t, err)

		l, err := c.GetHTTPListener(ctx)
		require.NoError(t, err)

		err = srv.Serve(l)
		//require.NoError(t, err)
	}()

	httpClient, err := c.GetHTTPClient(ctx)
	require.NoError(t, err)

	client, err := c.GetClient(ctx)
	require.NoError(t, err)

	t.Run("get web", func(t *testing.T) {
		resp, err := httpClient.Get(config.Communication.Client.Host + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.True(t, strings.Contains(string(b), "</html>"))
	})

	t.Run("get web without basic auth", func(t *testing.T) {
		roundtripper, err := c.GetRoundTripper(ctx)
		require.NoError(t, err)

		httpClient := &http.Client{Transport: roundtripper}

		resp, err := httpClient.Get(config.Communication.Client.Host + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid GetView request with successful command", func(t *testing.T) {
		rsp, err := client.GetViewOutput(ctx, business.GetViewOutputRequest{Name: NamePrintHello})
		require.NoError(t, err)

		expected := domain.ViewOutput{
			Stdout:   "hello\n",
		}

		assert.Equal(t, expected, rsp.Output)
	})

	t.Run("valid GetViewOutput request with failing command", func(t *testing.T) {
		rsp, err := client.GetViewOutput(ctx, business.GetViewOutputRequest{Name: NameFailing})
		require.NoError(t, err)

		expected := domain.ViewOutput{
			ExitCode: 1,
		}

		assert.Equal(t, expected, rsp.Output)
	})

	t.Run("valid GetViewOutput request with env", func(t *testing.T) {
		rsp, err := client.GetViewOutput(ctx, business.GetViewOutputRequest{
			Name: NameWithEnv,
			Env: []business.EnvValue{
				{
					Name: "FOO",
					Value: "bar",
				},
			},
		})
		require.NoError(t, err)

		expected := domain.ViewOutput{
			Stdout:   "bar\n",
		}

		assert.Equal(t, expected, rsp.Output)
	})

	t.Run("valid GetViewSpecs request", func(t *testing.T) {
		rsp, err := client.GetViewSpecs(ctx, business.GetViewSpecsRequest{})
		require.NoError(t, err)

		assert.Equal(t, config.Persistence.ViewSpecs, rsp.Specs)
	})

	errs := c.Close(ctx)
	require.Empty(t, errs)
}
