package testconfig

import (
	"os"
	"time"
)

type Config struct {
	TestDatabaseURL     string
	TestTimeout        time.Duration
	ParallelTestDBs    bool
	CleanupTestData     bool
	VerboseLogging      bool
}

func Load() *Config {
	cfg := &Config{
		TestDatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
		TestTimeout:     getDuration("TEST_TIMEOUT", 10*time.Second),
		ParallelTestDBs: getBool("PARALLEL_TEST_DBS", false),
		CleanupTestData:  getBool("CLEANUP_TEST_DATA", true),
		VerboseLogging:   getBool("VERBOSE_LOGGING", false),
	}

	if cfg.TestDatabaseURL == "" {
		cfg.TestDatabaseURL = "postgres://fintrack:fintrack@localhost:5432/fintrack_test"
	}

	return cfg
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultValue
}

func IsCI() bool {
	return os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"
}

func IsVerbose() bool {
	return getBool("VERBOSE_LOGGING", false)
}
