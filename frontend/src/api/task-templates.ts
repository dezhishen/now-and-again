import { api } from './client'
import type {
  TaskTemplate, TemplateProvider, RenderedTask,
  CreateTaskTemplateRequest, UpdateTaskTemplateRequest,
  TaskTemplateSubscription, CreateSubscriptionRequest, UpdateSubscriptionRequest,
} from '@/types'

// ─── Template queries (family-visible) ────────────────────────────

export async function listTemplates(kind?: string): Promise<TaskTemplate[]> {
  const params = kind ? `?kind=${encodeURIComponent(kind)}` : ''
  return api.getRaw<TaskTemplate[]>(`/task-templates${params}`)
}

export async function getTemplate(code: string): Promise<TaskTemplate> {
  return api.getRaw<TaskTemplate>(`/task-templates/${encodeURIComponent(code)}`)
}

export async function renderTemplate(code: string, params: Record<string, any>): Promise<RenderedTask> {
  return api.postRaw<RenderedTask>(`/task-templates/${encodeURIComponent(code)}/render`, params)
}

// ─── Family-level CRUD (owner only) ───────────────────────────────

export async function createFamilyTemplate(req: CreateTaskTemplateRequest): Promise<TaskTemplate> {
  return api.postRaw<TaskTemplate>('/task-templates', req)
}

export async function updateFamilyTemplate(code: string, req: UpdateTaskTemplateRequest): Promise<TaskTemplate> {
  return api.putRaw<TaskTemplate>(`/task-templates/${encodeURIComponent(code)}`, req)
}

export async function deleteFamilyTemplate(code: string): Promise<void> {
  return api.deleteRaw<void>(`/task-templates/${encodeURIComponent(code)}`)
}

// ─── Provider management ──────────────────────────────────────────

export async function listProviders(): Promise<TemplateProvider[]> {
  return api.getRaw<TemplateProvider[]>('/task-templates/providers')
}

export async function refreshFamilyProvider(code: string): Promise<void> {
  return api.postRaw<void>(`/task-templates/providers/${encodeURIComponent(code)}/refresh`)
}

export async function refreshSystemProvider(code: string): Promise<void> {
  return api.postRaw<void>(`/admin/task-templates/providers/${encodeURIComponent(code)}/refresh`)
}

// ─── Subscription management (family) ─────────────────────────────

export async function listFamilySubscriptions(): Promise<TaskTemplateSubscription[]> {
  return api.getRaw<TaskTemplateSubscription[]>('/task-template-subscriptions')
}

export async function createFamilySubscription(req: CreateSubscriptionRequest): Promise<TaskTemplateSubscription> {
  return api.postRaw<TaskTemplateSubscription>('/task-template-subscriptions', req)
}

export async function updateFamilySubscription(id: string, req: UpdateSubscriptionRequest): Promise<TaskTemplateSubscription> {
  return api.putRaw<TaskTemplateSubscription>(`/task-template-subscriptions/${encodeURIComponent(id)}`, req)
}

export async function deleteFamilySubscription(id: string): Promise<void> {
  return api.deleteRaw<void>(`/task-template-subscriptions/${encodeURIComponent(id)}`)
}

// ─── Subscription management (admin) ──────────────────────────────

export async function listAdminSubscriptions(): Promise<TaskTemplateSubscription[]> {
  return api.getRaw<TaskTemplateSubscription[]>('/admin/task-template-subscriptions')
}

export async function createAdminSubscription(req: CreateSubscriptionRequest): Promise<TaskTemplateSubscription> {
  return api.postRaw<TaskTemplateSubscription>('/admin/task-template-subscriptions', req)
}

export async function updateAdminSubscription(id: string, req: UpdateSubscriptionRequest): Promise<TaskTemplateSubscription> {
  return api.putRaw<TaskTemplateSubscription>(`/admin/task-template-subscriptions/${encodeURIComponent(id)}`, req)
}

export async function deleteAdminSubscription(id: string): Promise<void> {
  return api.deleteRaw<void>(`/admin/task-template-subscriptions/${encodeURIComponent(id)}`)
}
