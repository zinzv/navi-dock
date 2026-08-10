package model

import "time"

type Setting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:64;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AppSettings struct {
	SiteTitle         string  `json:"site_title"`
	SiteIcon          string  `json:"site_icon"` // /assets/icons/xxx.png
	Language          string  `json:"language"`  // zh | en
	Theme             string  `json:"theme"`     // light | dark | system
	NetworkMode       string  `json:"network_mode"` // auto | internal | external
	BackgroundImage   string  `json:"background_image"`   // /assets/wallpapers/xxx.png
	BackgroundOpacity float64 `json:"background_opacity"` // 0-1
	ClearBackground   bool    `json:"clear_background,omitempty"`
	ClearSiteIcon     bool    `json:"clear_site_icon,omitempty"`
}

func DefaultSettings() AppSettings {
	return AppSettings{
		SiteTitle:         "导航",
		SiteIcon:          "",
		Language:          "zh",
		Theme:             "dark",
		NetworkMode:       "auto",
		BackgroundImage:   "",
		BackgroundOpacity: 0.35,
	}
}
