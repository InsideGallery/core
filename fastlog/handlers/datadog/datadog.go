package datadog

import (
	"context"
	"log/slog"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	slogdatadog "github.com/samber/slog-datadog/v2"

	"github.com/InsideGallery/core/fastlog/handlers"
)

// OutKind is the registry key for the datadog handler.
const OutKind = "datadog"

func init() {
	handlers.RegisterHandlerFunc(OutKind, newHandler)
}

func newHandler() (slog.Handler, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()

	ctx := context.Background()
	ctx = context.WithValue(ctx, datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: cfg.APIKey},
	})
	ctx = context.WithValue(ctx, datadog.ContextServerVariables, map[string]string{
		"site": cfg.Endpoint,
	})

	return slogdatadog.Option{
		Level:    cfg.Level,
		Client:   datadog.NewAPIClient(datadog.NewConfiguration()),
		Context:  ctx,
		Hostname: hostname,
		Service:  cfg.Service,
		Timeout:  cfg.Timeout,
	}.NewDatadogHandler(), nil
}
