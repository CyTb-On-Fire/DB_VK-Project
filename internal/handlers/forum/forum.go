package forum

import (
	"DBProject/internal/common"
	"DBProject/internal/db/forum"
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
	Forums *forum.ForumStorage
}

func NewForumHandler(pool *pgx.ConnPool) *ForumHandler {
	return &ForumHandler{
		Forums: forum.NewForumStorage(pool),
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

func (handler *ForumHandler) GetUsers(c *gin.Context) {
	forumSlug := c.Param("slug")

	_, err := handler.Forums.GetBySlug(forumSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find thread by slug or id"})
		return
	}

	params := &common.ListParams{
		Slug: forumSlug,
	}

	err = c.Bind(params)
	if err != nil {
		utils.WriteError(c, http.StatusBadRequest, err)
		return
	}

	log.Println(params)

	users, err := handler.Forums.GetUsers(params)
	//users, err := handler.Forums.GetUsersWithInterTable(params)

	switch err {
	case nil:
		c.JSON(http.StatusOK, users)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can`t find forum with id: " + forumSlug})
	default:
		c.JSON(http.StatusInternalServerError, err)
	}

}

func (handler *ForumHandler) GetThreads(c *gin.Context) {
	forumSlug := c.Param("slug")

	params := &common.ThreadListParams{Slug: forumSlug}

	err := c.BindQuery(params)

	if err != nil {
		utils.WriteError(c, http.StatusBadRequest, err)
		return
	}

	threads, err := handler.Forums.GetThreads(params)

	switch err {
	case nil:
		c.JSON(http.StatusOK, threads)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can`t find forum with id: " + forumSlug})
	default:
		c.JSON(http.StatusInternalServerError, err)
	}
}
