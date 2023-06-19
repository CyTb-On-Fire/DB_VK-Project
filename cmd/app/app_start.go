package main

import (
	"DBProject/internal/app"
	"github.com/gin-gonic/gin"
)

func start() {
	//db, err := pgx.Connect()
	r := gin.Default()
	_ = r
}

func main() {
	mainApp := app.New()
	mainApp.Run()
}
