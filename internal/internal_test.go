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
	"github.com/ppwfx/shellpane/internal/utils/logutil"
)

var baseConfig = bootstrap.ContainerConfig{
	Logger: logutil.LoggerConfig{
		MinLevel:     logutil.LevelDebug,
		ReportCaller: true,
	},
	Communication: communication.Config{
		Client: communication.ClientConfig{
			Host: "http://bufconn",
		},
		HttpAddr: "bufconn",
		Listener: bootstrap.ListenerBufconn,
	},
}

func Test_Web(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := baseConfig

	config.ShellpaneConfig = &bootstrap.ShellpaneConfig{
		Categories: []bootstrap.CategoryConfig{
			{
				Slug:  "A",
				Name:  "A",
				Color: "A",
			},
		},
	}

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

	t.Run("get web", func(t *testing.T) {
		resp, err := httpClient.Get(config.Communication.Client.Host + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.True(t, strings.Contains(string(b), "</html>"))
	})

	t.Run("get categories css", func(t *testing.T) {
		resp, err := httpClient.Get(config.Communication.Client.Host + communication.RouteStaticCategoriesCSS)
		require.NoError(t, err)
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.True(t, strings.Contains(string(b), ".background--"))
	})

	errs := c.Close(ctx)
	require.Empty(t, errs)
}

func Test_BasicAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := baseConfig

	config.ShellpaneConfig = &bootstrap.ShellpaneConfig{
		Categories: []bootstrap.CategoryConfig{
			{
				Slug:  "A",
				Name:  "A",
				Color: "A",
			},
		},
	}

	config.Communication.Client.BasicAuth = communication.BasicAuthConfig{
		Username: "username",
		Password: "password",
	}

	config.Communication.Router.BasicAuth = communication.BasicAuthConfig{
		Username: "username",
		Password: "password",
	}

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

	t.Run("get with basic auth", func(t *testing.T) {
		resp, err := httpClient.Get(config.Communication.Client.Host + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.True(t, strings.Contains(string(b), "</html>"))
	})

	t.Run("get without basic auth", func(t *testing.T) {
		roundtripper, err := c.GetRoundTripper(ctx)
		require.NoError(t, err)

		httpClient := &http.Client{Transport: roundtripper}

		resp, err := httpClient.Get(config.Communication.Client.Host + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	errs := c.Close(ctx)
	require.Empty(t, errs)
}

func Test_UserIDHeader(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := baseConfig

	const userIDHeader = "user-id"
	config.Communication.UserIDHeader = userIDHeader

	config.ShellpaneConfig = &bootstrap.ShellpaneConfig{
		Categories: []bootstrap.CategoryConfig{
			{
				Slug:  "A",
				Name:  "A",
				Color: "A",
			},
		},
	}

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

	t.Run("get web with user id header", func(t *testing.T) {
		roundtripper, err := c.GetRoundTripper(ctx)
		require.NoError(t, err)

		httpClient := &http.Client{Transport: roundtripper}
		r, err := http.NewRequestWithContext(ctx, http.MethodGet, config.Communication.Client.Host+"/", nil)
		require.NoError(t, err)
		r.Header.Set(userIDHeader, "any")

		resp, err := httpClient.Do(r)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("get web without user id header", func(t *testing.T) {
		roundtripper, err := c.GetRoundTripper(ctx)
		require.NoError(t, err)

		httpClient := &http.Client{Transport: roundtripper}

		resp, err := httpClient.Get(config.Communication.Client.Host + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	errs := c.Close(ctx)
	require.Empty(t, errs)
}

func Test_ExecuteCommand(t *testing.T) {
	const (
		CommandPrintHello  = "command print hello"
		CommandFailing     = "command failing"
		CommandWithViewEnv = "command with view env"
		InputFOO           = "FOO"
	)

	t.Parallel()

	ctx := context.Background()

	config := baseConfig

	config.ShellpaneConfig = &bootstrap.ShellpaneConfig{
		Inputs: []bootstrap.InputConfig{
			{
				Slug: "FOO",
			},
		},
		Commands: []bootstrap.CommandConfig{
			{
				Slug:    CommandPrintHello,
				Command: "echo hello",
			},
			{
				Slug:    CommandFailing,
				Command: `exit 1`,
			},
			{
				Slug: CommandWithViewEnv,
				Inputs: []bootstrap.CommandInputConfig{
					{
						InputSlug: InputFOO,
					},
				},
				Command: "echo $FOO",
			},
		},
	}

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

	client, err := c.GetClient(ctx)
	require.NoError(t, err)

	t.Run("valid request with successful command", func(t *testing.T) {
		rsp, err := client.ExecuteCommand(ctx, business.ExecuteCommandRequest{
			Slug: CommandPrintHello,
		})
		require.NoError(t, err)

		expected := business.CommandOutput{
			Stdout: "hello\n",
		}

		assert.Equal(t, expected, rsp.Output)
	})

	t.Run("valid request with failing command", func(t *testing.T) {
		rsp, err := client.ExecuteCommand(ctx, business.ExecuteCommandRequest{
			Slug: CommandFailing,
		})
		require.NoError(t, err)

		expected := business.CommandOutput{
			ExitCode: 1,
		}

		assert.Equal(t, expected, rsp.Output)
	})

	t.Run("valid request with input", func(t *testing.T) {
		rsp, err := client.ExecuteCommand(ctx, business.ExecuteCommandRequest{
			Slug: CommandWithViewEnv,
			Inputs: []business.InputValue{
				{
					Name:  InputFOO,
					Value: "bar",
				},
			},
		})
		require.NoError(t, err)

		expected := business.CommandOutput{
			Stdout: "bar\n",
		}

		assert.Equal(t, expected, rsp.Output)
	})

	errs := c.Close(ctx)
	require.Empty(t, errs)
}

//func Test_GetViewConfigs(t *testing.T) {
//	t.Parallel()
//
//	ctx := context.Background()
//
//	config := baseConfig
//
//	config.ShellpaneConfig = &bootstrap.ShellpaneConfig{
//		Inputs: []bootstrap.InputConfig{
//			{
//				Slug: "B",
//			},
//		},
//		Commands: []bootstrap.CommandConfig{
//			{
//				Slug:    "A",
//				Command: "A",
//			},
//			{
//				Slug:    "B",
//				Command: "B",
//				Inputs: []bootstrap.CommandInputConfig{
//					{
//						InputSlug: "B",
//					},
//				},
//			},
//		},
//		Sequences: []bootstrap.SequenceConfig{
//			{
//				Slug: "A",
//				Steps: []bootstrap.StepConfig{
//					{
//						Name:        "A",
//						CommandSlug: "A",
//					},
//					{
//						Name:        "B",
//						CommandSlug: "B",
//					},
//				},
//			},
//		},
//		Views: []bootstrap.ViewConfig{
//			{
//				Name:        "A",
//				CommandSlug: "A",
//			},
//			{
//				Name:         "B",
//				SequenceSlug: "A",
//			},
//		},
//	}
//
//	c := bootstrap.NewContainer(bootstrap.ContainerOpts{
//		Config: config,
//	})
//
//	go func() {
//		srv, err := c.GetHTTPServer(ctx)
//		require.NoError(t, err)
//
//		l, err := c.GetHTTPListener(ctx)
//		require.NoError(t, err)
//
//		err = srv.Serve(l)
//		//require.NoError(t, err)
//	}()
//
//	client, err := c.GetClient(ctx)
//	require.NoError(t, err)
//
//	t.Run("valid request", func(t *testing.T) {
//		rsp, err := client.GetViewConfigs(ctx, business.GetViewConfigsRequest{})
//		require.NoError(t, err)
//
//		expected := []domain.ViewConfig{
//			{
//				Name: "A",
//				Command: domain.CommandConfig{
//					Slug:    "A",
//					Command: "A",
//				},
//			},
//			{
//				Name: "B",
//				Sequence: domain.SequenceConfig{
//					Slug: "A",
//					Steps: []domain.StepConfig{
//						{
//							Name: "A",
//							Command: domain.CommandConfig{
//								Slug:    "A",
//								Command: "A",
//							},
//						},
//						{
//							Name: "B",
//							Command: domain.CommandConfig{
//								Slug:    "B",
//								Command: "B",
//								Inputs: []domain.CommandInputConfig{
//									{
//										Input: domain.InputConfig{
//											Slug: "B",
//										},
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		}
//
//		assert.Equal(t, expected, rsp.ViewConfigs)
//	})
//
//	errs := c.Close(ctx)
//	require.Empty(t, errs)
//}

func Test_GetCategoryConfigs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := baseConfig

	config.ShellpaneConfig = &bootstrap.ShellpaneConfig{
		Categories: []bootstrap.CategoryConfig{
			{
				Slug:  "A",
				Name:  "A",
				Color: "A",
			},
			{
				Slug:  "B",
				Name:  "B",
				Color: "B",
			},
		},
	}

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

	client, err := c.GetClient(ctx)
	require.NoError(t, err)

	t.Run("valid request", func(t *testing.T) {
		rsp, err := client.GetCategoryConfigs(ctx, business.GetCategoryConfigsRequest{})
		require.NoError(t, err)

		expected := []domain.CategoryConfig{
			{
				Slug:  "A",
				Name:  "A",
				Color: "A",
			},
			{
				Slug:  "B",
				Name:  "B",
				Color: "B",
			},
		}

		assert.Equal(t, expected, rsp.CategoryConfigs)
	})

	errs := c.Close(ctx)
	require.Empty(t, errs)
}

func Test_Permissions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	config := baseConfig
	const userIDHeader = "user-id"
	config.Communication.UserIDHeader = userIDHeader

	config.ShellpaneConfig = &bootstrap.ShellpaneConfig{
		Users: []bootstrap.UserConfig{
			{
				ID: "user-a",
				Groups: []bootstrap.UserGroupConfig{
					{
						GroupSlug: "group-a",
					},
				},
			},
			{
				ID: "user-b",
				Groups: []bootstrap.UserGroupConfig{
					{
						GroupSlug: "group-b",
					},
				},
			},
			{
				ID: "user-all",
				Groups: []bootstrap.UserGroupConfig{
					{
						GroupSlug: "group-all",
					},
				},
			},
		},
		Groups: []bootstrap.GroupConfig{
			{
				Slug: "group-a",
				Roles: []bootstrap.GroupRoleConfig{
					{
						RoleSlug: "role-see-a-category",
					},
				},
			},
			{
				Slug: "group-b",
				Roles: []bootstrap.GroupRoleConfig{
					{
						RoleSlug: "role-see-b-view",
					},
				},
			},
			{
				Slug: "group-all",
				Roles: []bootstrap.GroupRoleConfig{
					{
						RoleSlug: "role-see-a-category",
					},
					{
						RoleSlug: "role-see-b-category",
					},
				},
			},
		},
		Roles: []bootstrap.RoleConfig{
			{
				Slug: "role-see-a-category",
				Categories: []bootstrap.RoleCategoryConfig{
					{
						CategorySlug: "category-a",
					},
				},
			},
			{
				Slug: "role-see-b-view",
				Views: []bootstrap.RoleViewConfig{
					{
						ViewSlug: "view-b",
					},
				},
			},
			{
				Slug: "role-see-b-category",
				Categories: []bootstrap.RoleCategoryConfig{
					{
						CategorySlug: "category-b",
					},
				},
			},
		},
		Categories: []bootstrap.CategoryConfig{
			{
				Slug:  "category-a",
				Name:  "a",
				Color: "a",
			},
			{
				Slug:  "category-b",
				Name:  "b",
				Color: "b",
			},
		},
		Views: []bootstrap.ViewConfig{
			{
				Slug:         "view-a",
				Name:         "a",
				CommandSlug:  "command-a",
				CategorySlug: "category-a",
			},
			{
				Slug:         "view-b",
				Name:         "b",
				CommandSlug:  "command-b",
				CategorySlug: "category-b",
			},
		},
		Commands: []bootstrap.CommandConfig{
			{
				Slug:    "command-a",
				Command: "a",
			},
			{
				Slug:    "command-b",
				Command: "b",
			},
		},
	}

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

	client, err := c.GetClient(ctx)
	require.NoError(t, err)

	t.Run("get categories", func(t *testing.T) {
		t.Run("with user a", func(t *testing.T) {
			rsp, err := client.WithUserID(userIDHeader, "user-a").GetCategoryConfigs(ctx, business.GetCategoryConfigsRequest{})
			require.NoError(t, err)

			expected := []domain.CategoryConfig{
				{
					Slug:  "category-a",
					Name:  "a",
					Color: "a",
				},
			}

			assert.Equal(t, expected, rsp.CategoryConfigs)
		})

		t.Run("with user b", func(t *testing.T) {
			rsp, err := client.WithUserID(userIDHeader, "user-b").GetCategoryConfigs(ctx, business.GetCategoryConfigsRequest{})
			require.NoError(t, err)

			expected := []domain.CategoryConfig{
				{
					Slug:  "category-b",
					Name:  "b",
					Color: "b",
				},
			}

			assert.Equal(t, expected, rsp.CategoryConfigs)
		})

		t.Run("with user all", func(t *testing.T) {
			rsp, err := client.WithUserID(userIDHeader, "user-all").GetCategoryConfigs(ctx, business.GetCategoryConfigsRequest{})
			require.NoError(t, err)

			expected := []domain.CategoryConfig{
				{
					Slug:  "category-a",
					Name:  "a",
					Color: "a",
				},
				{
					Slug:  "category-b",
					Name:  "b",
					Color: "b",
				},
			}

			assert.Equal(t, expected, rsp.CategoryConfigs)
		})
	})

	t.Run("get views", func(t *testing.T) {
		t.Run("with user a", func(t *testing.T) {
			rsp, err := client.WithUserID(userIDHeader, "user-a").GetViewConfigs(ctx, business.GetViewConfigsRequest{})
			require.NoError(t, err)

			expected := []domain.ViewConfig{
				{
					Slug: "view-a",
					Name: "a",
					Command: domain.CommandConfig{
						Slug:    "command-a",
						Command: "a",
					},
					Category: domain.CategoryConfig{
						Slug:  "category-a",
						Name:  "a",
						Color: "a",
					},
				},
			}

			assert.Equal(t, expected, rsp.ViewConfigs)
		})

		t.Run("with user b", func(t *testing.T) {
			rsp, err := client.WithUserID(userIDHeader, "user-b").GetViewConfigs(ctx, business.GetViewConfigsRequest{})
			require.NoError(t, err)

			expected := []domain.ViewConfig{
				{
					Slug: "view-b",
					Name: "b",
					Command: domain.CommandConfig{
						Slug:    "command-b",
						Command: "b",
					},
					Category: domain.CategoryConfig{
						Slug:  "category-b",
						Name:  "b",
						Color: "b",
					},
				},
			}

			assert.Equal(t, expected, rsp.ViewConfigs)
		})

		t.Run("with user all", func(t *testing.T) {
			rsp, err := client.WithUserID(userIDHeader, "user-all").GetViewConfigs(ctx, business.GetViewConfigsRequest{})
			require.NoError(t, err)

			expected := []domain.ViewConfig{
				{
					Slug: "view-a",
					Name: "a",
					Command: domain.CommandConfig{
						Slug:    "command-a",
						Command: "a",
					},
					Category: domain.CategoryConfig{
						Slug:  "category-a",
						Name:  "a",
						Color: "a",
					},
				},
				{
					Slug: "view-b",
					Name: "b",
					Command: domain.CommandConfig{
						Slug:    "command-b",
						Command: "b",
					},
					Category: domain.CategoryConfig{
						Slug:  "category-b",
						Name:  "b",
						Color: "b",
					},
				},
			}

			assert.Equal(t, expected, rsp.ViewConfigs)
		})
	})

	t.Run("execute", func(t *testing.T) {
		t.Run("command a", func(t *testing.T) {
			t.Run("with user a", func(t *testing.T) {
				_, err := client.WithUserID(userIDHeader, "user-a").ExecuteCommand(ctx, business.ExecuteCommandRequest{
					Slug: "command-a",
				})
				require.NoError(t, err)
			})

			t.Run("with user b", func(t *testing.T) {
				_, err := client.WithUserID(userIDHeader, "user-b").ExecuteCommand(ctx, business.ExecuteCommandRequest{
					Slug: "command-a",
				})
				require.Error(t, err)
			})

			t.Run("with user all", func(t *testing.T) {
				_, err := client.WithUserID(userIDHeader, "user-all").ExecuteCommand(ctx, business.ExecuteCommandRequest{
					Slug: "command-a",
				})
				require.NoError(t, err)
			})
		})

		t.Run("command b", func(t *testing.T) {
			t.Run("with user a", func(t *testing.T) {
				_, err := client.WithUserID(userIDHeader, "user-a").ExecuteCommand(ctx, business.ExecuteCommandRequest{
					Slug: "command-b",
				})
				require.Error(t, err)
			})

			t.Run("with user b", func(t *testing.T) {
				_, err := client.WithUserID(userIDHeader, "user-b").ExecuteCommand(ctx, business.ExecuteCommandRequest{
					Slug: "command-b",
				})
				require.NoError(t, err)
			})

			t.Run("with user all", func(t *testing.T) {
				_, err := client.WithUserID(userIDHeader, "user-all").ExecuteCommand(ctx, business.ExecuteCommandRequest{
					Slug: "command-b",
				})
				require.NoError(t, err)
			})
		})
	})

	errs := c.Close(ctx)
	require.Empty(t, errs)
}

//
//func Test_StepViews(t *testing.T) {
//	const (
//		StepsPrintHello      = "steps print hello"
//		StepsFailing         = "steps failing"
//		StepsWithStepEnv     = "steps withstepenv"
//		StepsWithViewEnv     = "steps withviewenv"
//		StepsWithViewStepEnv = "steps withviewstepenv"
//
//		StepExit  = "exit"
//		StepPrint = "print"
//	)
//
//	t.Parallel()
//
//	ctx := context.Background()
//
//	config := baseConfig
//
//	config.Persistence.ViewSpecs = []domain.ViewSpec{
//		{
//			Name: StepsPrintHello,
//			Steps: []domain.Step{
//				{
//					Name:    StepPrint,
//					Command: `echo hello`,
//				},
//			},
//		},
//		{
//			Name: StepsFailing,
//			Steps: []domain.Step{
//				{
//					Name:    StepExit,
//					Command: `exit 1`,
//				},
//			},
//		},
//		{
//			Name: StepsWithStepEnv,
//			Steps: []domain.Step{
//				{
//					Name:    StepPrint,
//					Env: []domain.EnvSpec{
//						{
//							Name: "FOO",
//						},
//					},
//					Command: "echo $FOO",
//				},
//			},
//		},
//		{
//			Name: StepsWithViewEnv,
//			Env: []domain.EnvSpec{
//				{
//					Name: "FOO",
//				},
//			},
//			Steps: []domain.Step{
//				{
//					Name:    StepPrint,
//					Command: "echo $FOO",
//				},
//			},
//		},
//		{
//			Name: StepsWithViewStepEnv,
//			Env: []domain.EnvSpec{
//				{
//					Name: "FOO",
//				},
//			},
//			Steps: []domain.Step{
//				{
//					Name:    StepPrint,
//					Env: []domain.EnvSpec{
//						{
//							Name: "BAR",
//						},
//					},
//					Command: "echo $FOO $BAR",
//				},
//			},
//		},
//	}
//
//	c := bootstrap.NewContainer(bootstrap.ContainerOpts{
//		Config: config,
//	})
//
//	go func() {
//		srv, err := c.GetHTTPServer(ctx)
//		require.NoError(t, err)
//
//		l, err := c.GetHTTPListener(ctx)
//		require.NoError(t, err)
//
//		err = srv.Serve(l)
//		//require.NoError(t, err)
//	}()
//
//	client, err := c.GetClient(ctx)
//	require.NoError(t, err)
//
//	t.Run("valid GetStepOutput request with successful command", func(t *testing.T) {
//		rsp, err := client.GetStepOutput(ctx, business.GetStepOutputRequest{
//			ViewName: StepsPrintHello,
//			StepName: StepPrint,
//		})
//		require.NoError(t, err)
//
//		expected := business.CommandOutput{
//			Stdout: "hello\n",
//		}
//
//		assert.Equal(t, expected, rsp.Output)
//	})
//
//	t.Run("valid GetStepOutput request with failing command", func(t *testing.T) {
//		rsp, err := client.GetStepOutput(ctx, business.GetStepOutputRequest{
//			ViewName: StepsFailing,
//			StepName: StepExit,
//		})
//		require.NoError(t, err)
//
//		expected := business.CommandOutput{
//			ExitCode: 1,
//		}
//
//		assert.Equal(t, expected, rsp.Output)
//	})
//
//	t.Run("valid GetStepOutput request with step env", func(t *testing.T) {
//		rsp, err := client.GetStepOutput(ctx, business.GetStepOutputRequest{
//			ViewName: StepsWithStepEnv,
//			StepName: StepPrint,
//			StepEnv: []business.EnvValue{
//				{
//					Name:  "FOO",
//					Value: "bar",
//				},
//			},
//		})
//		require.NoError(t, err)
//
//		expected := business.CommandOutput{
//			Stdout: "bar\n",
//		}
//
//		assert.Equal(t, expected, rsp.Output)
//	})
//
//	t.Run("valid GetStepOutput request with view env", func(t *testing.T) {
//		rsp, err := client.GetStepOutput(ctx, business.GetStepOutputRequest{
//			ViewName: StepsWithViewEnv,
//			StepName: StepPrint,
//			ViewEnv: []business.EnvValue{
//				{
//					Name:  "FOO",
//					Value: "bar",
//				},
//			},
//		})
//		require.NoError(t, err)
//
//		expected := business.CommandOutput{
//			Stdout: "bar\n",
//		}
//
//		assert.Equal(t, expected, rsp.Output)
//	})
//
//	t.Run("valid GetStepOutput request with view and step env", func(t *testing.T) {
//		rsp, err := client.GetStepOutput(ctx, business.GetStepOutputRequest{
//			ViewName: StepsWithViewStepEnv,
//			StepName: StepPrint,
//			ViewEnv: []business.EnvValue{
//				{
//					Name:  "FOO",
//					Value: "foo",
//				},
//			},
//			StepEnv: []business.EnvValue{
//				{
//					Name:  "BAR",
//					Value: "bar",
//				},
//			},
//		})
//		require.NoError(t, err)
//
//		expected := business.CommandOutput{
//			Stdout: "foo bar\n",
//		}
//
//		assert.Equal(t, expected, rsp.Output)
//	})
//
//	t.Run("valid GetConfigs request", func(t *testing.T) {
//		rsp, err := client.GetViewSpecs(ctx, business.GetViewSpecsRequest{})
//		require.NoError(t, err)
//
//		assert.Equal(t, config.Persistence.ViewSpecs, rsp.Specs)
//	})
//
//	errs := c.Close(ctx)
//	require.Empty(t, errs)
//}
