package directory

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/srabraham/ranger-ims-go/conf"
	"log"
	"log/slog"
	"strings"
	"time"
)

func MariaDB() *DB {
	// Capture connection properties.
	cfg := mysql.NewConfig()
	cfg.User = conf.Cfg.Directory.ClubhouseDB.Username
	cfg.Passwd = conf.Cfg.Directory.ClubhouseDB.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", conf.Cfg.Directory.ClubhouseDB.Hostname, conf.Cfg.Directory.ClubhouseDB.HostPort)
	cfg.DBName = conf.Cfg.Directory.ClubhouseDB.Database

	// Get a database handle.
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	return &DB{DB: db}
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
	slog.Debug("Clubhouse Query", "millis", time.Since(start).Milliseconds(), "query", queryName, "err", err)
}
