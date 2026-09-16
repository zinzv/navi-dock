package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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

func (r *NavigationRepo) ListGroupsWithItems(userID string) ([]model.NavGroup, error) {
	var groups []model.NavGroup
	err := r.db.Where("user_id = ?", userID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("user_id = ?", userID).Order("sort ASC, created_at ASC")
		}).
		Order("sort ASC, created_at ASC").
		Find(&groups).Error
	return groups, err
}

func (r *NavigationRepo) ListGroups(userID string) ([]model.NavGroup, error) {
	var groups []model.NavGroup
	err := r.db.Where("user_id = ?", userID).Order("sort ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

func (r *NavigationRepo) GetGroup(userID, id string) (*model.NavGroup, error) {
	var group model.NavGroup
	err := r.db.Where("user_id = ? AND id = ?", userID, id).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("user_id = ?", userID).Order("sort ASC, created_at ASC")
		}).
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *NavigationRepo) CreateGroup(userID, name, icon string) (*model.NavGroup, error) {
	var maxSort int
	_ = r.db.Model(&model.NavGroup{}).Where("user_id = ?", userID).
		Select("COALESCE(MAX(sort), -1)").Scan(&maxSort)
	group := model.NavGroup{
		ID:     newID(),
		UserID: userID,
		Name:   name,
		Icon:   icon,
		Sort:   maxSort + 1,
	}
	if err := r.db.Create(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *NavigationRepo) UpdateGroup(userID, id, name, icon string) (*model.NavGroup, error) {
	var group model.NavGroup
	if err := r.db.Where("user_id = ? AND id = ?", userID, id).First(&group).Error; err != nil {
		return nil, err
	}
	group.Name = name
	group.Icon = icon
	if err := r.db.Save(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *NavigationRepo) DeleteGroup(userID, id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND group_id = ?", userID, id).Delete(&model.NavItem{}).Error; err != nil {
			return err
		}
		res := tx.Where("user_id = ? AND id = ?", userID, id).Delete(&model.NavGroup{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *NavigationRepo) SortGroups(userID string, ids []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			res := tx.Model(&model.NavGroup{}).Where("user_id = ? AND id = ?", userID, id).Update("sort", i)
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

func (r *NavigationRepo) CountItems(userID, groupID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.NavItem{}).Where("user_id = ? AND group_id = ?", userID, groupID).Count(&count).Error
	return count, err
}

func (r *NavigationRepo) GetItem(userID, id string) (*model.NavItem, error) {
	var item model.NavItem
	if err := r.db.Where("user_id = ? AND id = ?", userID, id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *NavigationRepo) CreateItem(item *model.NavItem) error {
	var maxSort int
	_ = r.db.Model(&model.NavItem{}).
		Where("user_id = ? AND group_id = ?", item.UserID, item.GroupID).
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

func (r *NavigationRepo) DeleteItem(userID, id string) error {
	res := r.db.Where("user_id = ? AND id = ?", userID, id).Delete(&model.NavItem{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *NavigationRepo) SortItems(userID, groupID string, ids []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		listed := make(map[string]struct{}, len(ids))
		sourceGroups := map[string]struct{}{}

		for i, id := range ids {
			var current model.NavItem
			err := tx.Select("id", "group_id").
				Where("user_id = ? AND id = ?", userID, id).
				First(&current).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("item not found: %s", id)
			}
			if err != nil {
				return err
			}
			if current.GroupID != groupID {
				sourceGroups[current.GroupID] = struct{}{}
			}
			listed[id] = struct{}{}
			res := tx.Model(&model.NavItem{}).
				Where("user_id = ? AND id = ?", userID, id).
				Updates(map[string]interface{}{"group_id": groupID, "sort": i})
			if res.Error != nil {
				return res.Error
			}
		}

		var extras []model.NavItem
		if err := tx.Where("user_id = ? AND group_id = ?", userID, groupID).
			Order("sort ASC, created_at ASC").
			Find(&extras).Error; err != nil {
			return err
		}
		next := len(ids)
		for i := range extras {
			if _, ok := listed[extras[i].ID]; ok {
				continue
			}
			if err := tx.Model(&extras[i]).Update("sort", next).Error; err != nil {
				return err
			}
			next++
		}

		for srcID := range sourceGroups {
			var remaining []model.NavItem
			if err := tx.Where("user_id = ? AND group_id = ?", userID, srcID).
				Order("sort ASC, created_at ASC").
				Find(&remaining).Error; err != nil {
				return err
			}
			for i := range remaining {
				if err := tx.Model(&remaining[i]).Update("sort", i).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *NavigationRepo) GroupExists(userID, id string) (bool, error) {
	var count int64
	err := r.db.Model(&model.NavGroup{}).Where("user_id = ? AND id = ?", userID, id).Count(&count).Error
	return count > 0, err
}

func (r *NavigationRepo) ReplaceAll(userID string, groups []model.NavGroup, items []model.NavItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.NavItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.NavGroup{}).Error; err != nil {
			return err
		}
		for i := range groups {
			groups[i].UserID = userID
		}
		for i := range items {
			items[i].UserID = userID
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

func (r *NavigationRepo) DeleteByUser(userID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.NavItem{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&model.NavGroup{}).Error
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
