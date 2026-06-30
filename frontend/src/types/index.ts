export type FamilyRole = 'owner' | 'admin' | 'member'
export type GroupRole = 'owner' | 'member'
export type MemberStatus = 'active' | 'pending' | 'rejected'

export interface User {
  id: string
  display_name: string
  email: string
  phone?: string
  avatar_url?: string
  default_family_id?: string
  roles: string[]
  created_at: string
  updated_at: string
}

export interface Family {
  id: string
  name: string
  invite_code: string
  created_by: string
  archived: boolean
  thumbnail_url?: string
  created_at: string
  updated_at: string
}

export interface FamilyMember {
  id: string
  family_id: string
  user_id: string
  role: FamilyRole
  status: MemberStatus
  joined_at: string
  user?: User
}

export interface FamilyGroup {
  id: string
  family_id: string
  name: string
  description?: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface FamilyGroupMember {
  id: string
  group_id: string
  user_id: string
  role: GroupRole
  status: MemberStatus
  joined_at: string
  user?: User
}

export interface ApiKey {
  id: string
  name: string
  key_prefix: string
  raw_key?: string
  scopes?: string[]
  last_used_at?: string
  expires_at?: string
  created_at: string
}

// ─── Unified error types ─────────────────────────────────────────

export type ErrorCode = 'BAD_REQUEST' | 'VALIDATION_ERROR' | 'UNAUTHORIZED' | 'FORBIDDEN' | 'NOT_FOUND' | 'CONFLICT' | 'INTERNAL_ERROR'

export interface FieldError {
  field: string
  message: string
}

export interface ApiError {
  code: ErrorCode
  summary: string
  details?: FieldError[]
}

export class ApiRequestError extends Error {
  code: ErrorCode
  details: FieldError[]
  summary: string
  constructor(err: ApiError) {
    super(err.summary)
    this.name = 'ApiRequestError'
    this.code = err.code
    this.summary = err.summary
    this.details = err.details || []
  }
}

// ── ErrorCode → plain-text message (no i18n, for toast/log fallback) ──

type MessageBuilder = (err: ApiError) => string

export const ERROR_MESSAGES: Record<ErrorCode, MessageBuilder> = {
  BAD_REQUEST:       (e) => e.summary,
  VALIDATION_ERROR:  (e) => e.details?.map(d => d.field + ': ' + d.message).join('; ') || e.summary,
  UNAUTHORIZED:      () => '请先登录',
  FORBIDDEN:         () => '没有权限执行此操作',
  NOT_FOUND:         () => '请求的资源不存在',
  CONFLICT:          () => '数据冲突',
  INTERNAL_ERROR:    () => '服务器内部错误，请稍后重试',
}

/** Build a plain-text message from an ApiError, using ERROR_MESSAGES registry. */
export function formatError(err: ApiError): string {
  const h = ERROR_MESSAGES[err.code]
  return h ? h(err) : err.summary
}

export interface APIResponse<T> {
  success: boolean
  data: T
  error?: ApiError
}

// ─── Floor Plan ──────────────────────────────────────────────────

export interface Point {
  x: number
  y: number
}

export interface FloorPlan {
  id: string
  family_id: string
  label: string
  image_id?: string
  image_url: string
  is_cover: boolean
  width: number
  height: number
  locations?: Location[]
  created_at: string
  updated_at: string
}

export interface Location {
  id: string
  family_id: string
  floor_plan_id?: string
  kind: string
  name: string
  color: string
  created_at: string
  updated_at: string
}

export interface Task {
  id: string
  family_id: string
  group_id?: string
  location_id?: string
  parent_task_id?: string
  is_root: boolean
  name: string
  schedule_type: string
  schedule_data: any
  enabled: boolean
  kind: string
  display_summary?: string
  archived: boolean
  last_todo_at?: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface CheckItem {
  id?: string
  name: string
  sort_order?: number
  branches: BranchItem[]
}

export interface BranchItem {
  id?: string
  name: string
  create_todo: boolean
  branch_task_id?: string
  sort_order?: number
  /** 创建参数（输入）/ 已创建的子任务（输出）。复用 TaskWithExtra。 */
  branch_task?: TaskWithExtra
}

export interface TaskWithExtra {
  task: Task
  extra?: any
}

export interface CompleteTodoRequest {
  todo?: {
    status: 'done' | 'skipped'
    remark?: string
  }
  extra?: any
}

export interface Todo {
  id: string
  task_id: string
  family_id: string
  location_id?: string
  assigned_to?: string
  status: 'pending' | 'done' | 'skipped'
  remark?: string
  due_start: string
  due_date: string
  completed_at?: string
  completed_by?: string
  task?: Task
  user?: User
  created_at: string
  updated_at: string
}

// ─── Task Template ────────────────────────────────────────────────

export interface TemplateParameter {
  key: string
  label: string
  type: 'string' | 'int' | 'float' | 'bool' | 'select' | 'time'
  description?: string
  required?: boolean
  default?: any
  options?: { label: string; value: string }[]
  placeholder?: string
}

export interface TaskTemplate {
  id: string
  family_id?: string | null
  provider_code: string
  template_code: string
  name: string
  description?: string
  kind: string
  icon?: string
  sort_order: number
  enabled: boolean
  parameters?: TemplateParameter[]
  task_defaults?: any
  extra_schema?: any
  version?: string
  created_at: string
  updated_at: string
}

export interface TemplateProvider {
  code: string
  name: string
  description?: string
  last_sync_at?: string
  sync_status: string
}

export interface RenderedTask {
  task_defaults: any
  extra_schema?: any
}

export interface CreateTaskTemplateRequest {
  template_code: string
  name: string
  description?: string
  kind: string
  icon?: string
  sort_order?: number
  enabled?: boolean
  parameters?: TemplateParameter[]
  task_defaults?: any
  extra_schema?: any
}

export interface UpdateTaskTemplateRequest {
  name?: string
  description?: string
  kind?: string
  icon?: string
  sort_order?: number
  enabled?: boolean
  parameters?: TemplateParameter[]
  task_defaults?: any
  extra_schema?: any
}

export interface TaskTemplateSubscription {
  id: string
  family_id?: string | null
  provider_code: string
  url: string
  name: string
  auto_refresh: boolean
  refresh_interval_hours: number
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CreateSubscriptionRequest {
  provider_code: string
  url: string
  name: string
  auto_refresh?: boolean
  refresh_interval_hours?: number
}

export interface UpdateSubscriptionRequest {
  url?: string
  name?: string
  auto_refresh?: boolean
  refresh_interval_hours?: number
  enabled?: boolean
}
