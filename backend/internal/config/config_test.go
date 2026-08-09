package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	clearConfigurationEnvironment(t)
	t.Setenv(
		"DATABASE_URL",
		"postgres://example:example@localhost:5432/example",
	)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	if cfg.AppEnvironment != "development" {
		t.Errorf(
			"AppEnvironment = %q; want development",
			cfg.AppEnvironment,
		)
	}

	if cfg.Port != "8082" {
		t.Errorf("Port = %q; want 8082", cfg.Port)
	}

	wantOrigins := []string{"http://localhost:5173"}
	if !reflect.DeepEqual(cfg.AllowedOrigins, wantOrigins) {
		t.Errorf(
			"AllowedOrigins = %#v; want %#v",
			cfg.AllowedOrigins,
			wantOrigins,
		)
	}

	if cfg.DatabaseMaxConns != 5 {
		t.Errorf(
			"DatabaseMaxConns = %d; want 5",
			cfg.DatabaseMaxConns,
		)
	}

	if cfg.DatabaseMinConns != 0 {
		t.Errorf(
			"DatabaseMinConns = %d; want 0",
			cfg.DatabaseMinConns,
		)
	}

	if cfg.DatabaseConnectTimeout != 5*time.Second {
		t.Errorf(
			"DatabaseConnectTimeout = %s; want 5s",
			cfg.DatabaseConnectTimeout,
		)
	}

	if cfg.DatabaseQueryTimeout != 3*time.Second {
		t.Errorf(
			"DatabaseQueryTimeout = %s; want 3s",
			cfg.DatabaseQueryTimeout,
		)
	}
}

func TestLoadAcceptsConfiguredValues(t *testing.T) {
	clearConfigurationEnvironment(t)

	t.Setenv(
		"DATABASE_URL",
		"postgres://example:example@localhost:5432/example",
	)
	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "9090")
	t.Setenv(
		"ALLOWED_ORIGINS",
		"http://localhost:5173, https://example.test",
	)
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("ENGINE_VERSION", "phase-4")
	t.Setenv("DATABASE_MAX_CONNS", "8")
	t.Setenv("DATABASE_MIN_CONNS", "2")
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "7s")
	t.Setenv("DATABASE_QUERY_TIMEOUT", "1500ms")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}

	if cfg.AppEnvironment != "test" {
		t.Errorf(
			"AppEnvironment = %q; want test",
			cfg.AppEnvironment,
		)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q; want 9090", cfg.Port)
	}

	wantOrigins := []string{
		"http://localhost:5173",
		"https://example.test",
	}
	if !reflect.DeepEqual(cfg.AllowedOrigins, wantOrigins) {
		t.Errorf(
			"AllowedOrigins = %#v; want %#v",
			cfg.AllowedOrigins,
			wantOrigins,
		)
	}

	if cfg.DatabaseMaxConns != 8 {
		t.Errorf(
			"DatabaseMaxConns = %d; want 8",
			cfg.DatabaseMaxConns,
		)
	}

	if cfg.DatabaseMinConns != 2 {
		t.Errorf(
			"DatabaseMinConns = %d; want 2",
			cfg.DatabaseMinConns,
		)
	}

	if cfg.DatabaseConnectTimeout != 7*time.Second {
		t.Errorf(
			"DatabaseConnectTimeout = %s; want 7s",
			cfg.DatabaseConnectTimeout,
		)
	}

	if cfg.DatabaseQueryTimeout != 1500*time.Millisecond {
		t.Errorf(
			"DatabaseQueryTimeout = %s; want 1500ms",
			cfg.DatabaseQueryTimeout,
		)
	}
}

func TestLoadRejectsInvalidDatabaseConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		variable    string
		value       string
		wantMessage string
	}{
		{
			name:        "missing database URL",
			variable:    "DATABASE_URL",
			value:       "",
			wantMessage: "DATABASE_URL must be set",
		},
		{
			name:        "non-integer maximum connections",
			variable:    "DATABASE_MAX_CONNS",
			value:       "many",
			wantMessage: "DATABASE_MAX_CONNS must be an integer",
		},
		{
			name:        "zero maximum connections",
			variable:    "DATABASE_MAX_CONNS",
			value:       "0",
			wantMessage: "DATABASE_MAX_CONNS must be greater than zero",
		},
		{
			name:        "negative minimum connections",
			variable:    "DATABASE_MIN_CONNS",
			value:       "-1",
			wantMessage: "DATABASE_MIN_CONNS must be zero or greater",
		},
		{
			name:        "minimum exceeds maximum",
			variable:    "DATABASE_MIN_CONNS",
			value:       "6",
			wantMessage: "DATABASE_MIN_CONNS must not exceed DATABASE_MAX_CONNS",
		},
		{
			name:        "invalid connection timeout",
			variable:    "DATABASE_CONNECT_TIMEOUT",
			value:       "soon",
			wantMessage: "DATABASE_CONNECT_TIMEOUT must be a valid duration",
		},
		{
			name:        "zero connection timeout",
			variable:    "DATABASE_CONNECT_TIMEOUT",
			value:       "0s",
			wantMessage: "DATABASE_CONNECT_TIMEOUT must be greater than zero",
		},
		{
			name:        "invalid query timeout",
			variable:    "DATABASE_QUERY_TIMEOUT",
			value:       "later",
			wantMessage: "DATABASE_QUERY_TIMEOUT must be a valid duration",
		},
		{
			name:        "negative query timeout",
			variable:    "DATABASE_QUERY_TIMEOUT",
			value:       "-1s",
			wantMessage: "DATABASE_QUERY_TIMEOUT must be greater than zero",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearConfigurationEnvironment(t)

			t.Setenv(
				"DATABASE_URL",
				"postgres://example:example@localhost:5432/example",
			)
			t.Setenv(test.variable, test.value)

			_, err := Load()
			if err == nil {
				t.Fatal("Load() returned nil error; want an error")
			}

			if !strings.Contains(err.Error(), test.wantMessage) {
				t.Errorf(
					"Load() error = %q; want it to contain %q",
					err.Error(),
					test.wantMessage,
				)
			}
		})
	}
}

func clearConfigurationEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		"APP_ENV",
		"PORT",
		"ALLOWED_ORIGINS",
		"LOG_LEVEL",
		"ENGINE_VERSION",
		"DATABASE_URL",
		"DATABASE_MAX_CONNS",
		"DATABASE_MIN_CONNS",
		"DATABASE_CONNECT_TIMEOUT",
		"DATABASE_QUERY_TIMEOUT",
	} {
		t.Setenv(name, "")
	}
}
