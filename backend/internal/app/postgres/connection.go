package postgres

import (
	"context"
	"fmt"
	"time"

	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Database holds the database connection
type Database struct {
	DB     *gorm.DB
	logger *logger.Logger
}

// New creates a new database connection
func New(cfg config.DatabaseConfig, log *logger.Logger) (*Database, error) {
	// Configure GORM logger
	logLevel := gormlogger.Silent
	// if cfg.DebugLevel != " " {
	// 	logLevel = gormlogger.Info
	// }
	if cfg.SSLMode == "disable" { // Use debug mode in development
		logLevel = gormlogger.Info
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Kích hoạt tính năng "Prepared Statement" cache của GORM.
		PrepareStmt: true,
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configure connection pool
	// Tối ưu Connection Pool (Formula: n_cores * 2 + effective_spindle_count)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	// Tránh giữ connect quá lâu khi idle, giúp load balancer (nếu có) phân phối lại
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// Ping database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database connected successfully")

	return &Database{DB: db, logger: log}, nil
}

// AutoMigrate is removed in favor of Goose migrations
func (d *Database) AutoMigrate() error {
	d.logger.Warn("AutoMigrate is deprecated and disabled. Use 'make migrate-up' instead.")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks database health
func (d *Database) Health(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

// WithContext returns a new DB with context
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.DB.WithContext(ctx)
}
