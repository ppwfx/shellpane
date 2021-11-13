package bootstrap

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"google.golang.org/grpc/test/bufconn"
	"gopkg.in/yaml.v3"

	"github.com/ppwfx/shellpane/internal/business"
	"github.com/ppwfx/shellpane/internal/communication"
	"github.com/ppwfx/shellpane/internal/domain"
	"github.com/ppwfx/shellpane/internal/persistence"
	"github.com/ppwfx/shellpane/internal/utils/httputils"
	"github.com/ppwfx/shellpane/internal/utils/logutil"
)

const (
	ListenerNet     = "net"
	ListenerBufconn = "bufconn"
	FSMemory        = "memory"
	FSOS            = "os"
)

type ContainerConfig struct {
	Logger            logutil.LoggerConfig
	Business          business.Config
	Communication     communication.Config
	ShellpaneYAMLPath string
	FS                string
	ShellpaneConfig   *ShellpaneConfig
}

type ContainerOpts struct {
	Config ContainerConfig
}

type namedCloser struct {
	name   string
	closer io.Closer
}

type Container struct {
	opts              ContainerOpts
	closers           []namedCloser
	handler           *business.Handler
	router            http.Handler
	httpServer        *http.Server
	httpListener      net.Listener
	httpClient        *http.Client
	roundTripper      http.RoundTripper
	logger            *zap.SugaredLogger
	client            *communication.Client
	repository        *persistence.Repository
	userConfigs       map[string]domain.UserConfig
	viewConfigs       []domain.ViewConfig
	categoryConfigs   []domain.CategoryConfig
	commandsConfig    map[string]domain.CommandConfig
	allowedCategories map[string]map[string]struct{}
	allowedViews      map[string]map[string]struct{}
	allowedCommands   map[string]map[string]struct{}
	fs                afero.Fs
}

func NewContainer(opts ContainerOpts) Container {
	return Container{
		opts: opts,
	}
}

func (c *Container) Close(ctx context.Context) (errs []error) {
	for _, closer := range c.closers {
		log.Printf("closed %v\n", closer.name)
		err := closer.closer.Close()
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "failed to close %v", closer.name))
		}
	}

	return
}

func (c Container) GetHandler(ctx context.Context) (business.Handler, error) {
	if c.handler != nil {
		return *c.handler, nil
	}

	repository, err := c.GetRepository(ctx)
	if err != nil {
		return business.Handler{}, errors.Wrapf(err, "failed to get repository")
	}

	h := business.NewHandler(business.HandlerOpts{
		Config:     c.opts.Config.Business.Handler,
		Repository: repository,
	})

	c.handler = &h

	return *c.handler, nil
}

func (c Container) GetRouter(ctx context.Context) (http.Handler, error) {
	if c.router != nil {
		return c.router, nil
	}

	h, err := c.GetHandler(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get handler")
	}

	_, _, categoriesConfig, _, _, _, _, err := c.GetConfigs(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config")
	}

	router := communication.NewRouter(communication.RouterOpts{
		CategoryConfigs: categoriesConfig,
		Config:          c.opts.Config.Communication.Router,
		Handler:         h,
	})

	logger, err := c.GetLogger(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get logger")
	}

	router = logutil.LogRequestMiddleware(router)
	router = logutil.WithLoggerValueMiddleware(logger)(router)

	if c.opts.Config.Communication.Router.BasicAuth.Username != "" &&
		c.opts.Config.Communication.Router.BasicAuth.Password != "" {
		router = communication.WithBasicAuthMiddleware(router, c.opts.Config.Communication.Router.BasicAuth)
	}

	if c.opts.Config.Communication.UserIDHeader != "" {
		router = communication.WithUserIDMiddleware(router, c.opts.Config.Communication.UserIDHeader)
	}

	router = communication.CorsMiddleware(router)

	c.router = router

	return c.router, nil
}

func (c *Container) GetHTTPServer(ctx context.Context) (*http.Server, error) {
	if c.httpServer != nil {
		return c.httpServer, nil
	}

	r, err := c.GetRouter(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get router")
	}

	c.httpServer = &http.Server{
		Handler:           r,
		ReadTimeout:       1 * time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       5 * time.Second,
	}

	return c.httpServer, nil
}

func (c *Container) GetHTTPListener(ctx context.Context) (net.Listener, error) {
	if c.httpListener != nil {
		return c.httpListener, nil
	}

	switch c.opts.Config.Communication.Listener {
	case ListenerBufconn:
		c.httpListener = BufconnListener{Listener: bufconn.Listen(1024 * 1024)}
	case ListenerNet:
		var err error
		c.httpListener, err = net.Listen("tcp", c.opts.Config.Communication.HttpAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to listen")
		}
	default:
		return nil, errors.Errorf("listener=%v not supported", c.opts.Config.Communication.Listener)
	}
	c.closers = append(c.closers, namedCloser{name: "http listener", closer: c.httpListener})

	return c.httpListener, nil
}

func (c *Container) GetLogger(ctx context.Context) (*zap.SugaredLogger, error) {
	if c.logger != nil {
		return c.logger, nil
	}

	var err error
	c.logger, err = logutil.NewLogger(c.opts.Config.Logger)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create logger")
	}

	return c.logger, nil
}

func (c *Container) GetRoundTripper(ctx context.Context) (http.RoundTripper, error) {
	if c.roundTripper != nil {
		return c.roundTripper, nil
	}

	switch c.opts.Config.Communication.Listener {
	case ListenerBufconn:
		l, err := c.GetHTTPListener(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get http listener")
		}

		bufconnListener, ok := l.(BufconnListener)
		if !ok {
			return nil, errors.New("failed to type cast http listener to BufconnListener")
		}

		transport := &http.Transport{}
		transport.Dial = bufconnListener.Dial
		transport.DialContext = bufconnListener.DialContext

		c.roundTripper = transport
	case ListenerNet:
		c.roundTripper = http.DefaultTransport
	default:
		return nil, errors.Errorf("listener=%v not supported", c.opts.Config.Communication.Listener)
	}

	return c.roundTripper, nil
}

func (c *Container) GetHTTPClient(ctx context.Context) (*http.Client, error) {
	if c.httpClient != nil {
		return c.httpClient, nil
	}

	roundTripper, err := c.GetRoundTripper(ctx)
	if err != nil {
		return c.httpClient, errors.Wrapf(err, "failed to get roundtripper")
	}

	httpClient := &http.Client{
		Transport: roundTripper,
	}

	if c.opts.Config.Communication.Router.BasicAuth.Username != "" &&
		c.opts.Config.Communication.Router.BasicAuth.Password != "" {
		httpClient = httputils.WithBasicAuth(httpClient, httputils.BasicAuthConfig{
			Username: c.opts.Config.Communication.Router.BasicAuth.Username,
			Password: c.opts.Config.Communication.Router.BasicAuth.Password,
		})
	}

	c.httpClient = httpClient

	return c.httpClient, nil
}

func (c *Container) GetClient(ctx context.Context) (communication.Client, error) {
	if c.client != nil {
		return *c.client, nil
	}

	httpClient, err := c.GetHTTPClient(ctx)
	if err != nil {
		return communication.Client{}, errors.Wrapf(err, "failed to get http client")
	}

	client := communication.NewClient(communication.ClientOpts{
		Config: communication.ClientConfig{
			Host: c.opts.Config.Communication.Client.Host,
		},
		HttpClient: httpClient,
	})
	c.client = &client

	return *c.client, nil
}

func (c *Container) GetRepository(ctx context.Context) (persistence.Repository, error) {
	if c.repository != nil {
		return *c.repository, nil
	}

	userConfigs, viewConfigs, categoryConfigs, commandsConfigs, allowedCategories, allowedCommands, allowedViews, err := c.GetConfigs(ctx)
	if err != nil {
		return persistence.Repository{}, errors.Wrapf(err, "failed to get configs")
	}

	repository := persistence.NewRepository(persistence.RepositoryOpts{
		UserConfigs:           userConfigs,
		ViewConfigs:           viewConfigs,
		CommandConfigs:        commandsConfigs,
		CategoryConfigs:       categoryConfigs,
		UserAllowedCategories: allowedCategories,
		UserAllowedViews:      allowedViews,
		UserAllowedCommands:   allowedCommands,
	})

	c.repository = &repository

	return *c.repository, nil
}

func (c *Container) GetConfigs(ctx context.Context) (
	map[string]domain.UserConfig,
	[]domain.ViewConfig,
	[]domain.CategoryConfig,
	map[string]domain.CommandConfig,
	map[string]map[string]struct{},
	map[string]map[string]struct{},
	map[string]map[string]struct{},
	error,
) {
	if c.viewConfigs != nil {
		return c.userConfigs, c.viewConfigs, c.categoryConfigs, c.commandsConfig, c.allowedCategories, c.allowedCommands, c.allowedViews, nil
	}

	fs, err := c.GetFS(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, errors.Wrapf(err, "failed to get filesystem")
	}

	var config ShellpaneConfig
	switch {
	case c.opts.Config.ShellpaneConfig != nil:
		config = *c.opts.Config.ShellpaneConfig
	case c.opts.Config.ShellpaneYAMLPath != "":
		f, err := fs.Open(c.opts.Config.ShellpaneYAMLPath)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, errors.Wrapf(err, "failed to open file=%v", c.opts.Config.ShellpaneYAMLPath)
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, errors.Wrapf(err, "failed to read file=%v", c.opts.Config.ShellpaneYAMLPath)
		}

		err = yaml.Unmarshal(b, &config)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, errors.Wrapf(err, "failed to yaml unmarshal file=%v content=%v", c.opts.Config.ShellpaneYAMLPath, string(b))
		}
	default:
		return nil, nil, nil, nil, nil, nil, nil, errors.New("no config present")
	}

	err = ValidateShellpaneConfig(config)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, errors.Wrapf(err, "failed to validate shellpane config")
	}

	c.userConfigs, c.viewConfigs, c.categoryConfigs, c.commandsConfig, c.allowedCategories, c.allowedCommands, c.allowedViews = generateConfigs(config)

	return c.userConfigs, c.viewConfigs, c.categoryConfigs, c.commandsConfig, c.allowedCategories, c.allowedCommands, c.allowedViews, nil
}

func (c *Container) GetFS(ctx context.Context) (afero.Fs, error) {
	if c.fs != nil {
		return c.fs, nil
	}

	switch c.opts.Config.FS {
	case FSMemory:
		c.fs = afero.NewMemMapFs()
	case FSOS:
		c.fs = afero.NewOsFs()
	default:
		c.fs = afero.NewOsFs()
	}

	return c.fs, nil
}
