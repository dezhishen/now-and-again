package model

// migrationModels collects GORM models from plugins via RegisterModel.
// AutoMigrate uses this list so that plugin-specific models don't need
// to be imported by internal/ code.
var migrationModels []any

// RegisterModel adds a GORM model to the auto-migration list.
// Plugin packages call this in init().
func RegisterModel(m any) {
	migrationModels = append(migrationModels, m)
}

// MigrationModels returns all registered GORM models for auto-migration.
func MigrationModels() []any {
	return migrationModels
}
