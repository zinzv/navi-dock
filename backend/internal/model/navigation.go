package model

import "time"

type NavGroup struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	UserID    string    `gorm:"size:36;not null;default:'';index" json:"user_id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Icon      string    `gorm:"size:128" json:"icon"`
	Sort      int       `gorm:"not null;default:0;index" json:"sort"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Items     []NavItem `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	ItemCount int64     `gorm:"-" json:"item_count,omitempty"`
}

type NavItem struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	UserID      string    `gorm:"size:36;not null;default:'';index" json:"user_id"`
	GroupID     string    `gorm:"size:36;not null;index" json:"group_id"`
	Name        string    `gorm:"size:64;not null" json:"name"`
	Description string    `gorm:"size:200" json:"description"`
	IconType    string    `gorm:"size:32;default:iconify" json:"icon_type"` // iconify | text | image
	Icon        string    `gorm:"size:512" json:"icon"`
	InternalURL string    `gorm:"size:512" json:"internal_url"`
	ExternalURL string    `gorm:"size:512" json:"external_url"`
	OpenType    string    `gorm:"size:16;default:_blank" json:"open_type"` // _blank | _self
	Sort        int       `gorm:"not null;default:0;index" json:"sort"`
	Status      string    `gorm:"size:32;default:online" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateGroupInput struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type UpdateGroupInput struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type SortGroupsInput struct {
	IDs []string `json:"ids"`
}

type SortItemsInput struct {
	GroupID string   `json:"group_id"`
	IDs     []string `json:"ids"`
}

type UpsertItemInput struct {
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconType    string `json:"icon_type"`
	Icon        string `json:"icon"`
	InternalURL string `json:"internal_url"`
	ExternalURL string `json:"external_url"`
	OpenType    string `json:"open_type"`
}

// SunPanelExport is the JSON shape exported by Sun-Panel.
type SunPanelExport struct {
	Version    int              `json:"version"`
	AppName    string           `json:"appName"`
	ExportTime string           `json:"exportTime"`
	Icons      []SunPanelGroup  `json:"icons"`
}

type SunPanelGroup struct {
	Title    string          `json:"title"`
	Sort     int             `json:"sort"`
	Children []SunPanelChild `json:"children"`
}

type SunPanelChild struct {
	Title       string        `json:"title"`
	Sort        int           `json:"sort"`
	URL         string        `json:"url"`
	LanURL      string        `json:"lanUrl"`
	Description string        `json:"description"`
	OpenMethod  int           `json:"openMethod"`
	Icon        SunPanelIcon  `json:"icon"`
}

type SunPanelIcon struct {
	ItemType int    `json:"itemType"`
	Src      string `json:"src"`
	Text     string `json:"text"`
}

// NaviDockExport is the native backup format for this app.
type NaviDockExport struct {
	App        string              `json:"app"`
	Version    int                 `json:"version"`
	ExportedAt string              `json:"exported_at"`
	Groups     []NaviDockExportGroup `json:"groups"`
}

type NaviDockExportGroup struct {
	ID    string               `json:"id"`
	Name  string               `json:"name"`
	Icon  string               `json:"icon,omitempty"`
	Sort  int                  `json:"sort"`
	Items []NaviDockExportItem `json:"items"`
}

type NaviDockExportItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconType    string `json:"icon_type,omitempty"`
	Icon        string `json:"icon,omitempty"`
	ExternalURL string `json:"external_url,omitempty"`
	InternalURL string `json:"internal_url,omitempty"`
	OpenType    string `json:"open_type,omitempty"`
	Sort        int    `json:"sort"`
	Status      string `json:"status,omitempty"`
}
