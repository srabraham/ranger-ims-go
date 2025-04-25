package store

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/srabraham/ranger-ims-go/conf"
	"log"
	"log/slog"
	"strings"
	"time"
)

//go:embed schema.sql
var CurrentSchema string

func MariaDB(imsCfg *conf.IMSConfig) *sql.DB {
	// Capture connection properties.
	cfg := mysql.NewConfig()
	cfg.User = imsCfg.Store.MySQL.Username
	cfg.Passwd = imsCfg.Store.MySQL.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", imsCfg.Store.MySQL.HostName, imsCfg.Store.MySQL.HostPort)
	cfg.DBName = imsCfg.Store.MySQL.Database

	// Get a database handle.
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(20)
	pingErr := db.Ping()
	if pingErr != nil {
		log.Panic(pingErr)
	}
	slog.Info("Connected to IMS MariaDB")
	return db
}

type DB struct {
	*sql.DB
}

func (l DB) ExecContext(ctx context.Context, s string, i ...interface{}) (sql.Result, error) {
	start := time.Now()
	execContext, err := l.DB.ExecContext(ctx, s, i...)
	logQuery(s, start, err)
	return execContext, err
}

func (l DB) PrepareContext(ctx context.Context, s string) (*sql.Stmt, error) {
	start := time.Now()
	stmt, err := l.DB.PrepareContext(ctx, s)
	logQuery(s, start, err)
	return stmt, err
}

func (l DB) QueryContext(ctx context.Context, s string, i ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := l.DB.QueryContext(ctx, s, i...)
	logQuery(s, start, err)
	return rows, err
}

func (l DB) QueryRowContext(ctx context.Context, s string, i ...interface{}) *sql.Row {
	start := time.Now()
	row := l.DB.QueryRowContext(ctx, s, i...)
	logQuery(s, start, nil)
	return row
}

func logQuery(s string, start time.Time, err error) {
	queryName, _, _ := strings.Cut(s, "\n")
	queryName = strings.TrimPrefix(queryName, "-- name: ")
	// time.Since(start).Round(1*time.Microsecond)
	slog.Debug("Query", "time (ms)", time.Since(start).Milliseconds(), "query", queryName, "err", err)
}
