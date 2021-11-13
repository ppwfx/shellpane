package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/namsral/flag"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/ppwfx/shellpane/internal/bootstrap"
	"github.com/ppwfx/shellpane/internal/communication"
)

func getConfig(args []string) (bootstrap.ContainerConfig, error) {
	fs := flag.FlagSet{}

	conf := bootstrap.ContainerConfig{
		Communication: communication.Config{
			Listener: bootstrap.ListenerNet,
		},
	}

	fs.StringVar(&conf.Logger.Backend, "logger-backend", "", "use a logger that's optimized for a specific logging backend - possible values (stackdriver)")
	fs.StringVar(&conf.Logger.MinLevel, "logger-min-level", "info", "possible values (debug|info|warn|error|panic|fatal)")
	fs.BoolVar(&conf.Logger.UseColor, "logger-use-color", false, "possible values (true|false)")
	fs.BoolVar(&conf.Logger.ReportCaller, "logger-report-caller", false, "possible values (true|false)")
	fs.BoolVar(&conf.Logger.UseJSON, "logger-use-json", false, "possible values (true|false)")
	fs.StringVar(&conf.Communication.HttpAddr, "http-addr", "0.0.0.0:8000", "http address to listen on")
	fs.StringVar(&conf.Communication.Router.BasicAuth.Username, "basic-auth-username", "", "optional: specify a basic auth username")
	fs.StringVar(&conf.Communication.Router.BasicAuth.Password, "basic-auth-password", "info", "optional: specify a basic auth password")

	fs.StringVar(&conf.ShellpaneYAMLPath, "shellpane-yaml-path", "", "path to specs yaml")
	var specsYAML string
	fs.StringVar(&specsYAML, "shellpane-yaml", "", "specs as yaml")

	err := fs.Parse(args)
	if err != nil {
		return conf, errors.Wrapf(err, "failed to parse command line arguments")
	}

	if specsYAML != "" {
		err = yaml.Unmarshal([]byte(specsYAML), &conf.ShellpaneConfig)
		if err != nil {
			return conf, errors.Wrapf(err, "failed to unmarshal shellpane-yaml=%v", specsYAML)
		}
	}

	return conf, nil
}

func main() {
	config, err := getConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to get config: %v", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := bootstrap.NewContainer(bootstrap.ContainerOpts{
		Config: config,
	})
	defer c.Close(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		<-sigCh
		cancel()
		c.Close(ctx)
	}()

	err = func(ctx context.Context) (err error) {
		logger, err := c.GetLogger(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to get logger")
		}

		srv, err := c.GetHTTPServer(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to get server")
		}

		l, err := c.GetHTTPListener(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to get http listener")
		}

		logger.Infof("service listens on addr=%v", config.Communication.HttpAddr)

		err = srv.Serve(l)
		if err != nil {
			return errors.Wrapf(err, "failed to serve")
		}

		return nil
	}(ctx)
	if err != nil {
		log.Fatal("failed to run service: ", err)
	}

	return
}
