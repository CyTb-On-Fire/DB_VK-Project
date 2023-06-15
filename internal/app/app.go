package app

import "github.com/gin-gonic/gin"

type App struct {
}

func New() *App {
	return &App{}
}

func (app *App) Run() {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

}
