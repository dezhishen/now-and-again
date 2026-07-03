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
		return nil, ignoreNotFound(err)
	}
	return &u, nil
}

func (r *UserRepo) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Count(&count).Error
	return count, err
}

// SearchUsers searches users by display_name, email, or account username with pagination.
func (r *UserRepo) SearchUsers(query string, page, pageSize int) ([]UserModel, int64, error) {
	var users []UserModel
	var total int64

	db := r.db.Model(&UserModel{}).Preload("Roles.Role")

	if query != "" {
		like := "%" + query + "%"
		// Use GORM subquery: no raw table names
		accountSub := r.db.Model(&AccountModel{}).Select("user_id").
			Where("username LIKE ? AND provider = ?", like, "local")
		db = db.Where(
			"display_name LIKE ? OR email LIKE ? OR id IN (?)",
			like, like, accountSub,
		)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("created_at ASC").Find(&users).Error
	return users, total, err
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
		return nil, ignoreNotFound(err)
	}
	return &a, nil
}

func (r *UserRepo) FindAccountByUserID(userID string) (*AccountModel, error) {
	var a AccountModel
	err := r.db.Where("user_id = ? AND provider = ?", userID, "local").First(&a).Error
	if err != nil {
		return nil, ignoreNotFound(err)
	}
	return &a, nil
}

func (r *UserRepo) UpdateAccount(acc *AccountModel) error {
	return r.db.Save(acc).Error
}

// ─── Role ─────────────────────────────────────────────────────────

func (r *UserRepo) FindRoleByName(name string) (*RoleModel, error) {
	var role RoleModel
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, ignoreNotFound(err)
}

func (r *UserRepo) AddUserRole(userID, roleID string) error {
	ur := UserRoleModel{UserID: userID, RoleID: roleID}
	return r.db.Where(ur).FirstOrCreate(&ur).Error
}

// HasRole checks whether a user has a specific role by role name.
func (r *UserRepo) HasRole(userID, roleName string) (bool, error) {
	var count int64
	err := r.db.Model(&UserRoleModel{}).
		Joins("Role").
		Where(&UserRoleModel{UserID: userID}).
		Where("Role.name = ?", roleName).
		Count(&count).Error
	return count > 0, err
}

// ─── Transaction ──────────────────────────────────────────────────

func (r *UserRepo) Tx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}
