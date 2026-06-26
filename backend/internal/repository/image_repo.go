package repository

// ─── Image ───────────────────────────────────────────────────────

func (r *ImageRepo) CreateImage(img *ImageModel) error {
	return r.db.Create(img).Error
}

func (r *ImageRepo) FindImageByID(id string) (*ImageModel, error) {
	var img ImageModel
	err := r.db.Where("id = ?", id).First(&img).Error
	return &img, err
}

func (r *ImageRepo) DeleteImage(id string) error {
	return r.db.Where("id = ?", id).Delete(&ImageModel{}).Error
}
