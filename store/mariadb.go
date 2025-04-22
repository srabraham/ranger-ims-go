package store

import (
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
