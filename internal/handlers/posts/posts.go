package posts

import (
	"DBProject/internal/common"
	"DBProject/internal/db/forum"
	"DBProject/internal/db/posts"
	"DBProject/internal/db/threads"
	"DBProject/internal/db/user"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"io/ioutil"
	"net/http"
	"strconv"
)

type UpdateRequest struct {
	Message string `json:"message"`
}

type PostExtended struct {
	Post   *models.Post   `json:"post"`
	Author *models.User   `json:"author,omitempty"`
	Thread *models.Thread `json:"thread,omitempty"`
	Forum  *models.Forum  `json:"forum,omitempty"`
}

type CreatePostRequest struct {
	Posts []*models.Post `json:"posts" binding:"required"`
}

type PostHandler struct {
	Posts   *posts.PostStorage
	Users   *user.UserStorage
	Threads *threads.ThreadStorage
	Forums  *forum.ForumStorage
}

func New(pool *pgx.ConnPool) *PostHandler {
	return &PostHandler{
		Posts:   posts.NewStorage(pool),
		Users:   user.NewUserStorage(pool),
		Threads: threads.New(pool),
		Forums:  forum.NewForumStorage(pool),
	}
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

func (handler *PostHandler) Details(c *gin.Context) {
	postId := c.Param("id")

	id, err := strconv.Atoi(postId)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	params := common.PostViewParams{}

	err = c.Bind(params)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	params.Id = id

	response := &PostExtended{}

	response.Post, err = handler.Posts.Details(id)

	if err != nil {
		if err == utils.ErrNonExist {
			c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
			return
		}
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	for _, opt := range params.Params {
		switch opt {
		case "user":
			response.Author, err = handler.Users.GetByNickname(response.Post.Author)
		case "forum":
			response.Forum, err = handler.Forums.GetBySlug(response.Post.ForumSlug)
		case "thread":
			response.Thread, err = handler.Threads.GetById(response.Post.Id)
		}

		if err != nil {
			utils.WriteError(c, http.StatusInternalServerError, err)
			return
		}
	}

	c.JSON(http.StatusOK, response)
}

func (handler *PostHandler) Update(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	request := new(UpdateRequest)

	err = c.Bind(request)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	post, err := handler.Posts.Update(id, request.Message)

	switch err {
	case nil:
		c.JSON(http.StatusOK, post)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}
}
