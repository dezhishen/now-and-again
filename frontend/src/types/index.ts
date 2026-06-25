// Shared TypeScript type definitions mirroring shared/types in Go.
// In production, consider generating these from an OpenAPI spec.

export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'blocked' | 'archived'
export type TaskCategory = 'now' | 'again'
export type Priority = 'low' | 'medium' | 'high' | 'urgent'
export type FamilyRole = 'owner' | 'admin' | 'member'
export type DependencyType = 'blocks' | 'relates_to'

export interface User {
  id: string
  username: string
  email: string
  phone?: string
  display_name: string
  avatar_url?: string
  is_admin: boolean
  created_at: string
  updated_at: string
}

export interface Family {
  id: string
  name: string
  invite_code: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface FamilyMember {
  id: string
  family_id: string
  user_id: string
  role: FamilyRole
  joined_at: string
  user?: User
}

export interface SubGroup {
  id: string
  family_id: string
  name: string
  description?: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface ScheduleType {
  id: string
  code: string
  name: string
  category: TaskCategory
  default_priority: Priority
  icon?: string
  is_active: boolean
}

export interface Task {
  id: string
  family_id: string
  sub_group_id?: string
  task_code: string
  chain_id?: string
  title: string
  description?: string
  status: TaskStatus
  priority: Priority
  due_date?: string
  created_by: string
  completed_at?: string
  created_at: string
  updated_at: string
  task_type?: ScheduleType
  assignees?: TaskAssignee[]
  blocked_by?: TaskDependency[]
}

export interface TaskAssignee {
  id: string
  task_id: string
  user_id: string
  assigned_at: string
  user?: User
}

export interface TaskDependency {
  id: string
  blocked_task_id: string
  blocker_task_id: string
  dependency_type: DependencyType
}

export interface TaskChain {
  id: string
  family_id: string
  name: string
  description?: string
  icon?: string
  is_active: boolean
  steps?: TaskChainStep[]
}

export interface TaskChainStep {
  id: string
  chain_id: string
  sort_order: number
  title: string
  description?: string
  task_code: string
  assigned_role: string
  delay_after_previous: string
  is_optional: boolean
  priority: Priority
}

export interface Inspection {
  id: string
  family_id: string
  title: string
  description?: string
  status: string
  created_by: string
  completed_at?: string
  items?: InspectionItem[]
}

export interface InspectionItem {
  id: string
  inspection_id: string
  check_point: string
  result: 'ok' | 'issue_found'
  note?: string
  generated_task_id?: string
}

export interface APIResponse<T> {
  success: boolean
  data: T
  error?: string
}

export interface PagedResponse<T> {
  success: boolean
  data: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

// ─── Auth / Setup ────────────────────────────────────────────────

export interface SetupRequest {
  username: string
  email: string
  password: string
  display_name: string
}

export interface SystemStatus {
  initialized: boolean
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: number
  user: User
}
