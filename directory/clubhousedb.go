package directory

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
	return db
}
