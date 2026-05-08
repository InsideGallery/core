package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/InsideGallery/core/fastlog"
)

type loggerRuntime struct {
	handle       *fastlog.DefaultLoggerHandle
	closeDefault func() error
}

func newLoggerRuntime(ctx context.Context, options LoggerOptions) (*loggerRuntime, error) {
	if options.Logger == nil && options.Config == nil {
		return &loggerRuntime{}, nil
	}

	logger := options.Logger

	if logger == nil && options.Config != nil {
		if options.InstallDefault && options.HandlerRegistry == nil {
			closeDefault, err := fastlog.SetupDefault(ctx, options.Config)
			if err != nil {
				return nil, fmt.Errorf("log setup: %w", err)
			}

			return &loggerRuntime{
				closeDefault: closeDefault,
			}, nil
		}

		var err error

		logger, err = fastlog.NewLoggerWithRegistry(options.Config, options.HandlerRegistry)
		if err != nil {
			return nil, fmt.Errorf("log setup: %w", err)
		}
	}

	runtime := &loggerRuntime{}

	if options.InstallDefault {
		runtime.handle = fastlog.InstallDefaultLogger(logger)
	}

	return runtime, nil
}

func (r *loggerRuntime) Close() error {
	if r == nil || r.handle == nil {
		if r == nil || r.closeDefault == nil {
			return nil
		}

		return r.closeDefault()
	}

	return errors.Join(r.handle.Close(), closeDefaultLogger(r.closeDefault))
}

func closeDefaultLogger(closeDefault func() error) error {
	if closeDefault == nil {
		return nil
	}

	return closeDefault()
}
