package app

import (
	"DBProject/internal/config"
	"DBProject/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"log"
)

type App struct {
}

func New() *App {
	return &App{}
}

func (app *App) Run() {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	db, err := pgx.NewConnPool(config.PostgresPoolCong)

	if err != nil {
		log.Println("App-Run NewConnPool error: ", err)
		return
	}

	forumHandler := handlers.NewForumHandler(db)

	r.POST("/forum/create", forumHandler.Create)
	r.GET("/forum/:slug/details", forumHandler.Details)

	err = r.Run("0.0.0.0:5000")
	if err != nil {
		log.Println(err)
	}
}
