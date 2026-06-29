// Package repository provides database access layer.
// Model types are defined in pkg/model and re-exported here via type aliases
// for backward compatibility. New code should import pkg/model directly.
package repository

import "github.com/dezhishen/now-and-again/backend/pkg/model"

// ─── Type aliases (re-exported from pkg/model) ────────────────────

type BaseModel = model.BaseModel

type AccountModel = model.AccountModel
type UserModel = model.UserModel
type RoleModel = model.RoleModel
type UserRoleModel = model.UserRoleModel

type FamilyModel = model.FamilyModel
type FamilyMemberModel = model.FamilyMemberModel
type FamilyGroupModel = model.FamilyGroupModel
type FamilyGroupMemberModel = model.FamilyGroupMemberModel

type RefreshTokenModel = model.RefreshTokenModel
type ApiKeyModel = model.ApiKeyModel

type ImageModel = model.ImageModel
type SystemSettingModel = model.SystemSettingModel

type FloorPlanModel = model.FloorPlanModel
type LocationModel = model.LocationModel

type TaskModel = model.TaskModel
type TodoModel = model.TodoModel
type TaskLogModel = model.TaskLogModel

type IcsFeedModel = model.IcsFeedModel

type TaskTemplateModel = model.TaskTemplateModel

type TaskTemplateSubscriptionModel = model.TaskTemplateSubscriptionModel
