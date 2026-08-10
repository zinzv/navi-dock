package service

import (
	"strconv"

	"github.com/navi-dock/navi-dock/internal/model"
	"github.com/navi-dock/navi-dock/internal/repository"
)

type SettingsService struct {
	repo *repository.SettingsRepo
}

func NewSettingsService(repo *repository.SettingsRepo) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) Get() (model.AppSettings, error) {
	defaults := model.DefaultSettings()
	values, err := s.repo.GetAll()
	if err != nil {
		return defaults, err
	}
	if v, ok := values["site_title"]; ok && v != "" {
		defaults.SiteTitle = v
	}
	if v, ok := values["site_icon"]; ok {
		defaults.SiteIcon = v
	}
	if v, ok := values["language"]; ok && (v == "zh" || v == "en") {
		defaults.Language = v
	}
	if v, ok := values["theme"]; ok && (v == "light" || v == "dark" || v == "system") {
		defaults.Theme = v
	}
	if v, ok := values["network_mode"]; ok && (v == "auto" || v == "internal" || v == "external") {
		defaults.NetworkMode = v
	}
	if v, ok := values["background_image"]; ok {
		defaults.BackgroundImage = v
	}
	if v, ok := values["background_opacity"]; ok && v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			defaults.BackgroundOpacity = clampOpacity(f)
		}
	}
	return defaults, nil
}

func (s *SettingsService) Update(in model.AppSettings) (model.AppSettings, error) {
	current, err := s.Get()
	if err != nil {
		return current, err
	}
	if in.SiteTitle != "" {
		current.SiteTitle = in.SiteTitle
	}
	if in.ClearSiteIcon {
		current.SiteIcon = ""
	} else if in.SiteIcon != "" {
		current.SiteIcon = in.SiteIcon
	}
	if in.Language == "zh" || in.Language == "en" {
		current.Language = in.Language
	}
	if in.Theme == "light" || in.Theme == "dark" || in.Theme == "system" {
		current.Theme = in.Theme
	}
	if in.NetworkMode == "auto" || in.NetworkMode == "internal" || in.NetworkMode == "external" {
		current.NetworkMode = in.NetworkMode
	}
	if in.ClearBackground {
		current.BackgroundImage = ""
	} else if in.BackgroundImage != "" {
		current.BackgroundImage = in.BackgroundImage
	}
	current.BackgroundOpacity = clampOpacity(in.BackgroundOpacity)

	pairs := map[string]string{
		"site_title":         current.SiteTitle,
		"site_icon":          current.SiteIcon,
		"language":           current.Language,
		"theme":              current.Theme,
		"network_mode":       current.NetworkMode,
		"background_image":   current.BackgroundImage,
		"background_opacity": strconv.FormatFloat(current.BackgroundOpacity, 'f', 2, 64),
	}
	for k, v := range pairs {
		if err := s.repo.Upsert(k, v); err != nil {
			return current, err
		}
	}
	return current, nil
}

func (s *SettingsService) SetBackgroundImage(path string) (model.AppSettings, error) {
	current, err := s.Get()
	if err != nil {
		return current, err
	}
	current.BackgroundImage = path
	if err := s.repo.Upsert("background_image", path); err != nil {
		return current, err
	}
	return current, nil
}

func (s *SettingsService) SetSiteIcon(path string) (model.AppSettings, error) {
	current, err := s.Get()
	if err != nil {
		return current, err
	}
	current.SiteIcon = path
	if err := s.repo.Upsert("site_icon", path); err != nil {
		return current, err
	}
	return current, nil
}

func clampOpacity(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
