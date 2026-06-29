package handler

import (
	"github.com/dezhishen/now-and-again/backend/pkg/contracts"
	"github.com/gin-gonic/gin"
)

type CalendarHandlers struct {
	Svc contracts.CalendarContract
}

func (h *CalendarHandlers) GetCalendar(c *gin.Context) {
	year := queryInt(c, "year", 0)
	month := queryInt(c, "month", 0)
	groupID := c.Query("group_id")

	if year <= 0 || month <= 0 || month > 12 {
		badRequest(c, "year and month are required (1-12)")
		return
	}

	days, err := h.Svc.GetCalendar(userCtx(c), year, month, groupID)
	if err != nil {
		serverError(c, err)
		return
	}
	ok(c, days)
}
