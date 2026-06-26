package repository

// ─── ICS Feed ────────────────────────────────────────────────────

func (r *IcsRepo) CreateFeed(f *IcsFeedModel) error {
	return r.db.Create(f).Error
}

func (r *IcsRepo) FindFeedByID(id string) (*IcsFeedModel, error) {
	var f IcsFeedModel
	err := r.db.Preload("ApiKey").Where("id = ?", id).First(&f).Error
	return &f, err
}

func (r *IcsRepo) FindFeedByToken(token string) (*IcsFeedModel, error) {
	var f IcsFeedModel
	err := r.db.Preload("ApiKey").Where("access_token = ?", token).First(&f).Error
	return &f, err
}

func (r *IcsRepo) ListFeedsByFamily(familyID string) ([]IcsFeedModel, error) {
	var feeds []IcsFeedModel
	err := r.db.Preload("ApiKey").Where("family_id = ?", familyID).Order("created_at ASC").Find(&feeds).Error
	return feeds, err
}

func (r *IcsRepo) UpdateFeed(f *IcsFeedModel) error {
	return r.db.Save(f).Error
}

func (r *IcsRepo) DeleteFeed(id string) error {
	return r.db.Where("id = ?", id).Delete(&IcsFeedModel{}).Error
}

func (r *IcsRepo) IsUsernameTaken(username string) (bool, error) {
	var count int64
	err := r.db.Model(&IcsFeedModel{}).Where("app_username = ?", username).Count(&count).Error
	return count > 0, err
}
