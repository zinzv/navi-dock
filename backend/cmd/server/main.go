package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/navi-dock/navi-dock/internal/api"
	"github.com/navi-dock/navi-dock/internal/config"
	"github.com/navi-dock/navi-dock/internal/repository"
	"github.com/navi-dock/navi-dock/internal/service"
)

func main() {
	cfg := config.Load()
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("create data dir: %v (uid=%d gid=%d)", err, os.Getuid(), os.Getgid())
	}
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Fatalf(
			"create database dir: %v (uid=%d gid=%d, PUID=%s PGID=%s)",
			err, os.Getuid(), os.Getgid(), os.Getenv("PUID"), os.Getenv("PGID"),
		)
	}
	migrateLegacyDB(cfg.DataDir, cfg.DBPath)

	db, err := repository.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	settingsRepo := repository.NewSettingsRepo(db)
	settingsSvc := service.NewSettingsService(settingsRepo)

	navRepo := repository.NewNavigationRepo(db)
	navSvc := service.NewNavigationService(navRepo)

	handler := api.NewHandler(settingsSvc, navSvc, cfg.DataDir)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), corsMiddleware())

	assetsDir := filepath.Join(cfg.DataDir, "assets")
	_ = os.MkdirAll(filepath.Join(assetsDir, "wallpapers"), 0o755)
	_ = os.MkdirAll(filepath.Join(assetsDir, "icons"), 0o755)
	migrateSiteAssetsToIcons(cfg.DataDir, settingsSvc)
	migrateBackgroundsToWallpapers(cfg.DataDir, settingsSvc)
	r.Static("/assets", assetsDir)

	handler.Register(r)
	mountStatic(r, cfg.WebDir)

	addr := ":" + cfg.Port
	log.Printf("naviDock listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// migrateBackgroundsToWallpapers renames legacy /assets/backgrounds into /assets/wallpapers.
func migrateBackgroundsToWallpapers(dataDir string, settingsSvc *service.SettingsService) {
	oldDir := filepath.Join(dataDir, "assets", "backgrounds")
	newDir := filepath.Join(dataDir, "assets", "wallpapers")
	entries, err := os.ReadDir(oldDir)
	if err != nil {
		// Still rewrite setting path if needed even when folder already gone.
		rewriteBackgroundImagePath(settingsSvc)
		return
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		log.Printf("migrate backgrounds: create wallpapers dir: %v", err)
		return
	}

	renames := map[string]string{}
	moved := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		oldName := e.Name()
		newName := oldName
		src := filepath.Join(oldDir, oldName)
		dst := filepath.Join(newDir, newName)
		if _, err := os.Stat(dst); err == nil {
			ext := filepath.Ext(oldName)
			base := strings.TrimSuffix(oldName, ext)
			newName = base + "-bg" + ext
			dst = filepath.Join(newDir, newName)
		}
		if err := os.Rename(src, dst); err != nil {
			log.Printf("migrate background asset %s: %v", oldName, err)
			continue
		}
		renames[oldName] = newName
		moved++
	}

	if settings, err := settingsSvc.Get(); err == nil && strings.HasPrefix(settings.BackgroundImage, "/assets/backgrounds/") {
		oldName := path.Base(settings.BackgroundImage)
		newName := oldName
		if mapped, ok := renames[oldName]; ok {
			newName = mapped
		}
		newPath := "/assets/wallpapers/" + newName
		if _, err := settingsSvc.SetBackgroundImage(newPath); err != nil {
			log.Printf("migrate background_image setting: %v", err)
		} else {
			log.Printf("migrated background_image to %s", newPath)
		}
	} else {
		rewriteBackgroundImagePath(settingsSvc)
	}

	if left, err := os.ReadDir(oldDir); err == nil && len(left) == 0 {
		_ = os.Remove(oldDir)
	}
	if moved > 0 {
		log.Printf("migrated %d background asset(s) into wallpapers library", moved)
	}
}

func rewriteBackgroundImagePath(settingsSvc *service.SettingsService) {
	settings, err := settingsSvc.Get()
	if err != nil || !strings.HasPrefix(settings.BackgroundImage, "/assets/backgrounds/") {
		return
	}
	newPath := "/assets/wallpapers/" + path.Base(settings.BackgroundImage)
	if _, err := settingsSvc.SetBackgroundImage(newPath); err != nil {
		log.Printf("migrate background_image setting: %v", err)
		return
	}
	log.Printf("migrated background_image to %s", newPath)
}

// migrateSiteAssetsToIcons folds legacy /assets/site into the icons gallery library.
func migrateSiteAssetsToIcons(dataDir string, settingsSvc *service.SettingsService) {
	siteDir := filepath.Join(dataDir, "assets", "site")
	iconsDir := filepath.Join(dataDir, "assets", "icons")
	entries, err := os.ReadDir(siteDir)
	if err != nil {
		return
	}
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		log.Printf("migrate site assets: create icons dir: %v", err)
		return
	}

	renames := map[string]string{} // old base name -> new base name
	moved := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		oldName := e.Name()
		newName := oldName
		src := filepath.Join(siteDir, oldName)
		dst := filepath.Join(iconsDir, newName)
		if _, err := os.Stat(dst); err == nil {
			ext := filepath.Ext(oldName)
			base := strings.TrimSuffix(oldName, ext)
			newName = base + "-site" + ext
			dst = filepath.Join(iconsDir, newName)
		}
		if err := os.Rename(src, dst); err != nil {
			log.Printf("migrate site asset %s: %v", oldName, err)
			continue
		}
		renames[oldName] = newName
		moved++
	}

	if settings, err := settingsSvc.Get(); err == nil && strings.HasPrefix(settings.SiteIcon, "/assets/site/") {
		oldName := path.Base(settings.SiteIcon)
		newName := oldName
		if mapped, ok := renames[oldName]; ok {
			newName = mapped
		}
		newPath := "/assets/icons/" + newName
		if _, err := settingsSvc.SetSiteIcon(newPath); err != nil {
			log.Printf("migrate site_icon setting: %v", err)
		} else {
			log.Printf("migrated site_icon to %s", newPath)
		}
	}

	if left, err := os.ReadDir(siteDir); err == nil && len(left) == 0 {
		_ = os.Remove(siteDir)
	}
	if moved > 0 {
		log.Printf("migrated %d site asset(s) into icons library", moved)
	}
}

// migrateLegacyDB moves /data/navi-dock.db into /data/database/ when upgrading.
func migrateLegacyDB(dataDir, dbPath string) {
	if _, err := os.Stat(dbPath); err == nil {
		return
	}
	legacy := filepath.Join(dataDir, "navi-dock.db")
	if _, err := os.Stat(legacy); err != nil {
		return
	}
	if err := os.Rename(legacy, dbPath); err != nil {
		log.Printf("migrate legacy db failed: %v", err)
		return
	}
	log.Printf("migrated sqlite db to %s", dbPath)
}

func mountStatic(r *gin.Engine, webDir string) {
	info, err := os.Stat(webDir)
	if err != nil || !info.IsDir() {
		log.Printf("static web dir not found: %s", webDir)
		r.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"message": "frontend not built"})
		})
		return
	}

	fileServer := http.FileServer(http.Dir(webDir))
	r.NoRoute(func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if strings.HasPrefix(reqPath, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}

		clean := path.Clean("/" + reqPath)
		full := path.Join(webDir, clean)
		if st, err := os.Stat(full); err == nil && !st.IsDir() {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		index := path.Join(webDir, "index.html")
		http.ServeFile(c.Writer, c.Request, index)
	})
}
