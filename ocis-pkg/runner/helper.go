package runner

import (
	"context"

	"github.com/owncloud/ocis/v2/ocis-pkg/log"
)

// GroupRunnerWithLogging runs the provided GroupRunner and logs the results with the provided logger.
func GroupRunnerWithLogging(ctx context.Context, gr *GroupRunner, name string, logger log.Logger) error {
	grResults := gr.Run(ctx)
	logger.Warn().Msgf("the %s service is shuting down", name)
	// return the first non-nil error found in the results
	for _, grResult := range grResults {
		if grResult.RunnerError != nil {
			logger.Error().Err(grResult.RunnerError).Msgf("the %s service has stopped with error", name)
			return grResult.RunnerError
		}
	}
	logger.Warn().Msgf("the %s service is stopped without error", name)
	return nil
}
