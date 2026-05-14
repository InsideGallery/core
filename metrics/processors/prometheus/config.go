package prometheus

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

// Prometheus processor config uses METRICS_PROMETHEUS_* only for histogram tuning. Scraping is exposed by the profiler
// /metrics endpoint; Datadog and DogStatsD environment variables are intentionally not part of this config.
const envPrefix = "METRICS_PROMETHEUS"

type config struct {
	ClassicBuckets                  string        `env:"_CLASSIC_BUCKETS" envDefault:""`
	NativeHistogramBucketFactor     float64       `env:"_NATIVE_BUCKET_FACTOR" envDefault:"1.1"`
	NativeHistogramZeroThreshold    float64       `env:"_NATIVE_ZERO_THRESHOLD" envDefault:"0"`
	NativeHistogramMaxBucketNumber  uint32        `env:"_NATIVE_MAX_BUCKETS" envDefault:"160"`
	NativeHistogramMinResetDuration time.Duration `env:"_NATIVE_MIN_RESET_DURATION" envDefault:"1h"`
	NativeHistogramMaxZeroThreshold float64       `env:"_NATIVE_MAX_ZERO_THRESHOLD" envDefault:"0"`

	classicBuckets []float64
}

func getConfigFromEnv() (config, error) {
	var cfg config

	if err := env.ParseWithOptions(&cfg, env.Options{
		Prefix: envPrefix,
	}); err != nil {
		return config{}, err
	}

	if cfg.NativeHistogramBucketFactor <= 1 {
		return config{}, fmt.Errorf("native bucket factor must be greater than 1")
	}

	classicBuckets, err := parseBuckets(cfg.ClassicBuckets)
	if err != nil {
		return config{}, err
	}

	cfg.classicBuckets = classicBuckets

	return cfg, nil
}

func parseBuckets(raw string) ([]float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	buckets := make([]float64, 0, len(parts))

	for _, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("parse classic bucket %q: %w", part, err)
		}

		if value <= 0 || math.IsInf(value, 0) || math.IsNaN(value) {
			return nil, fmt.Errorf("classic bucket must be finite and positive: %q", part)
		}

		buckets = append(buckets, value)
	}

	sort.Float64s(buckets)

	return uniqueBuckets(buckets), nil
}

func uniqueBuckets(buckets []float64) []float64 {
	if len(buckets) == 0 {
		return nil
	}

	unique := buckets[:1]
	for _, bucket := range buckets[1:] {
		if bucket == unique[len(unique)-1] {
			continue
		}

		unique = append(unique, bucket)
	}

	return unique
}
