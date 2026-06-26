package repository

import "fmt"

// ─── Family CRUD ──────────────────────────────────────────────────

func (r *FamilyRepo) CreateFamily(f *FamilyModel) error {
	return r.db.Create(f).Error
}

func (r *FamilyRepo) FindFamilyByID(id string) (*FamilyModel, error) {
	var f FamilyModel
	err := r.db.Where("id = ?", id).First(&f).Error
	return &f, err
}

func (r *FamilyRepo) FindFamilyByInviteCode(code string) (*FamilyModel, error) {
	var f FamilyModel
	err := r.db.Where("invite_code = ?", code).First(&f).Error
	return &f, err
}

func (r *FamilyRepo) ListFamiliesByUserID(userID string) ([]FamilyModel, error) {
	var families []FamilyModel
	err := r.db.
		Joins("JOIN family_members ON family_members.family_id = families.id").
		Where("family_members.user_id = ? AND family_members.status = ?", userID, "active").
		Find(&families).Error
	return families, err
}

// ─── Family Member ────────────────────────────────────────────────

func (r *FamilyRepo) AddMember(m *FamilyMemberModel) error {
	return r.db.Create(m).Error
}

func (r *FamilyRepo) FindMember(familyID, userID string) (*FamilyMemberModel, error) {
	var m FamilyMemberModel
	err := r.db.Where("family_id = ? AND user_id = ?", familyID, userID).First(&m).Error
	return &m, err
}

func (r *FamilyRepo) ListMembers(familyID string) ([]FamilyMemberModel, error) {
	var members []FamilyMemberModel
	err := r.db.Preload("User").Where("family_id = ? AND status = ?", familyID, "active").Find(&members).Error
	return members, err
}

func (r *FamilyRepo) ListMembersByStatus(familyID, status string) ([]FamilyMemberModel, error) {
	var members []FamilyMemberModel
	err := r.db.Preload("User").Where("family_id = ? AND status = ?", familyID, status).Find(&members).Error
	return members, err
}

func (r *FamilyRepo) UpdateMember(m *FamilyMemberModel) error {
	return r.db.Save(m).Error
}

func (r *FamilyRepo) DeleteMember(familyID, userID string) error {
	return r.db.Where("family_id = ? AND user_id = ?", familyID, userID).Delete(&FamilyMemberModel{}).Error
}

// ─── Family Group ─────────────────────────────────────────────────

func (r *FamilyRepo) CreateGroup(g *FamilyGroupModel) error {
	return r.db.Create(g).Error
}

func (r *FamilyRepo) FindGroupByID(id string) (*FamilyGroupModel, error) {
	var g FamilyGroupModel
	err := r.db.Where("id = ?", id).First(&g).Error
	return &g, err
}

func (r *FamilyRepo) ListGroups(familyID string) ([]FamilyGroupModel, error) {
	var groups []FamilyGroupModel
	err := r.db.Where("family_id = ?", familyID).Find(&groups).Error
	return groups, err
}

func (r *FamilyRepo) DeleteGroup(id string) error {
	return r.db.Where("id = ?", id).Delete(&FamilyGroupModel{}).Error
}

// ─── Family Group Member ──────────────────────────────────────────

func (r *FamilyRepo) AddGroupMember(m *FamilyGroupMemberModel) error {
	return r.db.Create(m).Error
}

func (r *FamilyRepo) FindGroupMember(groupID, userID string) (*FamilyGroupMemberModel, error) {
	var m FamilyGroupMemberModel
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}
	return &m, nil
}

func (r *FamilyRepo) ListGroupMembers(groupID string) ([]FamilyGroupMemberModel, error) {
	var members []FamilyGroupMemberModel
	err := r.db.Preload("User").Where("group_id = ? AND status = ?", groupID, "active").Find(&members).Error
	return members, err
}

func (r *FamilyRepo) ListGroupMembersByStatus(groupID, status string) ([]FamilyGroupMemberModel, error) {
	var members []FamilyGroupMemberModel
	err := r.db.Preload("User").Where("group_id = ? AND status = ?", groupID, status).Find(&members).Error
	return members, err
}

func (r *FamilyRepo) UpdateGroupMember(m *FamilyGroupMemberModel) error {
	return r.db.Save(m).Error
}

func (r *FamilyRepo) DeleteGroupMember(groupID, userID string) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&FamilyGroupMemberModel{}).Error
}
