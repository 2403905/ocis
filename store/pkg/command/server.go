package command

import (
	"context"

	"github.com/oklog/run"

	"github.com/owncloud/ocis/ocis-pkg/version"
	"github.com/owncloud/ocis/store/pkg/config"
	"github.com/owncloud/ocis/store/pkg/config/parser"
	"github.com/owncloud/ocis/store/pkg/logging"
	"github.com/owncloud/ocis/store/pkg/metrics"
	"github.com/owncloud/ocis/store/pkg/server/debug"
	"github.com/owncloud/ocis/store/pkg/server/grpc"
	"github.com/owncloud/ocis/store/pkg/tracing"
	"github.com/urfave/cli/v2"
)

// Server is the entrypoint for the server command.
func Server(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Start integrated server",
		Before: func(ctx *cli.Context) error {
			if err := parser.ParseConfig(cfg); err != nil {
				return err
			}

			return nil
		},
		Action: func(c *cli.Context) error {
			logger := logging.Configure(cfg.Service.Name, cfg.Log)
			err := tracing.Configure(cfg)
			if err != nil {
				return err
			}

			var (
				gr          = run.Group{}
				ctx, cancel = func() (context.Context, context.CancelFunc) {
					if cfg.Context == nil {
						return context.WithCancel(context.Background())
					}
					return context.WithCancel(cfg.Context)
				}()
				metrics = metrics.New()
			)

			defer cancel()

			metrics.BuildInfo.WithLabelValues(version.String).Set(1)

			{
				server := grpc.Server(
					grpc.Logger(logger),
					grpc.Context(ctx),
					grpc.Config(cfg),
					grpc.Metrics(metrics),
				)

				gr.Add(server.Run, func(_ error) {
					logger.Info().
						Str("server", "grpc").
						Msg("Shutting down server")

					cancel()
				})
			}

			{
				server, err := debug.Server(
					debug.Logger(logger),
					debug.Context(ctx),
					debug.Config(cfg),
				)

				if err != nil {
					logger.Error().Err(err).Str("server", "debug").Msg("Failed to initialize server")
					return err
				}

				gr.Add(server.ListenAndServe, func(_ error) {
					_ = server.Shutdown(ctx)
					cancel()
				})
			}

			return gr.Run()
		},
	}
}
