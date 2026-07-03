package repository

// ─── System Settings ────────────────────────────────────────────

func (r *SettingsRepo) GetAll() ([]SystemSettingModel, error) {
	var settings []SystemSettingModel
	err := r.db.Order("key ASC").Find(&settings).Error
	return settings, err
}

func (r *SettingsRepo) GetByScope(scope string) ([]SystemSettingModel, error) {
	var settings []SystemSettingModel
	err := r.db.Where("scope = ?", scope).Order("key ASC").Find(&settings).Error
	return settings, err
}

func (r *SettingsRepo) Get(key string) (*SystemSettingModel, error) {
	var s SystemSettingModel
	err := r.db.Where("key = ?", key).First(&s).Error
	if err != nil {
		return nil, ignoreNotFound(err)
	}
	return &s, nil
}

// Set upserts a setting. For existing keys only the value is updated (scope preserved).
// New keys get scope "admin" by default.
func (r *SettingsRepo) Set(key, value string) error {
	var existing SystemSettingModel
	if err := r.db.Where("key = ?", key).First(&existing).Error; err != nil {
		return r.db.Create(&SystemSettingModel{Key: key, Value: value, Scope: "admin"}).Error
	}
	return r.db.Model(&existing).Update("value", value).Error
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
