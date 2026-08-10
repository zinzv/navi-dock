package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/navi-dock/navi-dock/internal/model"
	"github.com/navi-dock/navi-dock/internal/service"
)

type Handler struct {
	settings   *service.SettingsService
	navigation *service.NavigationService
	dataDir    string
}

func NewHandler(settings *service.SettingsService, navigation *service.NavigationService, dataDir string) *Handler {
	return &Handler{settings: settings, navigation: navigation, dataDir: dataDir}
}

func (h *Handler) Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/settings", h.GetSettings)
		api.PUT("/settings", h.UpdateSettings)
		api.POST("/settings/background", h.UploadBackground)
		api.DELETE("/settings/background", h.ClearBackground)
		api.POST("/settings/site-icon", h.UploadSiteIcon)
		api.DELETE("/settings/site-icon", h.ClearSiteIcon)
		api.GET("/assets/:kind", h.ListAssets)
		api.POST("/assets/:kind", h.UploadAsset)
		api.GET("/navigation", h.GetNavigation)

		api.GET("/groups", h.ListGroups)
		api.POST("/groups", h.CreateGroup)
		api.PUT("/groups/sort", h.SortGroups)
		api.PUT("/groups/:id", h.UpdateGroup)
		api.DELETE("/groups/:id", h.DeleteGroup)

		api.POST("/items", h.CreateItem)
		api.PUT("/items/sort", h.SortItems)
		api.PUT("/items/:id", h.UpdateItem)
		api.DELETE("/items/:id", h.DeleteItem)

		api.POST("/import/sunpanel", h.ImportSunPanel)
		api.GET("/export", h.ExportNavigation)
		api.POST("/import", h.ImportNavigation)
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "naviDock"})
}

func (h *Handler) GetSettings(c *gin.Context) {
	settings, err := h.settings.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var body model.AppSettings
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	settings, err := h.settings.Update(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UploadBackground(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing file"})
		return
	}
	if file.Size > 8<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file too large (max 8MB)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unsupported image type"})
		return
	}

	dir := filepath.Join(h.dataDir, "assets", "wallpapers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	name := fmt.Sprintf("bg-%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	publicPath := "/assets/wallpapers/" + name
	settings, err := h.settings.SetBackgroundImage(publicPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) ClearBackground(c *gin.Context) {
	settings, err := h.settings.SetBackgroundImage("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UploadSiteIcon(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing file"})
		return
	}
	if file.Size > 2<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file too large (max 2MB)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true, ".svg": true, ".ico": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unsupported image type"})
		return
	}

	// Site icons share the icons gallery library (no separate /assets/site).
	dir := filepath.Join(h.dataDir, "assets", "icons")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	name := fmt.Sprintf("icon-%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	publicPath := "/assets/icons/" + name
	settings, err := h.settings.SetSiteIcon(publicPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) ClearSiteIcon(c *gin.Context) {
	settings, err := h.settings.SetSiteIcon("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func assetKindDir(kind string) (dirName string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "icons", "icon":
		return "icons", true
	case "wallpapers", "wallpaper", "backgrounds", "background", "bg":
		return "wallpapers", true
	default:
		return "", false
	}
}

func (h *Handler) ListAssets(c *gin.Context) {
	dirName, ok := assetKindDir(c.Param("kind"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "kind must be icons or wallpapers"})
		return
	}
	dir := filepath.Join(h.dataDir, "assets", dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true, ".svg": true}
	if dirName == "icons" {
		allowed[".ico"] = true
	}
	type assetItem struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	items := make([]assetItem, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !allowed[ext] {
			continue
		}
		items = append(items, assetItem{
			Name: name,
			URL:  "/assets/" + dirName + "/" + name,
		})
	}
	// newest first by name prefix timestamp when possible
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	c.JSON(http.StatusOK, gin.H{"kind": dirName, "items": items})
}

func (h *Handler) UploadAsset(c *gin.Context) {
	dirName, ok := assetKindDir(c.Param("kind"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "kind must be icons or wallpapers"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing file"})
		return
	}
	if file.Size > 8<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file too large (max 8MB)"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true, ".svg": true}
	if dirName == "icons" {
		allowed[".ico"] = true
	}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unsupported image type"})
		return
	}
	dir := filepath.Join(h.dataDir, "assets", dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	prefix := "icon"
	if dirName == "wallpapers" {
		prefix = "bg"
	}
	name := fmt.Sprintf("%s-%d%s", prefix, time.Now().UnixNano(), ext)
	dst := filepath.Join(dir, name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	publicPath := "/assets/" + dirName + "/" + name
	c.JSON(http.StatusOK, gin.H{"kind": dirName, "name": name, "url": publicPath})
}

func (h *Handler) GetNavigation(c *gin.Context) {
	groups, err := h.navigation.ListNavigation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *Handler) ListGroups(c *gin.Context) {
	groups, err := h.navigation.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *Handler) CreateGroup(c *gin.Context) {
	var body model.CreateGroupInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	group, err := h.navigation.CreateGroup(body)
	if err == service.ErrGroupNameEmpty {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
}

func (h *Handler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var body model.UpdateGroupInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	group, err := h.navigation.UpdateGroup(id, body)
	if err == service.ErrGroupNameEmpty {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if err == service.ErrGroupNotFound {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	err := h.navigation.DeleteGroup(id)
	if err == service.ErrGroupNotFound {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) SortGroups(c *gin.Context) {
	var body model.SortGroupsInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	if err := h.navigation.SortGroups(body.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	groups, err := h.navigation.ListGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *Handler) SortItems(c *gin.Context) {
	var body model.SortItemsInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	if err := h.navigation.SortItems(body.GroupID, body.IDs); err != nil {
		switch err {
		case service.ErrGroupNotFound:
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) CreateItem(c *gin.Context) {
	var body model.UpsertItemInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	item, err := h.navigation.CreateItem(body)
	switch err {
	case service.ErrItemNameEmpty, service.ErrItemURLEmpty:
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	case service.ErrGroupNotFound:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	var body model.UpsertItemInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	item, err := h.navigation.UpdateItem(id, body)
	switch err {
	case service.ErrItemNameEmpty, service.ErrItemURLEmpty:
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	case service.ErrItemNotFound, service.ErrGroupNotFound:
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	err := h.navigation.DeleteItem(id)
	if err == service.ErrItemNotFound {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ImportSunPanel(c *gin.Context) {
	var body model.SunPanelExport
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid sun-panel json"})
		return
	}
	if len(body.Icons) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no icons/groups in export"})
		return
	}
	groups, items, err := h.navigation.ImportSunPanel(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"group_count": groups,
		"item_count":  items,
		"source":      body.AppName,
		"export_time": body.ExportTime,
	})
}

func (h *Handler) ExportNavigation(c *gin.Context) {
	payload, err := h.navigation.ExportNavigation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	filename := fmt.Sprintf("navi-dock-export-%s.json", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.JSON(http.StatusOK, payload)
}

func (h *Handler) ImportNavigation(c *gin.Context) {
	raw, err := c.GetRawData()
	if err != nil || len(raw) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "empty body"})
		return
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}

	if _, ok := probe["icons"]; ok {
		var body model.SunPanelExport
		if err := json.Unmarshal(raw, &body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid sun-panel json"})
			return
		}
		if len(body.Icons) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "no icons/groups in export"})
			return
		}
		groups, items, err := h.navigation.ImportSunPanel(body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"ok":          true,
			"format":      "sunpanel",
			"group_count": groups,
			"item_count":  items,
		})
		return
	}

	var body model.NaviDockExport
	if err := json.Unmarshal(raw, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid navi-dock json"})
		return
	}
	if len(body.Groups) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no groups in export"})
		return
	}
	groups, items, err := h.navigation.ImportNaviDock(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"format":      "navidock",
		"group_count": groups,
		"item_count":  items,
	})
}
