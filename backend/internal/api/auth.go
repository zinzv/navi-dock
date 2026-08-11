package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/navi-dock/navi-dock/internal/model"
	"github.com/navi-dock/navi-dock/internal/service"
)

func (h *Handler) AuthStatus(c *gin.Context) {
	status, err := h.auth.Status(bearerToken(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *Handler) Login(c *gin.Context) {
	var body model.LoginInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	result, err := h.auth.Login(body)
	if err == service.ErrInvalidCredentials {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": service.ErrUnauthorized.Error()})
		return
	}
	var body model.ChangePasswordInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	err := h.auth.ChangePassword(user.ID, body)
	switch err {
	case nil:
		c.JSON(http.StatusOK, gin.H{"ok": true})
	case service.ErrPasswordEmpty, service.ErrPasswordTooShort:
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case service.ErrWrongPassword:
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case service.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}

func (h *Handler) ListUsers(c *gin.Context) {
	user := currentUser(c)
	users, err := h.auth.ListUsers(user)
	if err == service.ErrUnauthorized {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var body model.CreateUserInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	created, err := h.auth.CreateUser(body, currentUser(c))
	switch err {
	case nil:
		c.JSON(http.StatusCreated, created)
	case service.ErrUsernameEmpty, service.ErrUsernameTooShort, service.ErrUsernameTooLong,
		service.ErrPasswordEmpty, service.ErrPasswordTooShort, service.ErrUserExists:
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case service.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}

func (h *Handler) DeleteUser(c *gin.Context) {
	targetID := c.Param("id")
	err := h.auth.DeleteUser(currentUser(c), targetID)
	switch err {
	case nil:
		if err := h.navigation.DeleteUserData(targetID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		if err := h.settings.DeleteUserData(targetID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	case service.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case service.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case service.ErrLastAdmin:
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}

func (h *Handler) requireAuth(c *gin.Context) {
	token := bearerToken(c)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": service.ErrUnauthorized.Error()})
		return
	}
	user, err := h.auth.UserFromToken(token)
	if err != nil || user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": service.ErrUnauthorized.Error()})
		return
	}
	c.Set("authUser", user)
	c.Next()
}

func (h *Handler) optionalAuth(c *gin.Context) {
	token := bearerToken(c)
	if token == "" {
		c.Next()
		return
	}
	user, err := h.auth.UserFromToken(token)
	if err == nil && user != nil {
		c.Set("authUser", user)
	}
	c.Next()
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	const prefix = "Bearer "
	if strings.HasPrefix(h, prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return strings.TrimSpace(h)
}

func currentUser(c *gin.Context) *model.User {
	v, ok := c.Get("authUser")
	if !ok {
		return nil
	}
	user, _ := v.(*model.User)
	return user
}
