package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConnLifetime    = 30 * time.Minute
	maxConnIdleTime    = 5 * time.Minute
	healthCheckPeriod  = 1 * time.Minute
	startupPingTimeout = 5 * time.Second
)

// Config contains only the settings needed by the PostgreSQL store.
type Config struct {
	DatabaseURL    string
	MaxConns       int32
	MinConns       int32
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration
}

// Store implements TimeTrap's repositories using PostgreSQL.
type Store struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	newUUID      func() uuid.UUID
}

// Open configures the connection pool and verifies database connectivity.
func Open(ctx context.Context, cfg Config) (*Store, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		// Do not wrap the parsing error because it could contain URL material.
		return nil, errors.New("parse PostgreSQL configuration")
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = maxConnLifetime
	poolConfig.MaxConnIdleTime = maxConnIdleTime
	poolConfig.HealthCheckPeriod = healthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL connection pool: %w", err)
	}

	startupContext, cancel := context.WithTimeout(
		ctx,
		startupPingTimeout,
	)
	defer cancel()

	if err := pool.Ping(startupContext); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return &Store{
		pool:         pool,
		queryTimeout: cfg.QueryTimeout,
		newUUID:      uuid.New,
	}, nil
}

// Close releases all connections held by the pool.
func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}

	s.pool.Close()
}

// Ping verifies that the PostgreSQL pool can reach the database.
//
// The caller controls the deadline through ctx.
func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.pool == nil {
		return errors.New("PostgreSQL connection pool is not open")
	}

	return s.pool.Ping(ctx)
}

func validateConfig(cfg Config) error {
	switch {
	case strings.TrimSpace(cfg.DatabaseURL) == "":
		return errors.New("PostgreSQL database URL must be set")
	case cfg.MaxConns <= 0:
		return errors.New("PostgreSQL maximum connections must be positive")
	case cfg.MinConns < 0:
		return errors.New("PostgreSQL minimum connections cannot be negative")
	case cfg.MinConns > cfg.MaxConns:
		return errors.New(
			"PostgreSQL minimum connections cannot exceed maximum connections",
		)
	case cfg.ConnectTimeout <= 0:
		return errors.New("PostgreSQL connection timeout must be positive")
	case cfg.QueryTimeout <= 0:
		return errors.New("PostgreSQL query timeout must be positive")
	default:
		return nil
	}
}
