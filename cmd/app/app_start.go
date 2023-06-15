package app

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
)

func start() {
	db, err := pgx.Connect()
	r := gin.Default()
	_ = r
}
