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
	"github.com/navi-dock/navi-dock/internal/version"
)

type Handler struct {
	settings    *service.SettingsService
	navigation  *service.NavigationService
	auth        *service.AuthService
	dataDir        string
	lanProbeDomain string
}

func NewHandler(
	settings *service.SettingsService,
	navigation *service.NavigationService,
	auth *service.AuthService,
	dataDir string,
	lanProbeDomain string,
) *Handler {
	return &Handler{
		settings:       settings,
		navigation:     navigation,
		auth:           auth,
		dataDir:        dataDir,
		lanProbeDomain: strings.TrimSpace(lanProbeDomain),
	}
}

func (h *Handler) Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/health", h.Health)
		api.GET("/network/ping", h.NetworkPing)

		api.GET("/auth/status", h.optionalAuth, h.AuthStatus)
		api.POST("/auth/login", h.Login)
		api.PUT("/auth/password", h.requireAuth, h.ChangePassword)

		api.GET("/users", h.requireAuth, h.ListUsers)
		api.POST("/users", h.optionalAuth, h.CreateUser)
		api.DELETE("/users/:id", h.requireAuth, h.DeleteUser)

		api.GET("/settings", h.requireAuth, h.GetSettings)
		api.PUT("/settings", h.requireAuth, h.UpdateSettings)
		api.GET("/network/config", h.requireAuth, h.GetNetworkConfig)
		api.POST("/settings/background", h.requireAuth, h.UploadBackground)
		api.DELETE("/settings/background", h.requireAuth, h.ClearBackground)
		api.POST("/settings/site-icon", h.requireAuth, h.UploadSiteIcon)
		api.DELETE("/settings/site-icon", h.requireAuth, h.ClearSiteIcon)
		api.GET("/assets/:kind", h.requireAuth, h.ListAssets)
		api.POST("/assets/:kind", h.requireAuth, h.UploadAsset)
		api.DELETE("/assets/:kind/:name", h.requireAuth, h.DeleteAsset)
		api.GET("/navigation", h.requireAuth, h.GetNavigation)

		api.GET("/groups", h.requireAuth, h.ListGroups)
		api.POST("/groups", h.requireAuth, h.CreateGroup)
		api.PUT("/groups/sort", h.requireAuth, h.SortGroups)
		api.PUT("/groups/:id", h.requireAuth, h.UpdateGroup)
		api.DELETE("/groups/:id", h.requireAuth, h.DeleteGroup)

		api.POST("/items", h.requireAuth, h.CreateItem)
		api.PUT("/items/sort", h.requireAuth, h.SortItems)
		api.PUT("/items/:id", h.requireAuth, h.UpdateItem)
		api.DELETE("/items/:id", h.requireAuth, h.DeleteItem)

		api.POST("/import/sunpanel", h.requireAuth, h.ImportSunPanel)
		api.GET("/export", h.requireAuth, h.ExportNavigation)
		api.POST("/import", h.requireAuth, h.ImportNavigation)
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "naviDock",
		"version": version.Current(),
	})
}

func (h *Handler) NetworkPing(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetNetworkConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"probe_domain": h.lanProbeDomain,
	})
}

func requireCurrentUser(c *gin.Context) *model.User {
	user := currentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": service.ErrUnauthorized.Error()})
		return nil
	}
	return user
}

func userAssetDir(dataDir, userID, dirName string) string {
	return filepath.Join(dataDir, "assets", "users", userID, dirName)
}

func userAssetURL(userID, dirName, name string) string {
	return "/assets/users/" + userID + "/" + dirName + "/" + name
}

func (h *Handler) GetSettings(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	settings, err := h.settings.Get(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	var body model.AppSettings
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	settings, err := h.settings.Update(user.ID, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UploadBackground(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
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
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".gif": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unsupported image type"})
		return
	}

	dir := userAssetDir(h.dataDir, user.ID, "wallpapers")
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

	publicPath := userAssetURL(user.ID, "wallpapers", name)
	settings, err := h.settings.SetBackgroundImage(user.ID, publicPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) ClearBackground(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	settings, err := h.settings.SetBackgroundImage(user.ID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) UploadSiteIcon(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
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

	dir := userAssetDir(h.dataDir, user.ID, "icons")
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

	publicPath := userAssetURL(user.ID, "icons", name)
	settings, err := h.settings.SetSiteIcon(user.ID, publicPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) ClearSiteIcon(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	settings, err := h.settings.SetSiteIcon(user.ID, "")
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	dirName, ok := assetKindDir(c.Param("kind"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "kind must be icons or wallpapers"})
		return
	}
	dir := userAssetDir(h.dataDir, user.ID, dirName)
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
			URL:  userAssetURL(user.ID, dirName, name),
		})
	}
	// newest first by name prefix timestamp when possible
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	c.JSON(http.StatusOK, gin.H{"kind": dirName, "items": items})
}

func (h *Handler) UploadAsset(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
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
	dir := userAssetDir(h.dataDir, user.ID, dirName)
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
	publicPath := userAssetURL(user.ID, dirName, name)
	c.JSON(http.StatusOK, gin.H{"kind": dirName, "name": name, "url": publicPath})
}

func (h *Handler) DeleteAsset(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	dirName, ok := assetKindDir(c.Param("kind"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "kind must be icons or wallpapers"})
		return
	}

	name := strings.TrimSpace(c.Param("name"))
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid asset name"})
		return
	}

	target := filepath.Join(userAssetDir(h.dataDir, user.ID, dirName), name)
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"message": "asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	publicPath := userAssetURL(user.ID, dirName, name)
	settings, err := h.settings.Get(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	if dirName == "icons" && settings.SiteIcon == publicPath {
		if _, err := h.settings.SetSiteIcon(user.ID, ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}
	if dirName == "wallpapers" && settings.BackgroundImage == publicPath {
		if _, err := h.settings.SetBackgroundImage(user.ID, ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) GetNavigation(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	groups, err := h.navigation.ListNavigation(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *Handler) ListGroups(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	groups, err := h.navigation.ListGroups(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *Handler) CreateGroup(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	var body model.CreateGroupInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	group, err := h.navigation.CreateGroup(user.ID, body)
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	id := c.Param("id")
	var body model.UpdateGroupInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	group, err := h.navigation.UpdateGroup(user.ID, id, body)
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	id := c.Param("id")
	err := h.navigation.DeleteGroup(user.ID, id)
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	var body model.SortGroupsInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	if err := h.navigation.SortGroups(user.ID, body.IDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	groups, err := h.navigation.ListGroups(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *Handler) SortItems(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	var body model.SortItemsInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	if err := h.navigation.SortItems(user.ID, body.GroupID, body.IDs); err != nil {
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	var body model.UpsertItemInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	item, err := h.navigation.CreateItem(user.ID, body)
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	id := c.Param("id")
	var body model.UpsertItemInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid json"})
		return
	}
	item, err := h.navigation.UpdateItem(user.ID, id, body)
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	id := c.Param("id")
	err := h.navigation.DeleteItem(user.ID, id)
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	var body model.SunPanelExport
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid sun-panel json"})
		return
	}
	if len(body.Icons) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "no icons/groups in export"})
		return
	}
	groups, items, err := h.navigation.ImportSunPanel(user.ID, body)
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
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
	payload, err := h.navigation.ExportNavigation(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	filename := fmt.Sprintf("navi-dock-export-%s.json", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.JSON(http.StatusOK, payload)
}

func (h *Handler) ImportNavigation(c *gin.Context) {
	user := requireCurrentUser(c)
	if user == nil {
		return
	}
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
		groups, items, err := h.navigation.ImportSunPanel(user.ID, body)
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
	groups, items, err := h.navigation.ImportNaviDock(user.ID, body)
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
