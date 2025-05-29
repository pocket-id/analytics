package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type InstancesStats struct {
	Total   int                `json:"total"`
	History []InstancesHistory `json:"history"`
}

type InstancesHistory struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func UpsertInstance(parentCtx context.Context, db *sql.DB, instanceID, version string) error {
	now := time.Now()

	// Upsert the instance
	const query = `
	INSERT INTO instances (id, first_seen, last_seen, latest_version)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		last_seen = excluded.last_seen,
		latest_version = excluded.latest_version
	`

	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()
	_, err := db.ExecContext(
		ctx,
		query,
		instanceID, now, now, version,
	)

	return err
}

func GetTotalInstances(parentCtx context.Context, db *sql.DB) (int, error) {
	// Only count instances that:
	// 1. Are older than 1 day
	// 2. Have been active in the last 2 days
	const query = `
	SELECT COUNT(*) 
	FROM instances 
	WHERE first_seen < datetime('now', '-1 day') 
	AND last_seen >= datetime('now', '-2 days')
	`

	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()
	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func GetInstancesOverTime(parentCtx context.Context, db *sql.DB, timeframe string) ([]InstancesHistory, error) {
	var query string

	switch timeframe {
	case "daily":
		// Get daily instance counts for the last 30 days
		// Only include instances that are older than 1 day and were active in the last 2 days
		query = `
		SELECT 
			DATE(first_seen) as date,
			COUNT(*) as daily_new,
			(SELECT COUNT(*) 
			 FROM instances i2 
			 WHERE DATE(i2.first_seen) <= DATE(i1.first_seen)
			 AND i2.first_seen < datetime('now', '-1 day')
			 AND i2.last_seen >= datetime('now', '-2 days')) as cumulative_count
		FROM instances i1
		WHERE first_seen >= datetime('now', '-30 days')
		AND first_seen < datetime('now', '-1 day')
		AND last_seen >= datetime('now', '-2 days')
		GROUP BY DATE(first_seen)
		ORDER BY date
		`
	case "monthly":
		// Get monthly instance counts for all time
		// Only include instances that are older than 1 day and were active in the last 2 days
		query = `
		SELECT 
			strftime('%Y-%m', first_seen) as date,
			COUNT(*) as monthly_new,
			(SELECT COUNT(*) 
			 FROM instances i2 
			 WHERE strftime('%Y-%m', i2.first_seen) <= strftime('%Y-%m', i1.first_seen)
			 AND i2.first_seen < datetime('now', '-1 day')
			 AND i2.last_seen >= datetime('now', '-2 days')) as cumulative_count
		FROM instances i1
		WHERE first_seen < datetime('now', '-1 day')
		AND last_seen >= datetime('now', '-2 days')
		GROUP BY strftime('%Y-%m', first_seen)
		ORDER BY date
		`
	default:
		return nil, fmt.Errorf("invalid timeframe: %s. Use 'daily' or 'monthly'", timeframe)
	}

	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chartData := make([]InstancesHistory, 0, 36)
	for rows.Next() {
		var date string
		var newCount, cumulativeCount int

		err := rows.Scan(&date, &newCount, &cumulativeCount)
		if err != nil {
			return nil, err
		}

		chartData = append(chartData, InstancesHistory{
			Date:  date,
			Count: cumulativeCount,
		})
	}

	return chartData, nil
}

func initDB() (*sql.DB, error) {
	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", "./data/pocket-id-analytics.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_txlock=immediate")
	if err != nil {
		return nil, err
	}

	// Create instances table
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS instances (
        id TEXT PRIMARY KEY,
        first_seen DATETIME NOT NULL,
        last_seen DATETIME NOT NULL,
        latest_version TEXT NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_first_seen ON instances(first_seen);
    CREATE INDEX IF NOT EXISTS idx_last_seen ON instances(last_seen);
    `

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
