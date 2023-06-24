package app

import (
	"DBProject/internal/config"
	"DBProject/internal/handlers/forum"
	"DBProject/internal/handlers/posts"
	"DBProject/internal/handlers/service"
	threads2 "DBProject/internal/handlers/threads"
	"DBProject/internal/handlers/user"
	"github.com/gin-gonic/gin"
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

	//db, err := pgx.NewConnPool(config.PostgresPoolCong)

	db, err := config.InitPostgres()

	if err != nil {
		log.Println("App-Run NewConnPool error: ", err)
		return
	}

	forumHandler := forum.NewForumHandler(db)
	userHandler := user.New(db)
	threadHandler := threads2.New(db)
	postHandler := posts.New(db)
	serviceHandler := service.New(db)

	r.POST("/api/forum/create", forumHandler.Create)
	r.GET("/api/forum/:slug/details", forumHandler.Details)
	r.GET("/api/forum/:slug/users", forumHandler.GetUsers)
	r.GET("/api/forum/:slug/threads", forumHandler.GetThreads)

	r.POST("/api/user/:nickname/create", userHandler.Create)
	r.POST("/api/user/:nickname/profile", userHandler.EditProfile)
	r.GET("/api/user/:nickname/profile", userHandler.Profile)

	r.POST("/api/forum/:slug/create", threadHandler.Create)
	r.GET("/api/thread/:slug/details", threadHandler.Detailss)
	r.POST("/api/thread/:slug/vote", threadHandler.Vote)
	r.POST("/api/thread/:slug/details", threadHandler.Update)

	r.POST("/api/thread/:slug/create", postHandler.Create)
	r.GET("/api/thread/:slug/posts", threadHandler.GetPosts)
	r.GET("/api/post/:id/details", postHandler.Details)
	r.POST("/api/post/:id/details", postHandler.Update)

	r.GET("/api/service/status", serviceHandler.Status)
	r.POST("/api/service/clear", serviceHandler.Clear)

	err = r.Run("0.0.0.0:5000")
	if err != nil {
		log.Println(err)
	}
}
