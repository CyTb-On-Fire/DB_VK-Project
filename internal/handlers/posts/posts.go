package posts

import (
	"DBProject/internal/db/posts"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
)

type CreatePostRequest struct {
	Posts []*models.Post `json:"posts" binding:"required"`
}

type PostHandler struct {
	Posts *posts.PostStorage
}

func New(pool *pgx.ConnPool) *PostHandler {
	return &PostHandler{Posts: posts.NewStorage(pool)}
}

func (handler *PostHandler) Create(c *gin.Context) {
	threadId := c.Param("slug")

	var stockPosts []*models.Post

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	err = json.Unmarshal(body, &stockPosts)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	for _, post := range stockPosts {
		post.ThreadId = threadId
	}

	posts, err := handler.Posts.Insert(stockPosts)

	switch err {
	case nil:
		c.JSON(http.StatusOK, posts)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
	case utils.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"message": "Can't find user with id #42\n"})
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}
}
