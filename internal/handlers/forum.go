package handlers

import (
	"DBProject/internal/db"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"log"
	"net/http"
)

type CreateRequest struct {
	Title string `json:"title" binding:"required"`
	User  string `json:"user" binding:"required"`
	Slug  string `json:"slug" binding:"required"`
}

type ForumHandler struct {
	Forums *db.ForumStorage
}

func NewForumHandler(pool *pgx.ConnPool) *ForumHandler {
	return &ForumHandler{
		Forums: db.NewForumStorage(pool),
	}
}

func (handler *ForumHandler) Create(c *gin.Context) {
	request := new(CreateRequest)

	err := c.Bind(request)
	if err != nil {
		log.Println("Bind Error in Forum-Create: ", err)
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	forum, err := handler.Forums.InsertForum(&models.Forum{
		Title:    request.Title,
		UserName: request.User,
		Slug:     request.Slug,
	})

	code := http.StatusOK

	switch err {
	case nil:
		code = http.StatusCreated
	case utils.ErrConflict:
		code = http.StatusConflict
		forum, err = handler.Forums.GetBySlug(request.Slug)
		if err != nil {
			utils.WriteError(c, http.StatusInternalServerError, err)
		}

	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Can`t find user with id #%v\n", request.User)})
		return
	}

	c.JSON(code, forum)
}

func (handler *ForumHandler) Details(c *gin.Context) {
	forumSlug := c.Param("slug")

	forum, err := handler.Forums.GetBySlug(forumSlug)

	if err != nil {
		if err == utils.ErrNonExist {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Can`t find forum with id: %v", forumSlug)})
			return
		}
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, forum)
}

func (handler *ForumHandler) CreateThread()
