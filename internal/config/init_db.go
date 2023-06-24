package config

import (
	"context"
	"github.com/jackc/pgx"
	"log"

	"time"
)

const (
	secondsToWait = 90

	dsn = "user=postgres dbname=forumservice password=12345 host=localhost port=5432 sslmode=disable pool_max_conns=20"
)

func InitPostgres() (*pgx.ConnPool, error) {

	//initConn, err := pgx.Connect(PostgresConf)

	end := time.Now().Add(time.Second * secondsToWait)

	for time.Now().Before(end) {
		log.Println("Trying to open pg connection")
		initConn, err := pgx.Connect(PostgresConf)
		if err != nil {
			log.Println(err)
			continue
		}

		err = initConn.Ping(context.Background())
		if err == nil {
			log.Println("Ping sucessful")
			break
		}
		//
		time.Sleep(time.Second)
	}

	db, err := pgx.NewConnPool(PostgresPoolCong)
	if err != nil {
		return nil, err
	}

	return db, nil
}
