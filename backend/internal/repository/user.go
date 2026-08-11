package repository

import (
	"github.com/navi-dock/navi-dock/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Count() (int64, error) {
	var n int64
	err := r.db.Model(&model.User{}).Count(&n).Error
	return n, err
}

func (r *UserRepo) List() ([]model.User, error) {
	var users []model.User
	err := r.db.Order("created_at asc").Find(&users).Error
	return users, err
}

func (r *UserRepo) FindByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) UpdatePassword(id, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

func (r *UserRepo) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

func (r *UserRepo) CountAdmins() (int64, error) {
	var n int64
	err := r.db.Model(&model.User{}).Where("role = ?", "admin").Count(&n).Error
	return n, err
}
