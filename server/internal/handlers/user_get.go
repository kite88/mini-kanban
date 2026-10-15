package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-kanban/internal/db"
	"mini-kanban/internal/models"
)

func GetUser(c *gin.Context) {
	id, good := parseUserID(c, "id")
	if !good {
		return
	}
	var u models.User
	if err := db.DB.First(&u, id).Error; err != nil {
		fail(c, http.StatusNotFound, "用户不存在")
		return
	}
	ok(c, u)
}
