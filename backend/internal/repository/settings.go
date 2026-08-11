package repository

import (
	"os"
	"path/filepath"
	"strings"

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
	if err := migrateAddUserIDColumns(db); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.User{}, &model.Setting{}, &model.NavGroup{}, &model.NavItem{}); err != nil {
		return nil, err
	}
	if err := assignOrphanedData(db); err != nil {
		return nil, err
	}
	if err := dropLegacySettingsKeyUnique(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateAddUserIDColumns(db *gorm.DB) error {
	type colTarget struct {
		model any
		table string
	}
	targets := []colTarget{
		{&model.Setting{}, "settings"},
		{&model.NavGroup{}, "nav_groups"},
		{&model.NavItem{}, "nav_items"},
	}
	for _, t := range targets {
		if !db.Migrator().HasTable(t.model) {
			continue
		}
		if db.Migrator().HasColumn(t.model, "user_id") {
			continue
		}
		// SQLite cannot ADD a NOT NULL column without a default.
		sql := "ALTER TABLE `" + t.table + "` ADD COLUMN `user_id` text NOT NULL DEFAULT ''"
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func dropLegacySettingsKeyUnique(db *gorm.DB) error {
	if !db.Migrator().HasTable("settings") {
		return nil
	}
	type idxRow struct {
		Name string
		SQL  string
	}
	var rows []idxRow
	if err := db.Raw(`SELECT name, sql FROM sqlite_master WHERE type = 'index' AND tbl_name = 'settings'`).Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		sql := strings.ToUpper(row.SQL)
		name := strings.ToLower(row.Name)
		if strings.Contains(sql, "UNIQUE") && strings.Contains(sql, "KEY") && !strings.Contains(sql, "USER_ID") {
			if err := db.Exec("DROP INDEX IF EXISTS `" + row.Name + "`").Error; err != nil {
				return err
			}
			continue
		}
		if name == "idx_settings_key" || name == "uni_settings_key" {
			_ = db.Exec("DROP INDEX IF EXISTS `" + row.Name + "`").Error
		}
	}
	return nil
}

// assignOrphanedData attaches pre-multi-user rows to the first account.
func assignOrphanedData(db *gorm.DB) error {
	var first model.User
	if err := db.Order("created_at asc").First(&first).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	_ = db.Exec("UPDATE nav_groups SET user_id = ? WHERE user_id IS NULL OR user_id = ''", first.ID).Error
	_ = db.Exec("UPDATE nav_items SET user_id = ? WHERE user_id IS NULL OR user_id = ''", first.ID).Error
	_ = db.Exec("UPDATE settings SET user_id = ? WHERE user_id IS NULL OR user_id = ''", first.ID).Error
	return nil
}

type SettingsRepo struct {
	db *gorm.DB
}

func NewSettingsRepo(db *gorm.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) GetAll(userID string) (map[string]string, error) {
	var rows []model.Setting
	if err := r.db.Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		out[row.Key] = row.Value
	}
	return out, nil
}

func (r *SettingsRepo) Upsert(userID, key, value string) error {
	var existing model.Setting
	err := r.db.Where("user_id = ? AND `key` = ?", userID, key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&model.Setting{UserID: userID, Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	existing.Value = value
	return r.db.Save(&existing).Error
}

func (r *SettingsRepo) ListByKey(key string) ([]model.Setting, error) {
	var rows []model.Setting
	err := r.db.Where("`key` = ?", key).Find(&rows).Error
	return rows, err
}

func (r *SettingsRepo) DeleteByUser(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.Setting{}).Error
}
