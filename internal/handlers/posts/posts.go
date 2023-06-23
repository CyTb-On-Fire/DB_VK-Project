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
	"log"
	"net/http"
	"strconv"
	"strings"
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

type ProxyPostExtended struct {
	Post   *models.ProxyPost `json:"post"`
	Author *models.User      `json:"author,omitempty"`
	Thread *models.Thread    `json:"thread,omitempty"`
	Forum  *models.Forum     `json:"forum,omitempty"`
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

	id, err := strconv.Atoi(threadId)

	if err != nil {
		_, err = handler.Threads.GetBySlug(threadId)
	} else {
		_, err = handler.Threads.GetById(id)
	}

	if err != nil {
		c.JSON(http.StatusNotFound,
			gin.H{
				"message": "Can't find post thread by id"})
		return
	}

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

	log.Println(stockPosts)

	posts, err := handler.Posts.Insert(stockPosts)

	if err == utils.ErrConflict {
		c.JSON(http.StatusConflict, gin.H{"message": "Parent post was created in another thread"})
		return
	}

	if err == utils.ErrNonExist {
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find post thread by id"})
		return
	}

	proxyPosts := make([]*models.ProxyPost, 0)

	for _, post := range posts {
		id, err := strconv.Atoi(post.ThreadId)
		if err != nil {
			log.Println("Got error")
			break
		}
		proxyPosts = append(proxyPosts, &models.ProxyPost{
			Id:        post.Id,
			ParentId:  post.ParentId,
			Author:    post.Author,
			Message:   post.Message,
			Edited:    post.Edited,
			ForumSlug: post.ForumSlug,
			Created:   post.Created,
			ThreadId:  id,
		})
	}
	log.Println(posts)

	switch err {
	case nil:
		c.JSON(http.StatusCreated, proxyPosts)
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

	params := &common.PostViewParams{}

	err = c.BindQuery(params)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	log.Println(params)

	params.Id = id

	response := &ProxyPostExtended{}

	post, err := handler.Posts.Details(id)

	if err == utils.ErrNonExist {
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find post with id"})
		return
	}
	tId, _ := strconv.Atoi(post.ThreadId)
	response.Post = &models.ProxyPost{
		Id:        post.Id,
		ParentId:  post.ParentId,
		Author:    post.Author,
		Message:   post.Message,
		Edited:    post.Edited,
		ForumSlug: post.ForumSlug,
		Created:   post.Created,
		ThreadId:  tId,
	}

	log.Println(post)

	if err != nil {
		if err == utils.ErrNonExist {
			c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
			return
		}
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}
	log.Println(params.Params)

	if len(params.Params) > 0 {
		params.Params = strings.Split(params.Params[0], ",")
	}

	for _, opt := range params.Params {
		switch opt {
		case "user":
			log.Println("entered user")
			response.Author, err = handler.Users.GetByNickname(post.Author)
			if err != nil {
				log.Println(err)
			}
		case "forum":
			log.Println("entered forum")
			response.Forum, err = handler.Forums.GetBySlug(post.ForumSlug)
			if err != nil {
				log.Println(err)
			}
		case "thread":
			log.Println("entered thread")
			response.Thread, err = handler.Threads.GetById(tId)
			if err != nil {
				log.Println(err)
			}
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
	log.Println("Started Update handler")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	request := new(UpdateRequest)

	err = c.BindQuery(request)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	post, err := handler.Posts.Update(id, request.Message)
	log.Println(err)

	tId, _ := strconv.Atoi(post.ThreadId)

	proxyPost := &models.ProxyPost{
		Id:        post.Id,
		ParentId:  post.ParentId,
		Author:    post.Author,
		Message:   post.Message,
		Edited:    post.Edited,
		ForumSlug: post.ForumSlug,
		Created:   post.Created,
		ThreadId:  tId,
	}
	switch err {
	case nil:
		c.JSON(http.StatusOK, proxyPost)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}
}
