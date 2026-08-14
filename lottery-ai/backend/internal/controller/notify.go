package controller

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"

	"youthlab/lottery-ai/internal/model"
)

type notifyReq struct {
	Type    string          `json:"type"`
	Title   string          `json:"title"`
	Body    string          `json:"body"`
	Payload json.RawMessage `json:"payload"`
}

func (a *API) InternalNotify(c *gin.Context) {
	var req notifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "invalid json")
		return
	}
	if req.Title == "" {
		Fail(c, 400, "title required")
		return
	}
	if req.Type == "" {
		req.Type = "kl8"
	}
	if a.Notify == nil {
		Fail(c, 500, "notify service unavailable")
		return
	}
	if err := a.Notify.Publish(c.Request.Context(), model.AppNotification{
		Type:    req.Type,
		Title:   req.Title,
		Body:    req.Body,
		Payload: req.Payload,
	}); err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

func (a *API) ListNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	list, total, err := a.Store.ListNotifications(c.Request.Context(), page, pageSize)
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	unread, _ := a.Store.UnreadCount(c.Request.Context())
	OK(c, gin.H{"list": list, "total": total, "unread": unread, "page": page})
}

func (a *API) UnreadCount(c *gin.Context) {
	n, err := a.Store.UnreadCount(c.Request.Context())
	if err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"unread": n})
}

func (a *API) MarkNotificationRead(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		Fail(c, 400, "invalid id")
		return
	}
	if err := a.Store.MarkNotificationRead(c.Request.Context(), id); err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

func (a *API) MarkAllRead(c *gin.Context) {
	if err := a.Store.MarkAllNotificationsRead(c.Request.Context()); err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

type batchNotifyReq struct {
	IDs  []int64 `json:"ids"`
	Read bool    `json:"read"`
}

func (a *API) BatchSetNotificationRead(c *gin.Context) {
	var req batchNotifyReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		Fail(c, 400, "ids required")
		return
	}
	if err := a.Store.SetNotificationsRead(c.Request.Context(), req.IDs, req.Read); err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

func (a *API) MarkNotificationUnread(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id <= 0 {
		Fail(c, 400, "invalid id")
		return
	}
	if err := a.Store.SetNotificationsRead(c.Request.Context(), []int64{id}, false); err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

type deviceReq struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func (a *API) RegisterDevice(c *gin.Context) {
	var req deviceReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		Fail(c, 400, "token required")
		return
	}
	if req.Platform == "" {
		req.Platform = "android"
	}
	if err := a.Store.UpsertPushDevice(c.Request.Context(), req.Token, req.Platform); err != nil {
		Fail(c, 500, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}
