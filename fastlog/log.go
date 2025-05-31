package fastlog

import (
	"log"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
)

func SetupDefaultLog(m ...slogmulti.Middleware) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := cfg.GetHandler(m...)
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
