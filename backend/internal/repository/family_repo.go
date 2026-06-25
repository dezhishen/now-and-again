package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ─── FamilyRepo ───────────────────────────────────────────────────

func (r *FamilyRepo) Create(family *FamilyModel) error {
	family.InviteCode = generateInviteCode()
	return r.db.Create(family).Error
}

func (r *FamilyRepo) FindByID(id string) (*FamilyModel, error) {
	var f FamilyModel
	err := r.db.Where("id = ?", id).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &f, err
}

func (r *FamilyRepo) FindByInviteCode(code string) (*FamilyModel, error) {
	var f FamilyModel
	err := r.db.Where("invite_code = ?", code).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &f, err
}

// ─── FamilyMember ─────────────────────────────────────────────────

func (r *FamilyRepo) AddMember(familyID, userID, role string) error {
	m := &FamilyMemberModel{
		FamilyID: familyID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}
	return r.db.Create(m).Error
}

func (r *FamilyRepo) FindMember(familyID, userID string) (*FamilyMemberModel, error) {
	var m FamilyMemberModel
	err := r.db.Where("family_id = ? AND user_id = ?", familyID, userID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &m, err
}

func (r *FamilyRepo) ListMembers(familyID string) ([]FamilyMemberModel, error) {
	var members []FamilyMemberModel
	err := r.db.Where("family_id = ?", familyID).Preload("User").Find(&members).Error
	return members, err
}

func (r *FamilyRepo) ListFamiliesByUser(userID string) ([]FamilyModel, error) {
	var memberIDs []string
	r.db.Model(&FamilyMemberModel{}).Where("user_id = ?", userID).Pluck("family_id", &memberIDs)
	if len(memberIDs) == 0 {
		return nil, nil
	}
	var families []FamilyModel
	err := r.db.Where("id IN ?", memberIDs).Find(&families).Error
	return families, err
}

func (r *FamilyRepo) UpdateMemberRole(familyID, userID, role string) error {
	return r.db.Model(&FamilyMemberModel{}).
		Where("family_id = ? AND user_id = ?", familyID, userID).
		Update("role", role).Error
}

func (r *FamilyRepo) RemoveMember(familyID, userID string) error {
	return r.db.Where("family_id = ? AND user_id = ?", familyID, userID).
		Delete(&FamilyMemberModel{}).Error
}

func generateInviteCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
