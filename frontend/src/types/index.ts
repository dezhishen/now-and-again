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

export interface APIResponse<T> {
  success: boolean
  data: T
  error?: string
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
