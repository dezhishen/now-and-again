package repository

import (
	"gorm.io/gorm"
)

// ─── User CRUD ────────────────────────────────────────────────────

func (r *UserRepo) CreateUser(user *UserModel) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) FindUserByID(id string) (*UserModel, error) {
	var u UserModel
	err := r.db.Preload("Roles.Role").Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Count(&count).Error
	return count, err
}

func (r *UserRepo) ListUsers() ([]UserModel, error) {
	var users []UserModel
	err := r.db.Preload("Roles.Role").Order("created_at ASC").Find(&users).Error
	return users, err
}

func (r *UserRepo) UpdateUser(user *UserModel) error {
	return r.db.Save(user).Error
}

// ─── Account CRUD ─────────────────────────────────────────────────

func (r *UserRepo) CreateAccount(acc *AccountModel) error {
	return r.db.Create(acc).Error
}

func (r *UserRepo) FindAccountByUsername(username string) (*AccountModel, error) {
	var a AccountModel
	err := r.db.Where("username = ? AND provider = ?", username, "local").First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *UserRepo) FindAccountByUserID(userID string) (*AccountModel, error) {
	var a AccountModel
	err := r.db.Where("user_id = ? AND provider = ?", userID, "local").First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ─── Role ─────────────────────────────────────────────────────────

func (r *UserRepo) FindRoleByName(name string) (*RoleModel, error) {
	var role RoleModel
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *UserRepo) AddUserRole(userID, roleID string) error {
	ur := UserRoleModel{UserID: userID, RoleID: roleID}
	return r.db.Where(ur).FirstOrCreate(&ur).Error
}

// ─── Transaction ──────────────────────────────────────────────────

func (r *UserRepo) Tx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
