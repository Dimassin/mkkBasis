package mysql

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 25))
	db.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 10))
	db.SetConnMaxLifetime(envDurationMinutes("DB_CONN_MAX_LIFETIME_MIN", 30))
	db.SetConnMaxIdleTime(envDurationMinutes("DB_CONN_MAX_IDLE_MIN", 5))

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func envInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func envDurationMinutes(key string, defaultMinutes int) time.Duration {
	return time.Duration(envInt(key, defaultMinutes)) * time.Minute
}
