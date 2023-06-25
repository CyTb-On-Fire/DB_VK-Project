package threads

import (
	"DBProject/internal/common"
	"DBProject/internal/db/forum"
	"DBProject/internal/db/threads"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"log"
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

type ThreadProxy struct {
	models.Thread
}

type UpdateRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ThreadsHandler struct {
	Threads *threads.ThreadStorage
	Forums  *forum.ForumStorage
}

func New(pool *pgx.ConnPool) *ThreadsHandler {
	return &ThreadsHandler{
		Threads: threads.New(pool),
		Forums:  forum.NewForumStorage(pool),
	}
}

func (handler *ThreadsHandler) Create(c *gin.Context) {
	forumSlug := c.Param("slug")

	linkingForum, err := handler.Forums.GetBySlug(forumSlug)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Can`t find user or forum"})
		return
	}

	thread := &models.Thread{
		Forum: forumSlug,
	}

	err = c.Bind(thread)

	if err != nil {
		utils.WriteError(c, http.StatusBadRequest, err)
		return
	}

	if thread.Created.Equal(time.Time{}) {
		thread.Created = time.Now()
	}

	threadSlug := thread.Slug

	var code int

	thread, err = handler.Threads.Insert(thread, linkingForum)

	switch err {
	case nil:
		code = http.StatusCreated
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can`t find user or forum"})
		return
	case utils.ErrConflict:
		code = http.StatusConflict
		thread, err = handler.Threads.GetBySlug(threadSlug)
		log.Println(thread)
	default:
		code = http.StatusInternalServerError
	}

	c.JSON(code, thread)
}

func (handler *ThreadsHandler) Detailss(c *gin.Context) {
	threadSlug := c.Param("slug")

	var thread *models.Thread
	var err error

	threadId, convErr := strconv.Atoi(threadSlug)

	if convErr != nil {
		thread, err = handler.Threads.GetBySlug(threadSlug)
	} else {
		thread, err = handler.Threads.GetById(threadId)
	}

	log.Println(thread)

	if err != nil {
		if err == utils.ErrNonExist {
			c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
			return
		}
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	log.Println(thread)
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

	log.Println(thread)

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

	log.Println(request)

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

	tId, convErr := strconv.Atoi(threadSlug)

	var err error
	var th *models.Thread

	if convErr != nil {
		th, err = handler.Threads.GetBySlug(threadSlug)
	} else {
		th, err = handler.Threads.GetById(tId)
	}

	if err == utils.ErrNonExist {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("can`t find thread with id %s", threadSlug)})
		return
	}

	Params := new(common.FilterParams)

	Params.ThreadSlug = threadSlug

	err = c.Bind(Params)
	if err != nil {
		log.Println(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Println(Params)

	posts := make([]*models.Post, 0)

	switch Params.Sort {
	case "flat":
		posts, err = handler.Threads.GetPostsWithFlat(Params)
	case "tree":
		posts, err = handler.Threads.GetPostsWithTree(Params)
	case "parent_tree":
		posts, err = handler.Threads.GetPostsWithParentTree(Params)
	default:
		posts, err = handler.Threads.GetPostsWithFlat(Params)
	}

	log.Println(posts)

	switch err {
	case nil:
		proxyPosts := make([]*models.ProxyPost, 0)
		for _, post := range posts {
			id, err := strconv.Atoi(post.ThreadId)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			proxyPosts = append(proxyPosts, &models.ProxyPost{
				Id:        post.Id,
				ParentId:  post.ParentId,
				Author:    post.Author,
				Message:   post.Message,
				Edited:    post.Edited,
				ForumSlug: th.Forum,
				Created:   post.Created,
				ThreadId:  id,
			})
		}
		c.JSON(http.StatusOK, proxyPosts)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": "Can't find user with id #42\n"})
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}

}

func (handler *ThreadsHandler) Vote(c *gin.Context) {
	threadSlug := c.Param("slug")

	log.Println(threadSlug)

	voteReq := new(common.Vote)

	voteReq.ThreadSlug = threadSlug

	err := c.Bind(voteReq)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	log.Println(voteReq)

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
