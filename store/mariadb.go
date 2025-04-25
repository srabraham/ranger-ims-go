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

	pingErr := db.Ping()
	if pingErr != nil {
		log.Panic(pingErr)
	}
	slog.Info("Connected to IMS MariaDB")
	return db
}

type TimedDBTX struct {
	*sql.DB
}

func (l TimedDBTX) ExecContext(ctx context.Context, s string, i ...interface{}) (sql.Result, error) {
	//start := time.Now()
	//defer func() {
	//	slog.Info("ExecContext complete", "s", s, "time", time.Since(start))
	//}()
	return l.DB.ExecContext(ctx, s, i...)
}

func (l TimedDBTX) PrepareContext(ctx context.Context, s string) (*sql.Stmt, error) {
	//start := time.Now()
	//defer func() {
	//	slog.Info("PrepareContext complete", "s", s, "time", time.Since(start))
	//}()
	return l.DB.PrepareContext(ctx, s)
}

func (l TimedDBTX) QueryContext(ctx context.Context, s string, i ...interface{}) (*sql.Rows, error) {
	//start := time.Now()
	//defer func() {
	//	slog.Info("QueryContext complete", "s", s, "time", time.Since(start))
	//}()
	return l.DB.QueryContext(ctx, s, i...)
}

func (l TimedDBTX) QueryRowContext(ctx context.Context, s string, i ...interface{}) *sql.Row {
	//start := time.Now()
	//defer func() {
	//	slog.Info("QueryRowContext complete", "s", s, "time", time.Since(start))
	//}()
	return l.DB.QueryRowContext(ctx, s, i...)
}
