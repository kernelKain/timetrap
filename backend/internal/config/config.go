package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAppEnvironment         = "development"
	defaultPort                   = "8082"
	defaultAllowedOrigins         = "http://localhost:5173"
	defaultLogLevel               = "debug"
	defaultEngineVersion          = "dev"
	defaultDatabaseMaxConns int32 = 5
	defaultDatabaseMinConns int32 = 0
	defaultConnectTimeout         = 5 * time.Second
	defaultQueryTimeout           = 3 * time.Second
)

type Config struct {
	AppEnvironment string
	Port           string
	AllowedOrigins []string
	LogLevel       string
	EngineVersion  string

	DatabaseURL            string
	DatabaseMaxConns       int32
	DatabaseMinConns       int32
	DatabaseConnectTimeout time.Duration
	DatabaseQueryTimeout   time.Duration
}

func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL must be set")
	}

	databaseMaxConns, err := integerEnvironment(
		"DATABASE_MAX_CONNS",
		defaultDatabaseMaxConns,
	)
	if err != nil {
		return Config{}, err
	}

	if databaseMaxConns <= 0 {
		return Config{}, errors.New(
			"DATABASE_MAX_CONNS must be greater than zero",
		)
	}

	databaseMinConns, err := integerEnvironment(
		"DATABASE_MIN_CONNS",
		defaultDatabaseMinConns,
	)
	if err != nil {
		return Config{}, err
	}

	if databaseMinConns < 0 {
		return Config{}, errors.New(
			"DATABASE_MIN_CONNS must be zero or greater",
		)
	}

	if databaseMinConns > databaseMaxConns {
		return Config{}, errors.New(
			"DATABASE_MIN_CONNS must not exceed DATABASE_MAX_CONNS",
		)
	}

	connectTimeout, err := durationEnvironment(
		"DATABASE_CONNECT_TIMEOUT",
		defaultConnectTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	if connectTimeout <= 0 {
		return Config{}, errors.New(
			"DATABASE_CONNECT_TIMEOUT must be greater than zero",
		)
	}

	queryTimeout, err := durationEnvironment(
		"DATABASE_QUERY_TIMEOUT",
		defaultQueryTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	if queryTimeout <= 0 {
		return Config{}, errors.New(
			"DATABASE_QUERY_TIMEOUT must be greater than zero",
		)
	}

	allowedOrigins, err := originListEnvironment(
		"CORS_ALLOWED_ORIGINS",
		defaultAllowedOrigins,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnvironment: environment(
			"APP_ENV",
			defaultAppEnvironment,
		),
		Port: environment(
			"PORT",
			defaultPort,
		),
		AllowedOrigins: allowedOrigins,
		LogLevel: environment(
			"LOG_LEVEL",
			defaultLogLevel,
		),
		EngineVersion: environment(
			"ENGINE_VERSION",
			defaultEngineVersion,
		),
		DatabaseURL:            databaseURL,
		DatabaseMaxConns:       databaseMaxConns,
		DatabaseMinConns:       databaseMinConns,
		DatabaseConnectTimeout: connectTimeout,
		DatabaseQueryTimeout:   queryTimeout,
	}, nil
}

func environment(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}

func integerEnvironment(name string, fallback int32) (int32, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}

	return int32(parsed), nil
}

func durationEnvironment(
	name string,
	fallback time.Duration,
) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration", name)
	}

	return parsed, nil
}

func originListEnvironment(name, fallback string) ([]string, error) {
	value, configured := os.LookupEnv(name)
	if !configured {
		value = fallback
	}

	values := make([]string, 0)

	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("%s must not contain empty origins", name)
		}
		if item == "*" {
			return nil, fmt.Errorf("%s must contain exact origins, not *", name)
		}
		values = append(values, item)
	}

	return values, nil
}
