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

func (r *FamilyRepo) UpdateFamily(f *FamilyModel) error {
	return r.db.Save(f).Error
}

func (r *FamilyRepo) DeleteFamily(id string) error {
	return r.db.Model(&FamilyModel{}).Where("id = ?", id).Update("archived", true).Error
}

func (r *FamilyRepo) RestoreFamily(id string) error {
	return r.db.Model(&FamilyModel{}).Where("id = ?", id).Update("archived", false).Error
}

func (r *FamilyRepo) FindFamilyByInviteCode(code string) (*FamilyModel, error) {
	var f FamilyModel
	err := r.db.Where("invite_code = ?", code).First(&f).Error
	return &f, err
}

func (r *FamilyRepo) FindFamilyByCreator(userID string) (*FamilyModel, error) {
	var f FamilyModel
	err := r.db.Where("created_by = ?", userID).First(&f).Error
	return &f, err
}

func (r *FamilyRepo) ListFamiliesByUserID(userID string) ([]FamilyModel, error) {
	var families []FamilyModel

	familyIDSub := r.db.Model(&FamilyMemberModel{}).Select("family_id").
		Where("user_id = ? AND status = ?", userID, "active")

	err := r.db.
		Preload("CoverImage").
		Where("id IN (?)", familyIDSub).
		Find(&families).Error
	return families, err
}

// SetFamilyCoverImage syncs the family's cover_image_id column.
func (r *FamilyRepo) SetFamilyCoverImage(familyID, imageID string) error {
	return r.db.Model(&FamilyModel{}).Where("id = ?", familyID).
		Update("cover_image_id", imageID).Error
}

// ClearFamilyCoverImage removes the cover image reference from a family.
func (r *FamilyRepo) ClearFamilyCoverImage(familyID string) error {
	return r.db.Model(&FamilyModel{}).Where("id = ?", familyID).
		Update("cover_image_id", nil).Error
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

func (r *FamilyRepo) ListGroups(familyID, name string) ([]FamilyGroupModel, error) {
	var groups []FamilyGroupModel
	q := r.db.Where("family_id = ?", familyID)
	if name != "" {
		q = q.Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%")
	}
	err := q.Find(&groups).Error
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

// ListUserGroupIDs returns the IDs of all groups the user is an active member of in the given family.
func (r *FamilyRepo) ListUserGroupIDs(userID, familyID string) ([]string, error) {
	var ids []string
	familyGroupSub := r.db.Model(&FamilyGroupModel{}).Select("id").Where("family_id = ?", familyID)
	err := r.db.Model(&FamilyGroupMemberModel{}).
		Where("user_id = ? AND status = ? AND group_id IN (?)", userID, "active", familyGroupSub).
		Pluck("group_id", &ids).Error
	return ids, err
}

// ─── Validation ──────────────────────────────────────────────────

func (r *FamilyRepo) ValidateMembership(userID, familyID string) error {
	var count int64
	err := r.db.Model(&FamilyMemberModel{}).
		Where("user_id = ? AND family_id = ? AND status = ?", userID, familyID, "active").
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("not a member of this family")
	}
	return nil
}
