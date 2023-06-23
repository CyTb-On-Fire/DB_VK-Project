package service

import "github.com/jackc/pgx"

type ServiceController struct {
	db *pgx.ConnPool
}

func New(pool *pgx.ConnPool) *ServiceController {
	return &ServiceController{db: pool}
}

func (handler *ServiceController) Clear() error {
	handler.db.QueryRow(``)
	return nil
}

func (handler *ServiceController) Status() {

}
