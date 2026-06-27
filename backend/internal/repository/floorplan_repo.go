package repository

// ─── Floor Plan ──────────────────────────────────────────────────

func (r *FloorPlanRepo) CreateFloorPlan(fp *FloorPlanModel) error {
	return r.db.Create(fp).Error
}

func (r *FloorPlanRepo) FindFloorPlanByID(id string) (*FloorPlanModel, error) {
	var fp FloorPlanModel
	err := r.db.Preload("Image").Where("id = ?", id).First(&fp).Error
	return &fp, err
}

func (r *FloorPlanRepo) ListFloorPlansByFamilyID(familyID string) ([]FloorPlanModel, error) {
	var plans []FloorPlanModel
	err := r.db.Preload("Image").Where("family_id = ?", familyID).Order("created_at ASC").Find(&plans).Error
	return plans, err
}

func (r *FloorPlanRepo) DeleteFloorPlanByID(id string) error {
	return r.db.Where("id = ?", id).Delete(&FloorPlanModel{}).Error
}

func (r *FloorPlanRepo) ClearCoverForFamily(familyID string) error {
	return r.db.Model(&FloorPlanModel{}).Where("family_id = ?", familyID).Update("is_cover", false).Error
}

func (r *FloorPlanRepo) SetCover(id string) error {
	return r.db.Model(&FloorPlanModel{}).Where("id = ?", id).Update("is_cover", true).Error
}

// ─── Location ──────────────────────────────────────────────────

func (r *FloorPlanRepo) CreateLocation(loc *LocationModel) error {
	return r.db.Create(loc).Error
}

func (r *FloorPlanRepo) ListLocationsByFloorPlanID(floorPlanID string) ([]LocationModel, error) {
	var locs []LocationModel
	err := r.db.Where("floor_plan_id = ?", floorPlanID).Order("created_at ASC").Find(&locs).Error
	return locs, err
}

func (r *FloorPlanRepo) ListLocationsByFamilyID(familyID string) ([]LocationModel, error) {
	var locs []LocationModel
	err := r.db.Where("family_id = ?", familyID).Order("created_at ASC").Find(&locs).Error
	return locs, err
}

func (r *FloorPlanRepo) FindLocationByID(id string) (*LocationModel, error) {
	var loc LocationModel
	err := r.db.Where("id = ?", id).First(&loc).Error
	return &loc, err
}

func (r *FloorPlanRepo) UpdateLocation(loc *LocationModel) error {
	return r.db.Save(loc).Error
}

func (r *FloorPlanRepo) DeleteLocation(id string) error {
	return r.db.Where("id = ?", id).Delete(&LocationModel{}).Error
}

func (r *FloorPlanRepo) CountTasksByLocationID(locationID string) (int64, error) {
	var count int64
	err := r.db.Model(&TaskTemplateModel{}).Where("location_id = ?", locationID).Count(&count).Error
	return count, err
}
