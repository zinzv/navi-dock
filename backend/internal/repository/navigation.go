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
