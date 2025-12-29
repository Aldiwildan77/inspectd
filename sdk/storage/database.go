package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Aldiwildan77/inspectd/sdk/types"
)

// DatabaseStorage stores snapshots in a SQL database.
// Supports PostgreSQL, MySQL, and other SQL databases via database/sql.
type DatabaseStorage struct {
	db     *sql.DB
	driver string
	dsn    string
}

// DatabaseStorageConfig configures database storage.
type DatabaseStorageConfig struct {
	// Driver is the database driver name (e.g., "postgres", "mysql").
	Driver string

	// DSN is the database connection string.
	DSN string

	// TableName is the table name for storing snapshots (default: "inspectd_snapshots").
	TableName string

	// MaxConnections is the maximum number of database connections (default: 10).
	MaxConnections int
}

// NewDatabaseStorage creates a new database storage instance.
// The table will be created automatically if it doesn't exist.
func NewDatabaseStorage(config DatabaseStorageConfig) (*DatabaseStorage, error) {
	if config.Driver == "" {
		return nil, fmt.Errorf("database driver is required")
	}
	if config.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}
	if config.TableName == "" {
		config.TableName = "inspectd_snapshots"
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = 10
	}

	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxConnections)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	storage := &DatabaseStorage{
		db:     db,
		driver: config.Driver,
		dsn:    config.DSN,
	}

	// Create table if it doesn't exist
	if err := storage.createTable(config.TableName); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return storage, nil
}

// createTable creates the snapshots table if it doesn't exist.
func (d *DatabaseStorage) createTable(tableName string) error {
	var createSQL string

	switch d.driver {
	case "postgres":
		createSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id SERIAL PRIMARY KEY,
				timestamp TIMESTAMP NOT NULL,
				data JSONB NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp);
		`, tableName, tableName, tableName)
	case "mysql":
		createSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				timestamp DATETIME NOT NULL,
				data JSON NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				INDEX idx_timestamp (timestamp)
			);
		`, tableName)
	default:
		// Generic SQL (may need adjustment for specific databases)
		createSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id INTEGER PRIMARY KEY AUTO_INCREMENT,
				timestamp TIMESTAMP NOT NULL,
				data TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp);
		`, tableName, tableName, tableName)
	}

	_, err := d.db.Exec(createSQL)
	return err
}

// Store saves a snapshot to the database.
func (d *DatabaseStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Parse timestamp
	timestamp, err := snapshot.ParseTimestamp()
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	// Marshal to JSON
	jsonData, err := snapshot.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Insert into database
	var query string
	switch d.driver {
	case "postgres":
		query = `INSERT INTO inspectd_snapshots (timestamp, data) VALUES ($1, $2::jsonb)`
	case "mysql":
		query = `INSERT INTO inspectd_snapshots (timestamp, data) VALUES (?, ?)`
	default:
		query = `INSERT INTO inspectd_snapshots (timestamp, data) VALUES (?, ?)`
	}

	_, err = d.db.ExecContext(ctx, query, timestamp, jsonData)
	if err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}

	return nil
}

// StoreBatch saves multiple snapshots in a transaction.
func (d *DatabaseStorage) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var query string
	switch d.driver {
	case "postgres":
		query = `INSERT INTO inspectd_snapshots (timestamp, data) VALUES ($1, $2::jsonb)`
	case "mysql":
		query = `INSERT INTO inspectd_snapshots (timestamp, data) VALUES (?, ?)`
	default:
		query = `INSERT INTO inspectd_snapshots (timestamp, data) VALUES (?, ?)`
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, snapshot := range snapshots {
		timestamp, err := snapshot.ParseTimestamp()
		if err != nil {
			continue // Skip invalid snapshots
		}

		jsonData, err := snapshot.ToJSON()
		if err != nil {
			continue // Skip invalid snapshots
		}

		_, err = stmt.ExecContext(ctx, timestamp, jsonData)
		if err != nil {
			return fmt.Errorf("failed to insert snapshot: %w", err)
		}
	}

	return tx.Commit()
}

// Query retrieves snapshots from the database.
func (d *DatabaseStorage) Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if opts == nil {
		opts = &QueryOptions{}
	}

	// Build query
	query := "SELECT data FROM inspectd_snapshots WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if opts.StartTime != nil {
		switch d.driver {
		case "postgres":
			query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		default:
			query += " AND timestamp >= ?"
		}
		args = append(args, *opts.StartTime)
		argIndex++
	}

	if opts.EndTime != nil {
		switch d.driver {
		case "postgres":
			query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		default:
			query += " AND timestamp <= ?"
		}
		args = append(args, *opts.EndTime)
		argIndex++
	}

	// Ordering
	if opts.OrderBy == OrderByTimeAsc {
		query += " ORDER BY timestamp ASC"
	} else {
		query += " ORDER BY timestamp DESC"
	}

	// Limit
	if opts.Limit > 0 {
		switch d.driver {
		case "postgres", "mysql":
			query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		default:
			query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		}
	}

	// Execute query
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshots: %w", err)
	}
	defer rows.Close()

	results := make([]*types.Snapshot, 0)

	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			continue // Skip invalid rows
		}

		snapshot, err := types.FromJSON(jsonData)
		if err != nil {
			continue // Skip invalid JSON
		}

		results = append(results, snapshot)
	}

	return results, rows.Err()
}

// Close closes the database connection.
func (d *DatabaseStorage) Close() error {
	return d.db.Close()
}

// Ping checks the database connection.
func (d *DatabaseStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return d.db.PingContext(ctx)
}
