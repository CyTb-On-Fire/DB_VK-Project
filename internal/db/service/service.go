package service

import (
	"DBProject/internal/common"
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

	err := handler.db.QueryRow(`SELECT count(*) from users`).Scan(&statusInfo.User)
	if err != nil {
		return nil, err
	}

	err = handler.db.QueryRow(`SELECT count(*) from forum`).Scan(&statusInfo.Forum)
	if err != nil {
		return nil, err
	}

	err = handler.db.QueryRow(`SELECT sum(thread_count) from forum`).Scan(&statusInfo.Thread)
	if err != nil {
		return nil, err
	}

	err = handler.db.QueryRow(`SELECT sum(post_count) from forum`).Scan(&statusInfo.Post)
	if err != nil {
		return nil, err
	}

	return statusInfo, nil
}
