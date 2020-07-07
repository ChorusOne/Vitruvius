package types

import "os"

// Config is used to wrap up runtime choices to pass around the app.
type Config struct {
	DailySnapshots    bool   // Take snapshot at the first observed timestamp of the day.
	DatabasePath      string // Note: Postgres expected.
	ListenPort        string // Default port is 10100 if none provided.
	OasisSocket       string // UNIX Socket for Oasis gRPC
	SnapshotFrequency int    // 0 means never.
}

// ConfigFromEnv produces a config option from the currently available
// environment variables. Hippias does not yet support other configuration
// methods like files.
func ConfigFromEnv() Config {
	defaultEnv := func(key, def string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		return def
	}

	return Config{
		DailySnapshots:    true,
		DatabasePath:      defaultEnv("HIPPIAS_DB", ""),
		ListenPort:        defaultEnv("HIPPIAS_PORT", "10100"),
		OasisSocket:       defaultEnv("HIPPIAS_SOCKET", "./internal.sock"),
		SnapshotFrequency: 0,
	}
}
