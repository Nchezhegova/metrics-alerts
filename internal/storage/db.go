package storage

import (
	"database/sql"
	"flag"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	_ "github.com/lib/pq"
	"os"
)

var DB *sql.DB

func init() {
	var err error
	var addr string

	flag.StringVar(&addr, "d", config.DATEBASE, "input addr db")
	if envDBaddr := os.Getenv("DATABASE_DSN"); envDBaddr != "" {
		addr = envDBaddr
	}

	//ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
	//	host, config.DBuser, config.DBpassword, config.DBname)

	DB, err = sql.Open("postgres", addr)
	if err != nil {
		panic(err)
	}

}

func CheckConnect(db *sql.DB) error {
	err := db.Ping()
	return err
}
