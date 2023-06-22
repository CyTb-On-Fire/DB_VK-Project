package threads

import (
	"DBProject/internal/db/threads"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"net/http"
	"time"
)

type CreateRequest struct {
	Title   string    `json:"title" binding:"required"`
	Author  string    `json:"author" binding:"required"`
	Message string    `json:"message" binding:"required"`
	Created time.Time `json:"created" binding:"required"`
}

type ThreadsHandler struct {
	Threads *threads.ThreadStorage
}

func New(pool *pgx.ConnPool) *ThreadsHandler {
	return &ThreadsHandler{Threads: threads.New(pool)}
}

func (handler *ThreadsHandler) Create(c *gin.Context) {
	forumSlug := c.Param("slug")

	thread := &models.Thread{
		Forum: forumSlug,
	}

	err := c.Bind(thread)

	if err != nil {
		utils.WriteError(c, http.StatusBadRequest, err)
		return
	}

	var code int

	thread, err = handler.Threads.Insert(thread)

	switch err {
	case nil:
		code = http.StatusOK
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can`t find user or forum"})
		return
	case utils.ErrConflict:
		code = http.StatusConflict
		thread, err = handler.Threads.GetBySlug(forumSlug)
	default:
		code = http.StatusInternalServerError
	}

	c.JSON(code, thread)
}
