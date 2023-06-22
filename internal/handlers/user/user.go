package user

import (
	"DBProject/internal/db/user"
	"DBProject/internal/models"
	"DBProject/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"net/http"
)

type CreateReqeust struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	About    string `json:"about"`
}

type UserHandler struct {
	Users *user.UserStorage
}

func New(pool *pgx.ConnPool) *UserHandler {
	return &UserHandler{Users: user.NewUserStorage(pool)}
}

func (handler *UserHandler) Create(c *gin.Context) {
	userName := c.Param("nickname")

	request := new(CreateReqeust)
	err := c.Bind(request)

	if err != nil {
		utils.WriteError(c, http.StatusBadRequest, err)
		return
	}

	user, err := handler.Users.InsertUser(&models.User{
		Nickname: userName,
		Fullname: request.Fullname,
		About:    request.About,
		Email:    request.Email,
	})

	var code int

	switch err {
	case nil:
		code = http.StatusCreated
	case utils.ErrConflict:
		code = http.StatusConflict
		user, err = handler.Users.GetByNickname(userName)
		if err != nil {
			code = http.StatusInternalServerError
		}
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(code, user)
}

func (handler *UserHandler) Profile(c *gin.Context) {
	userName := c.Param("nickname")

	user, err := handler.Users.GetByNickname(userName)

	switch err {
	case nil:
		c.JSON(http.StatusOK, user)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Can`t find user with name: %v", userName)})
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}
}

func (handler *UserHandler) EditProfile(c *gin.Context) {
	userName := c.Param("nickname")

	updatingUser := new(models.User)

	err := c.Bind(updatingUser)

	updatingUser.Nickname = userName
	if err != nil {
		utils.WriteError(c, http.StatusBadRequest, err)
		return
	}

	updatingUser, err = handler.Users.UpdateUser(updatingUser)
	switch err {
	case nil:
		c.JSON(http.StatusOK, updatingUser)
	case utils.ErrNonExist:
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Can`t find user with nickname %v", userName)})
	case utils.ErrConflict:
		updatingUser, err = handler.Users.GetByEmail(updatingUser.Email)
		c.JSON(http.StatusConflict, updatingUser)
	default:
		utils.WriteError(c, http.StatusInternalServerError, err)
	}
}
