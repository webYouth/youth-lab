package controller

import (
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"youthlab/lottery-ai/internal/auth"
)

type authBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *API) Register(c *gin.Context) {
	var req authBody
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid json")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if utf8.RuneCountInString(req.Username) < 3 || utf8.RuneCountInString(req.Username) > 32 {
		Fail(c, 400, "用户名长度需 3～32")
		return
	}
	if len(req.Password) < 6 {
		Fail(c, 400, "密码至少 6 位")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	u, err := a.Store.CreateUser(c.Request.Context(), req.Username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			Fail(c, 409, "用户名已存在")
			return
		}
		Fail(c, 500, err.Error())
		return
	}
	token, err := auth.SignToken(a.JWTSecret, u.ID, u.Username)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"token": token, "user": gin.H{"id": u.ID, "username": u.Username}})
}

func (a *API) Login(c *gin.Context) {
	var req authBody
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid json")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	u, err := a.Store.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil || u == nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		Fail(c, 401, "用户名或密码错误")
		return
	}
	token, err := auth.SignToken(a.JWTSecret, u.ID, u.Username)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"token": token, "user": gin.H{"id": u.ID, "username": u.Username}})
}

func (a *API) Me(c *gin.Context) {
	uid, _ := c.Get("user_id")
	username, _ := c.Get("username")
	OK(c, gin.H{"id": uid, "username": username})
}
