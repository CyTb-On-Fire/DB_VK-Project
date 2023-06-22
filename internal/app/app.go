package app

import (
	"DBProject/internal/config"
	"DBProject/internal/handlers/forum"
	threads2 "DBProject/internal/handlers/threads"
	"DBProject/internal/handlers/user"
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

	forumHandler := forum.NewForumHandler(db)
	userHandler := user.New(db)
	threadHandler := threads2.New(db)

	r.POST("/api/forum/create", forumHandler.Create)
	r.GET("/api/forum/:slug/details", forumHandler.Details)
	r.GET("/api/forum/:slug/users", forumHandler.GetUsers)
	r.GET("/api/forum/:slug/threads", forumHandler.GetThreads)

	r.POST("/api/user/:nickname/create", userHandler.Create)
	r.POST("/api/user/:nickname/profile", userHandler.EditProfile)
	r.GET("/api/user/:nickname/profile", userHandler.Profile)

	r.POST("/api/forum/:slug/create", threadHandler.Create)

	err = r.Run("0.0.0.0:5000")
	if err != nil {
		log.Println(err)
	}
}
