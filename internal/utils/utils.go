package utils

import "github.com/gin-gonic/gin"

func WriteError(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{"Error": err.Error()})
}
