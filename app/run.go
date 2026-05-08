// Package app provides optional bootstrap helpers for consumer-owned web and NATS processes.
//
// This package is still library code: consumers own main(), ports, routes,
// subjects, config, shutdown, and deployment.
//
// New code should use error-returning helpers with explicit option structs:
//
//	import "github.com/InsideGallery/core/app"
//
//	err := app.RunWeb(ctx, app.WebOptions{ServerName: "api", InitRoutes: routes})
//	err = app.RunNATS(ctx, app.NATSOptions{ServiceName: "worker", InitSubscriptions: subscriptions})
//
// Prefer RunWeb, RunNATS, WebOptions, NATSOptions, LoggerOptions, and
// MetricsClientOptions so applications can handle setup/runtime errors and own
// logger, metrics, profiler, signal, route, and subscription dependencies.
//
// Compatibility: WebMain, WebMainWithOptions, InitRouter, and NATSMain remain
// available for existing main-style wiring that logs returned errors instead of
// returning them to the caller.
package app

import (
	"context"
	"errors"
	"fmt"
)

// RunWeb starts the web bootstrap flow and returns setup or runtime errors to the caller.
func RunWeb(ctx context.Context, config any, initRouter ...InitRouter) error {
	switch value := config.(type) {
	case nil:
		return runWebConfig(ctx, nil, firstInitRouter(initRouter))
	case *Config:
		return runWebConfig(ctx, value, firstInitRouter(initRouter))
	case Config:
		return runWebConfig(ctx, &value, firstInitRouter(initRouter))
	case WebOptions:
		return runWebConfig(ctx, configFromWebOptions(value), firstInitRouter(initRouter))
	case *WebOptions:
		if value == nil {
			return fmt.Errorf("web options are not set")
		}

		return runWebConfig(ctx, configFromWebOptions(*value), firstInitRouter(initRouter))
	default:
		return fmt.Errorf("unsupported web config %T", config)
	}
}

// RunNATS starts the NATS bootstrap flow and returns setup or runtime errors to the caller.
func RunNATS(ctx context.Context, config any, initSubscriptions ...InitSubscriptions) error {
	switch value := config.(type) {
	case nil:
		return runNATSConfig(ctx, nil, firstInitSubscriptions(initSubscriptions))
	case *Config:
		return runNATSConfig(ctx, value, firstInitSubscriptions(initSubscriptions))
	case Config:
		return runNATSConfig(ctx, &value, firstInitSubscriptions(initSubscriptions))
	case NATSOptions:
		return runNATSConfig(ctx, configFromNATSOptions(value), firstInitSubscriptions(initSubscriptions))
	case *NATSOptions:
		if value == nil {
			return fmt.Errorf("nats options are not set")
		}

		return runNATSConfig(ctx, configFromNATSOptions(*value), firstInitSubscriptions(initSubscriptions))
	default:
		return fmt.Errorf("unsupported nats config %T", config)
	}
}

func runWebConfig(ctx context.Context, cfg *Config, initRouter InitRouter) (runErr error) {
	if cfg == nil {
		var err error

		cfg, err = webConfigFromEnv("", "")
		if err != nil {
			return err
		}
	}

	loggerRuntime, err := newLoggerRuntime(ctx, loggerOptionsFromConfig(cfg))
	if err != nil {
		return err
	}

	defer func() {
		runErr = errors.Join(runErr, loggerRuntime.Close())
	}()

	options := webOptionsFromConfig(cfg, initRouter)

	return runWebOptions(ctx, options)
}

func runNATSConfig(ctx context.Context, cfg *Config, initSubscriptions InitSubscriptions) (runErr error) {
	if cfg == nil {
		var err error

		cfg, err = ConfigFromEnv()
		if err != nil {
			return err
		}
	}

	loggerRuntime, err := newLoggerRuntime(ctx, loggerOptionsFromConfig(cfg))
	if err != nil {
		return err
	}

	defer func() {
		runErr = errors.Join(runErr, loggerRuntime.Close())
	}()

	options := natsOptionsFromConfig(cfg, initSubscriptions)

	return runNATSOptions(ctx, options)
}

func webOptionsFromConfig(cfg *Config, initRouter InitRouter) WebOptions {
	if cfg.legacyWebOptions != nil {
		options := *cfg.legacyWebOptions
		if initRouter != nil {
			options.InitRouter = initRouter
		}

		return options
	}

	serviceName := webServiceName(cfg)

	return WebOptions{
		Port:       cfg.Port,
		ServerName: serviceName,
		Metrics: MetricsClientOptions{
			Config:              configMetricsValue(cfg),
			ServiceName:         serviceName,
			HealthState:         cfg.ProfilerState,
			InstallDefault:      true,
			RegisterHealthCheck: true,
		},
		ProfilerState:    cfg.ProfilerState,
		MonitorAddr:      cfg.MonitorAddr,
		ShutdownListener: cfg.ShutdownListener,
		ShutdownTimeout:  cfg.ShutdownTimeout,
		Environment: RuntimeEnvironmentOptions{
			EnablePrintRoutes: cfg.EnablePrintRoutes,
		},
		InitRoutes: cfg.InitRoutes,
		InitRouter: initRouter,
	}
}

func natsOptionsFromConfig(cfg *Config, initSubscriptions InitSubscriptions) NATSOptions {
	if cfg.legacyNATSOptions != nil {
		options := *cfg.legacyNATSOptions
		if initSubscriptions != nil {
			options.InitSubscriptions = initSubscriptions
		}

		return options
	}

	serviceName := natsServiceNameFromConfig(cfg)

	return NATSOptions{
		ServiceName: serviceName,
		Metrics: MetricsClientOptions{
			Config:              configMetricsValue(cfg),
			ServiceName:         serviceName,
			HealthState:         cfg.ProfilerState,
			InstallDefault:      true,
			RegisterHealthCheck: true,
		},
		ProfilerState:     cfg.ProfilerState,
		MonitorAddr:       cfg.MonitorAddr,
		ShutdownListener:  cfg.ShutdownListener,
		NATSConfig:        cfg.NATS,
		NATSClient:        cfg.NATSClient,
		Subscriber:        cfg.Subscriber,
		Middleware:        cfg.Middleware,
		InitSubscriptions: initSubscriptions,
		CloseNATSClient:   cfg.CloseNATSClient,
	}
}

func firstInitRouter(initRouters []InitRouter) InitRouter {
	if len(initRouters) == 0 {
		return nil
	}

	return initRouters[0]
}

func firstInitSubscriptions(initSubscriptions []InitSubscriptions) InitSubscriptions {
	if len(initSubscriptions) == 0 {
		return nil
	}

	return initSubscriptions[0]
}

func loggerOptionsFromConfig(cfg *Config) LoggerOptions {
	if cfg.legacyLoggerOptions != nil {
		options := *cfg.legacyLoggerOptions
		if options.Logger != nil || options.Config != nil {
			options.InstallDefault = true
		}

		return options
	}

	return LoggerOptions{
		Config:         cfg.Log,
		InstallDefault: true,
	}
}
