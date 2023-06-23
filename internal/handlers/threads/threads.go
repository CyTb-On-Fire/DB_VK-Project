package threads

import (
	"DBProject/internal/common"
	"DBProject/internal/db/threads"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"
	"time"
)

type CreateRequest struct {
	Title   string    `json:"title" binding:"required"`
	Author  string    `json:"author" binding:"required"`
	Message string    `json:"message" binding:"required"`
	Created time.Time `json:"created" binding:"required"`
}

type UpdateRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
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

func (handler *ThreadsHandler) Details(c *gin.Context) {
	threadSlug := c.Param("slug")

	var thread *models.Thread
	var err error

	threadId, convErr := strconv.Atoi(threadSlug)

	if convErr != nil {
		thread, err = handler.Threads.GetBySlug(threadSlug)
	} else {
		thread, err = handler.Threads.GetById(threadId)
	}

	if err != nil {
		if err == utils.ErrNonExist {
			c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
			return
		}
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, thread)
}

func (handler *ThreadsHandler) Update(c *gin.Context) {
	threadSlug := c.Param("slug")

	threadId, convErr := strconv.Atoi(threadSlug)

	var thread *models.Thread
	var err error

	if convErr != nil {
		thread, err = handler.Threads.GetBySlug(threadSlug)
	} else {
		thread, err = handler.Threads.GetById(threadId)
	}

	if err != nil {
		if err == utils.ErrNonExist {
			c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
			return
		}
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	request := new(UpdateRequest)

	err = c.Bind(request)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	if request.Title != "" {
		thread.Title = request.Title
	}
	if request.Message != "" {
		thread.Message = request.Message
	}

	thread, err = handler.Threads.Update(thread)
	c.JSON(http.StatusOK, thread)
}

func (handler *ThreadsHandler) GetPosts(c *gin.Context) {
	threadSlug := c.Param("slug")

	Params := new(common.FilterParams)

	Params.ThreadSlug = threadSlug

	err := c.Bind(Params)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	posts := make([]*models.Post, 0)

	switch Params.Sort {
	case "flat":
		posts, err = handler.Threads.GetPostsWithFlat(Params)
	case "tree":
		posts, err = handler.Threads.GetPostsWithTree(Params)
	case "parent_tree":
		posts, err = handler.Threads.GetPostsWithParentTree(Params)
	default:
		c.AbortWithError(http.StatusBadRequest, nil)
		return
	}

	switch err {
	case nil:
		c.JSON(http.StatusOK, posts)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}

}

func (handler *ThreadsHandler) Vote(c *gin.Context) {
	threadSlug := c.Param("slug")

	voteReq := new(common.Vote)

	voteReq.ThreadSlug = threadSlug

	err := c.Bind(voteReq)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	thread, err := handler.Threads.NewVote(voteReq)
	switch err {
	case nil:
		c.JSON(http.StatusOK, thread)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}
}
