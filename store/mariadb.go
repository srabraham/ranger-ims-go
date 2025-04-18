package store

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/srabraham/ranger-ims-go/conf"
	"log"
)

func MariaDB() *sql.DB {
	// Capture connection properties.
	cfg := mysql.NewConfig()
	cfg.User = conf.Cfg.Store.MySQL.Username
	cfg.Passwd = conf.Cfg.Store.MySQL.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", conf.Cfg.Store.MySQL.HostName, conf.Cfg.Store.MySQL.HostPort)
	cfg.DBName = conf.Cfg.Store.MySQL.Database

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
	return db
}
