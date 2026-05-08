package app

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/queue/nats/middleware"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

const defaultWebServiceName = "web"

// Config is the root app bootstrap configuration.
type Config struct {
	ServiceName       string
	ServerName        string
	NATSServiceName   string
	Port              string
	Log               *fastlog.Config
	Metrics           *metrics.Config
	NATS              *client.Config
	MonitorAddr       string
	ShutdownTimeout   time.Duration
	EnablePrintRoutes bool
	ProfilerState     *profiler.State
	ShutdownListener  *oslistener.SignalListener
	InitRoutes        InitRoutes
	NATSClient        *client.Client
	Subscriber        *subscriber.Subscriber
	Middleware        *middleware.Middleware
	CloseNATSClient   bool

	legacyWebOptions    *WebOptions
	legacyNATSOptions   *NATSOptions
	legacyLoggerOptions *LoggerOptions
}

type environmentConfig struct {
	ServiceName           string        `env:"SERVICE_NAME"`
	ServerName            string        `env:"SERVER_NAME"`
	NATSServiceName       string        `env:"NATS_SERVICE_NAME"`
	Port                  string        `env:"PORT" envDefault:":8080"`
	MonitorAddr           string        `env:"MONITOR_ADDR"`
	ShutdownTimeout       time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	DeploymentEnvironment string        `env:"DEPLOYMENT_ENVIRONMENT"`
}

// ConfigFromEnv reads app, logger, metrics, and NATS configuration from environment variables.
func ConfigFromEnv() (*Config, error) {
	cfg, err := configFromEnv(true)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func configFromEnv(includeNATS bool) (*Config, error) {
	environment, err := readEnvironmentConfig()
	if err != nil {
		return nil, fmt.Errorf("app config: %w", err)
	}

	logConfig, err := fastlog.GetConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("log config: %w", err)
	}

	metricsConfig, err := metrics.GetEnvConfig()
	if err != nil {
		return nil, fmt.Errorf("metrics config: %w", err)
	}

	var natsConfig *client.Config
	if includeNATS {
		natsConfig, err = client.GetNATSConnectionConfigFromEnv()
		if err != nil {
			return nil, fmt.Errorf("nats config: %w", err)
		}
	}

	cfg := configFromEnvironment(environment, logConfig, &metricsConfig, natsConfig)

	return cfg, nil
}

func readEnvironmentConfig() (environmentConfig, error) {
	var cfg environmentConfig

	if err := env.Parse(&cfg); err != nil {
		return environmentConfig{}, err
	}

	return cfg, nil
}

func configFromEnvironment(
	environment environmentConfig,
	logConfig *fastlog.Config,
	metricsConfig *metrics.Config,
	natsConfig *client.Config,
) *Config {
	return &Config{
		ServiceName:       environment.ServiceName,
		ServerName:        environment.ServerName,
		NATSServiceName:   environment.NATSServiceName,
		Port:              environment.Port,
		Log:               logConfig,
		Metrics:           metricsConfig,
		NATS:              natsConfig,
		MonitorAddr:       environment.MonitorAddr,
		ShutdownTimeout:   environment.ShutdownTimeout,
		EnablePrintRoutes: environment.DeploymentEnvironment == "",
		ShutdownListener:  oslistener.DefaultListener(),
		CloseNATSClient:   true,
	}
}

func webConfigFromEnv(port string, serverName string) (*Config, error) {
	cfg, err := configFromEnv(false)
	if err != nil {
		return nil, err
	}

	if port != "" {
		cfg.Port = port
	}

	if serverName != "" {
		cfg.ServerName = serverName
	}

	return cfg, nil
}

func configFromWebOptions(options WebOptions) *Config {
	optionsCopy := options
	loggerOptions := options.Logger

	return &Config{
		ServiceName:         options.ServerName,
		ServerName:          options.ServerName,
		Port:                options.Port,
		MonitorAddr:         options.MonitorAddr,
		ShutdownTimeout:     options.ShutdownTimeout,
		EnablePrintRoutes:   options.Environment.EnablePrintRoutes,
		ProfilerState:       options.ProfilerState,
		ShutdownListener:    options.ShutdownListener,
		InitRoutes:          options.InitRoutes,
		legacyWebOptions:    &optionsCopy,
		legacyLoggerOptions: &loggerOptions,
	}
}

func configFromNATSOptions(options NATSOptions) *Config {
	optionsCopy := options
	loggerOptions := options.Logger

	return &Config{
		ServiceName:         options.ServiceName,
		NATSServiceName:     options.ServiceName,
		Log:                 loggerOptions.Config,
		NATS:                options.NATSConfig,
		MonitorAddr:         options.MonitorAddr,
		ProfilerState:       options.ProfilerState,
		ShutdownListener:    options.ShutdownListener,
		NATSClient:          options.NATSClient,
		Subscriber:          options.Subscriber,
		Middleware:          options.Middleware,
		CloseNATSClient:     options.CloseNATSClient,
		legacyNATSOptions:   &optionsCopy,
		legacyLoggerOptions: &loggerOptions,
	}
}

func webServiceName(cfg *Config) string {
	if cfg.ServerName != "" {
		return cfg.ServerName
	}

	if cfg.ServiceName != "" {
		return cfg.ServiceName
	}

	return defaultWebServiceName
}

func natsServiceNameFromConfig(cfg *Config) string {
	if cfg.NATSServiceName != "" {
		return cfg.NATSServiceName
	}

	if cfg.ServiceName != "" {
		return cfg.ServiceName
	}

	return defaultNATSServiceName
}

func configMetricsValue(cfg *Config) metrics.Config {
	if cfg.Metrics == nil {
		return metrics.Config{}
	}

	return *cfg.Metrics
}
