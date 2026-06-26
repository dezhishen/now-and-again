export type FamilyRole = 'owner' | 'admin' | 'member'
export type GroupRole = 'owner' | 'member'
export type MemberStatus = 'active' | 'pending' | 'rejected'

export interface User {
  id: string
  display_name: string
  email: string
  phone?: string
  avatar_url?: string
  roles: string[]
  created_at: string
  updated_at: string
}

export interface SystemStatus {
  initialized: boolean
}

export interface Family {
  id: string
  name: string
  invite_code: string
  created_by: string
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
  floor_plan_id: string
  name: string
  point: Point
  color: string
  created_at: string
  updated_at: string
}
