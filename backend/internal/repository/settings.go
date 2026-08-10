package repository

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/navi-dock/navi-dock/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.Setting{}, &model.NavGroup{}, &model.NavItem{}); err != nil {
		return nil, err
	}
	return db, nil
}

type SettingsRepo struct {
	db *gorm.DB
}

func NewSettingsRepo(db *gorm.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) GetAll() (map[string]string, error) {
	var rows []model.Setting
	if err := r.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		out[row.Key] = row.Value
	}
	return out, nil
}

func (r *SettingsRepo) Upsert(key, value string) error {
	var existing model.Setting
	err := r.db.Where("`key` = ?", key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&model.Setting{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	existing.Value = value
	return r.db.Save(&existing).Error
}
