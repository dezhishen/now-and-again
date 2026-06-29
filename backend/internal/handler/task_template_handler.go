package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/dezhishen/now-and-again/backend/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskTemplateHandlers holds the contract dependency.
type TaskTemplateHandlers struct {
	Svc contracts.TaskTemplateContract
}

// ─── Family-visible (all members) ─────────────────────────────────

// List returns templates visible to the current family (system + family-level).
func (h *TaskTemplateHandlers) List(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	kind := c.Query("kind")
	templates, err := h.Svc.List(userCtx(c), fid, kind)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, templates)
}

// Get returns a single template visible to the current family.
func (h *TaskTemplateHandlers) Get(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	code := c.Param("code")
	tmpl, err := h.Svc.GetByCode(userCtx(c), fid, code)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, tmpl)
}

// Render renders a template with the supplied parameter values.
func (h *TaskTemplateHandlers) Render(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	code := c.Param("code")
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		badRequest(c, "invalid request body")
		return
	}
	rendered, err := h.Svc.Render(userCtx(c), fid, code, params)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, rendered)
}

// ─── Family-level CRUD (owner only) ──────────────────────────────

// CreateFamily creates a new family-level template.
func (h *TaskTemplateHandlers) CreateFamily(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateTaskTemplateRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	tmpl, err := h.Svc.CreateFamily(userCtx(c), fid, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, tmpl)
}

// UpdateFamily updates a family-level template.
func (h *TaskTemplateHandlers) UpdateFamily(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	code := c.Param("code")
	req, err := bindJSON[types.UpdateTaskTemplateRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	tmpl, err := h.Svc.UpdateFamily(userCtx(c), fid, code, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, tmpl)
}

// DeleteFamily deletes a family-level template.
func (h *TaskTemplateHandlers) DeleteFamily(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	code := c.Param("code")
	if err := h.Svc.DeleteFamily(userCtx(c), fid, code); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"status": "deleted"})
}

// ─── Provider management ──────────────────────────────────────────

// ListProviders returns metadata for all registered providers.
func (h *TaskTemplateHandlers) ListProviders(c *gin.Context) {
	providers, err := h.Svc.ListProviders(userCtx(c))
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, providers)
}

// RefreshFamilyProvider triggers a sync for a family-level provider.
func (h *TaskTemplateHandlers) RefreshFamilyProvider(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	providerCode := c.Param("code")
	if err := h.Svc.RefreshFamilyProvider(userCtx(c), fid, providerCode); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"status": "ok"})
}

// ─── Admin operations ─────────────────────────────────────────────

// AdminRefreshProvider triggers a sync for a system-level provider.
func (h *TaskTemplateHandlers) AdminRefreshProvider(c *gin.Context) {
	providerCode := c.Param("code")
	if err := h.Svc.RefreshSystemProvider(userCtx(c), providerCode); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"status": "ok"})
}

// ─── Subscription management ──────────────────────────────────────

// AdminListSubscriptions returns system-level subscriptions.
func (h *TaskTemplateHandlers) AdminListSubscriptions(c *gin.Context) {
	subs, err := h.Svc.ListSubscriptions(userCtx(c), nil)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, subs)
}

// AdminCreateSubscription creates a system-level subscription.
func (h *TaskTemplateHandlers) AdminCreateSubscription(c *gin.Context) {
	req, err := bindJSON[types.CreateSubscriptionRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	sub, err := h.Svc.CreateSubscription(userCtx(c), nil, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, sub)
}

// AdminUpdateSubscription updates a system-level subscription.
func (h *TaskTemplateHandlers) AdminUpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	req, err := bindJSON[types.UpdateSubscriptionRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	sub, err := h.Svc.UpdateSubscription(userCtx(c), id, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, sub)
}

// AdminDeleteSubscription deletes a system-level subscription.
func (h *TaskTemplateHandlers) AdminDeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := h.Svc.DeleteSubscription(userCtx(c), id); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"status": "deleted"})
}

// FamilyListSubscriptions returns family-level subscriptions.
func (h *TaskTemplateHandlers) FamilyListSubscriptions(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	subs, err := h.Svc.ListSubscriptions(userCtx(c), &fid)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, subs)
}

// FamilyCreateSubscription creates a family-level subscription.
func (h *TaskTemplateHandlers) FamilyCreateSubscription(c *gin.Context) {
	fid, err := uuid.Parse(familyID(c))
	if err != nil {
		badRequest(c, "invalid family_id")
		return
	}
	req, err := bindJSON[types.CreateSubscriptionRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	sub, err := h.Svc.CreateSubscription(userCtx(c), &fid, req)
	if err != nil {
		serverError(c, err)
		return
	}
	created(c, sub)
}

// FamilyUpdateSubscription updates a family-level subscription.
func (h *TaskTemplateHandlers) FamilyUpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	req, err := bindJSON[types.UpdateSubscriptionRequest](c)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	sub, err := h.Svc.UpdateSubscription(userCtx(c), id, req)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, sub)
}

// FamilyDeleteSubscription deletes a family-level subscription.
func (h *TaskTemplateHandlers) FamilyDeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := h.Svc.DeleteSubscription(userCtx(c), id); err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"status": "deleted"})
}
