package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ppwfx/shellpane/internal/bootstrap"
	"github.com/ppwfx/shellpane/internal/bootstrap/convert"
	"github.com/ppwfx/shellpane/internal/communication"
)

func main() {
	var specsYAML string
	conf := bootstrap.ContainerConfig{
		Communication: communication.Config{
			Listener: bootstrap.ListenerNet,
		},
	}

	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			if specsYAML != "" {
				err := yaml.Unmarshal([]byte(specsYAML), &conf.ShellpaneConfig)
				if err != nil {
					return errors.Wrapf(err, "failed to unmarshal shellpane-yaml=%v", specsYAML)
				}
			}

			return serve(cmd.Context(), conf)
		},
	}
	serveCmd.Flags().StringVarP(&conf.Communication.HttpAddr, "http-addr", "", "0.0.0.0:8000", "http address to listen on")
	serveCmd.Flags().StringVarP(&conf.Communication.Router.BasicAuth.Username, "basic-auth-username", "", "", "optional: specify a basic auth username")
	serveCmd.Flags().StringVarP(&conf.Communication.Router.BasicAuth.Password, "basic-auth-password", "", "", "optional: specify a basic auth password")
	serveCmd.Flags().StringVarP(&conf.Communication.UserIDHeader, "user-id-header", "", "", "optional: name of user id header; implicitly activates permissions")
	serveCmd.Flags().StringVarP(&conf.Communication.DefaultUserID, "default-user-id", "", "", "optional: specify a default user id, if the user id header is not set or the user id header is empty")
	serveCmd.Flags().StringVarP(&conf.Communication.CorsOrigin, "cors-origin", "", "", "optional: specify a cors origin")
	serveCmd.Flags().StringVarP(&conf.ShellpaneYAMLPath, "shellpane-yaml-path", "", "", "path to specs yaml")
	serveCmd.Flags().StringVarP(&specsYAML, "shellpane-yaml", "", "", "specs as yaml")

	var fromSwagger2JSONPath string
	var outputPath string
	var setCategory string
	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "generate",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate(fromSwagger2JSONPath, outputPath, setCategory)
		},
	}
	generateCmd.Flags().StringVarP(&fromSwagger2JSONPath, "from-swagger2-json-path", "", "", "")
	generateCmd.Flags().StringVarP(&outputPath, "output-path", "o", "", "")
	generateCmd.Flags().StringVarP(&setCategory, "set-category", "", "", "")

	var rootCmd = &cobra.Command{Use: "shellpane"}
	rootCmd.Flags().StringVarP(&conf.Logger.Backend, "logger-backend", "", "", "use a logger that's optimized for a specific logging backend (stackdriver)")
	rootCmd.Flags().StringVarP(&conf.Logger.MinLevel, "logger-min-level", "", "info", "(debug|info|warn|error|panic|fatal)")
	rootCmd.Flags().BoolVar(&conf.Logger.UseColor, "logger-use-color", false, "(true|false)")
	rootCmd.Flags().BoolVar(&conf.Logger.ReportCaller, "logger-report-caller", false, "(true|false)")
	rootCmd.Flags().BoolVar(&conf.Logger.UseJSON, "logger-use-json", false, "(true|false)")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(generateCmd)

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		log.Fatalf("failed to run shellpane: %v", err.Error())
	}

	return
}

func generate(fromSwagger2JSONPath string, outputPath string, setCategory string) error {
	b, err := os.ReadFile(fromSwagger2JSONPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read swagger2 file=%v", fromSwagger2JSONPath)
	}

	var swaggerFile convert.SwaggerFile
	err = json.Unmarshal(b, &swaggerFile)
	if err != nil {
		return errors.Wrapf(err, "failed to json unmarshal swagger file=%v content=%v", fromSwagger2JSONPath, string(b))
	}

	c, err := convert.ConvertSwaggerfileToShellpaneConfig(swaggerFile, setCategory)
	if err != nil {
		return errors.Wrapf(err, "failed convert swagger file to shellpane config")
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create shellpane config output file=%v", outputPath)
	}

	err = yaml.NewEncoder(f).Encode(c)
	if err != nil {
		return errors.Wrapf(err, "failed to yaml encode shellpane config")
	}

	return nil
}

func serve(ctx context.Context, config bootstrap.ContainerConfig) error {
	c := bootstrap.NewContainer(bootstrap.ContainerOpts{
		Config: config,
	})
	go func() {
		<-ctx.Done()
		c.Close(ctx)
	}()

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
}
