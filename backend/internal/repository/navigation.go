package repository

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/navi-dock/navi-dock/internal/model"
	"gorm.io/gorm"
)

type NavigationRepo struct {
	db *gorm.DB
}

func NewNavigationRepo(db *gorm.DB) *NavigationRepo {
	return &NavigationRepo{db: db}
}

func (r *NavigationRepo) SeedIfEmpty() error {
	var count int64
	if err := r.db.Model(&model.NavGroup{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	groups := []model.NavGroup{
		{ID: "app", Name: "APP", Icon: "", Sort: 0, CreatedAt: now, UpdatedAt: now},
		{ID: "ai", Name: "AI", Icon: "", Sort: 1, CreatedAt: now, UpdatedAt: now},
	}
	items := []model.NavItem{
		{ID: "nas", GroupID: "app", Name: "NAS", Icon: "mdi:nas", ExternalURL: "https://nas.example.com", InternalURL: "http://192.168.1.10:5000", Sort: 0, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "kavita", GroupID: "app", Name: "Kavita", Icon: "mdi:book-open-page-variant-outline", ExternalURL: "https://kavita.example.com", InternalURL: "http://192.168.1.20:5001", Sort: 1, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "plex", GroupID: "app", Name: "Plex", Icon: "mdi:plex", ExternalURL: "https://plex.example.com", InternalURL: "http://192.168.1.10:32400", Sort: 2, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "blog", GroupID: "app", Name: "博客", Icon: "mdi:post-outline", ExternalURL: "https://blog.example.com", InternalURL: "http://192.168.1.30:8080", Sort: 3, Status: "unknown", CreatedAt: now, UpdatedAt: now},
		{ID: "photos", GroupID: "app", Name: "图片", Icon: "mdi:image-outline", ExternalURL: "https://photos.example.com", InternalURL: "http://192.168.1.40:2283", Sort: 4, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "memos", GroupID: "app", Name: "Memos", Icon: "mdi:note-text-outline", ExternalURL: "https://memos.example.com", InternalURL: "http://192.168.1.50:5230", Sort: 5, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "bookverse", GroupID: "app", Name: "书元", Icon: "mdi:bookshelf", ExternalURL: "https://books.example.com", InternalURL: "http://192.168.1.60:7525", Sort: 6, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "audio", GroupID: "app", Name: "音元", Icon: "mdi:music-note-outline", ExternalURL: "https://music.example.com", InternalURL: "http://192.168.1.70:7526", Sort: 7, Status: "offline", CreatedAt: now, UpdatedAt: now},
		{ID: "gemini", GroupID: "ai", Name: "Gemini", Icon: "mdi:star-four-points-outline", ExternalURL: "https://gemini.google.com", InternalURL: "", Sort: 0, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "deepseek", GroupID: "ai", Name: "DeepSeek", Icon: "mdi:fish", ExternalURL: "https://chat.deepseek.com", InternalURL: "", Sort: 1, Status: "online", CreatedAt: now, UpdatedAt: now},
		{ID: "chatgpt", GroupID: "ai", Name: "ChatGPT", Icon: "simple-icons:openai", ExternalURL: "https://chatgpt.com", InternalURL: "", Sort: 2, Status: "online", CreatedAt: now, UpdatedAt: now},
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&groups).Error; err != nil {
			return err
		}
		return tx.Create(&items).Error
	})
}

func (r *NavigationRepo) ListGroupsWithItems() ([]model.NavGroup, error) {
	var groups []model.NavGroup
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC, created_at ASC")
	}).Order("sort ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

func (r *NavigationRepo) ListGroups() ([]model.NavGroup, error) {
	var groups []model.NavGroup
	err := r.db.Order("sort ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

func (r *NavigationRepo) GetGroup(id string) (*model.NavGroup, error) {
	var group model.NavGroup
	err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC, created_at ASC")
	}).First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *NavigationRepo) CreateGroup(name, icon string) (*model.NavGroup, error) {
	var maxSort int
	_ = r.db.Model(&model.NavGroup{}).Select("COALESCE(MAX(sort), -1)").Scan(&maxSort)
	group := model.NavGroup{
		ID:   newID(),
		Name: name,
		Icon: icon,
		Sort: maxSort + 1,
	}
	if err := r.db.Create(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *NavigationRepo) UpdateGroup(id, name, icon string) (*model.NavGroup, error) {
	var group model.NavGroup
	if err := r.db.First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}
	group.Name = name
	group.Icon = icon
	if err := r.db.Save(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *NavigationRepo) DeleteGroup(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&model.NavItem{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.NavGroup{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *NavigationRepo) SortGroups(ids []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			res := tx.Model(&model.NavGroup{}).Where("id = ?", id).Update("sort", i)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("group not found: %s", id)
			}
		}
		return nil
	})
}

func (r *NavigationRepo) CountItems(groupID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.NavItem{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

func (r *NavigationRepo) GetItem(id string) (*model.NavItem, error) {
	var item model.NavItem
	if err := r.db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *NavigationRepo) CreateItem(item *model.NavItem) error {
	var maxSort int
	_ = r.db.Model(&model.NavItem{}).
		Where("group_id = ?", item.GroupID).
		Select("COALESCE(MAX(sort), -1)").
		Scan(&maxSort)
	item.Sort = maxSort + 1
	if item.ID == "" {
		item.ID = newID()
	}
	if item.Status == "" {
		item.Status = "online"
	}
	return r.db.Create(item).Error
}

func (r *NavigationRepo) UpdateItem(item *model.NavItem) error {
	return r.db.Save(item).Error
}

func (r *NavigationRepo) DeleteItem(id string) error {
	res := r.db.Delete(&model.NavItem{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *NavigationRepo) SortItems(groupID string, ids []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			res := tx.Model(&model.NavItem{}).
				Where("id = ? AND group_id = ?", id, groupID).
				Update("sort", i)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("item not found in group: %s", id)
			}
		}
		return nil
	})
}

func (r *NavigationRepo) GroupExists(id string) (bool, error) {
	var count int64
	err := r.db.Model(&model.NavGroup{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *NavigationRepo) ReplaceAll(groups []model.NavGroup, items []model.NavItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.NavItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&model.NavGroup{}).Error; err != nil {
			return err
		}
		if len(groups) > 0 {
			if err := tx.Create(&groups).Error; err != nil {
				return err
			}
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// NewID exposes ID generation for import helpers.
func NewID() string {
	return newID()
}
