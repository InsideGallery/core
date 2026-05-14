package prometheus

import "testing"

func TestParseBuckets(t *testing.T) {
	got, err := parseBuckets("100,10,10,50")
	if err != nil {
		t.Fatalf("parseBuckets() error: %v", err)
	}

	want := []float64{10, 50, 100}
	if len(got) != len(want) {
		t.Fatalf("buckets = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("buckets = %v, want %v", got, want)
		}
	}
}

func TestParseBucketsRejectsInvalidValue(t *testing.T) {
	if _, err := parseBuckets("10,nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetConfigFromEnv(t *testing.T) {
	t.Setenv("METRICS_PROMETHEUS_CLASSIC_BUCKETS", "5,10")
	t.Setenv("METRICS_PROMETHEUS_NATIVE_BUCKET_FACTOR", "1.2")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	want := []float64{5, 10}
	if len(cfg.classicBuckets) != len(want) {
		t.Fatalf("classicBuckets = %v, want %v", cfg.classicBuckets, want)
	}

	for i := range want {
		if cfg.classicBuckets[i] != want[i] {
			t.Fatalf("classicBuckets = %v, want %v", cfg.classicBuckets, want)
		}
	}

	if cfg.NativeHistogramBucketFactor != 1.2 {
		t.Fatalf("NativeHistogramBucketFactor = %v", cfg.NativeHistogramBucketFactor)
	}
}

func TestGetConfigFromEnvUsesPrometheusOnlyDeploymentDefaults(t *testing.T) {
	t.Setenv("DD_STATSD_ADDR", "datadog:8125")
	t.Setenv("METRICS_DATADOG_ADDR", "datadog:8125")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if len(cfg.classicBuckets) != 0 {
		t.Fatalf("classicBuckets = %v, want empty", cfg.classicBuckets)
	}

	if cfg.NativeHistogramBucketFactor != 1.1 {
		t.Fatalf("NativeHistogramBucketFactor = %v, want 1.1", cfg.NativeHistogramBucketFactor)
	}
}

func TestGetConfigFromEnvRejectsInvalidNativeFactor(t *testing.T) {
	t.Setenv("METRICS_PROMETHEUS_NATIVE_BUCKET_FACTOR", "1")

	if _, err := getConfigFromEnv(); err == nil {
		t.Fatal("expected error")
	}
}
