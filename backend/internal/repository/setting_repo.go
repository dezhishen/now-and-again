package repository

// ─── System Settings ────────────────────────────────────────────

func (r *SettingsRepo) GetAll() ([]SystemSettingModel, error) {
	var settings []SystemSettingModel
	err := r.db.Order("key ASC").Find(&settings).Error
	return settings, err
}

func (r *SettingsRepo) Get(key string) (*SystemSettingModel, error) {
	var s SystemSettingModel
	err := r.db.Where("key = ?", key).First(&s).Error
	return &s, err
}

func (r *SettingsRepo) Set(key, value string) error {
	return r.db.Save(&SystemSettingModel{Key: key, Value: value}).Error
}

func (r *SettingsRepo) SetDefaults(defaults map[string]string) error {
	for k, v := range defaults {
		var existing SystemSettingModel
		if err := r.db.Where("key = ?", k).First(&existing).Error; err != nil {
			r.db.Create(&SystemSettingModel{Key: k, Value: v})
		}
	}
	return nil
}
