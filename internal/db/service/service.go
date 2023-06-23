package service

import (
	"DBProject/internal/common"
	"database/sql"
	"github.com/jackc/pgx"
	"log"
)

type ServiceController struct {
	db *pgx.ConnPool
}

func New(pool *pgx.ConnPool) *ServiceController {
	return &ServiceController{db: pool}
}

func (handler *ServiceController) Clear() error {
	_, err := handler.db.Exec(`TRUNCATE users cascade`)
	if err != nil {
		log.Println(err)
	}

	_, err = handler.db.Exec(`TRUNCATE Forum cascade`)
	if err != nil {
		log.Println(err)
	}

	_, err = handler.db.Exec(`TRUNCATE Thread cascade`)
	if err != nil {
		log.Println(err)
	}

	_, err = handler.db.Exec(`TRUNCATE Post cascade`)
	if err != nil {
		log.Println(err)
	}

	_, err = handler.db.Exec(`TRUNCATE Vote cascade`)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (handler *ServiceController) Status() (*common.DbStatus, error) {
	statusInfo := &common.DbStatus{}

	temp := sql.NullInt64{}

	err := handler.db.QueryRow(`SELECT count(*) from users`).Scan(&temp)
	if err != nil {
		return nil, err
	}

	statusInfo.User = int(temp.Int64)

	err = handler.db.QueryRow(`SELECT count(*) from forum`).Scan(&temp)
	if err != nil {
		return nil, err
	}

	statusInfo.Forum = int(temp.Int64)

	err = handler.db.QueryRow(`SELECT sum(coalesce(thread_count, 0)) from forum`).Scan(&temp)
	if err != nil {
		return nil, err
	}

	statusInfo.Thread = int(temp.Int64)

	err = handler.db.QueryRow(`SELECT sum(coalesce(post_count, 0)) from forum`).Scan(&temp)
	if err != nil {
		return nil, err
	}

	statusInfo.Post = int(temp.Int64)

	return statusInfo, nil
}
