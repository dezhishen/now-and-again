package repository

import (
	"errors"

	"gorm.io/gorm"
)

// ─── UserRepo Methods ─────────────────────────────────────────────

// Create inserts a new user. Returns error on duplicate username/email.
func (r *UserRepo) Create(user *UserModel) error {
	return r.db.Create(user).Error
}

// FindByUsername looks up a user by username.
func (r *UserRepo) FindByUsername(username string) (*UserModel, error) {
	var u UserModel
	err := r.db.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// FindByEmail looks up a user by email.
func (r *UserRepo) FindByEmail(email string) (*UserModel, error) {
	var u UserModel
	err := r.db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// FindByID looks up a user by ID.
func (r *UserRepo) FindByID(id string) (*UserModel, error) {
	var u UserModel
	err := r.db.Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// Count returns the total number of users (used to check initialization).
func (r *UserRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Count(&count).Error
	return count, err
}

// Update saves changes to an existing user.
func (r *UserRepo) Update(user *UserModel) error {
	return r.db.Save(user).Error
}

// ListAll returns all users (admin only).
func (r *UserRepo) ListAll() ([]UserModel, error) {
	var users []UserModel
	err := r.db.Order("created_at ASC").Find(&users).Error
	return users, err
}
