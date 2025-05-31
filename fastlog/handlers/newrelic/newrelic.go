package newrelic

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/nrslog"
	"github.com/newrelic/go-agent/v3/newrelic"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "newrelic"

var (
	app *newrelic.Application
	mu  sync.Mutex
)

func init() {
	handlers.RegisterHandler(OutKind,
		Handler(context.Background()),
	)
}

func setApp(a *newrelic.Application) {
	mu.Lock()
	defer mu.Unlock()

	app = a
}

func getApp() *newrelic.Application {
	mu.Lock()
	defer mu.Unlock()

	return app
}

func Handler(_ context.Context) slog.Handler {
	application, err := newrelic.NewApplication(
		newrelic.ConfigFromEnvironment(),
	)
	if err != nil {
		slog.Default().Error("Error get newrelic app", "err", err)
		return nil
	}

	setApp(application)

	instrumentedTextHandler := nrslog.JSONHandler(app, os.Stderr, &slog.HandlerOptions{})

	return instrumentedTextHandler
}

func GetTraceContext(
	ctx context.Context, name string,
) (context.Context, *newrelic.Transaction) {
	txn := getApp().StartTransaction(name)

	return newrelic.NewContext(ctx, txn), txn
}

func TracerEnd(txn *newrelic.Transaction) {
	txn.End()
}
