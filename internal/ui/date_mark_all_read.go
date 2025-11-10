// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/timezone"
)

// CUSTOM: markDateEntriesAsRead marks entries as read within the selected date section
func (h *handler) markDateEntriesAsRead(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)

	user, err := h.store.UserByID(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	// Get section filter from query parameter
	section := request.QueryStringParam(r, "section", "all")

	// Get current time in user's timezone
	now := timezone.Now(user.Timezone)

	// Calculate date boundaries matching the showDateEntriesPage logic
	todayStart := now.Add(-24 * time.Hour)
	last2dStart := now.Add(-48 * time.Hour)
	last7dStart := now.Add(-7 * 24 * time.Hour)
	last30dStart := now.Add(-30 * 24 * time.Hour)

	var afterDate, beforeDate *time.Time

	// Determine date range based on section
	switch section {
	case "today":
		afterDate = &todayStart
		beforeDate = nil // Up to now
	case "last2d":
		afterDate = &last2dStart
		beforeDate = &todayStart
	case "last7d":
		afterDate = &last7dStart
		beforeDate = &last2dStart
	case "last30d":
		afterDate = &last30dStart
		beforeDate = &last7dStart
	case "earlier":
		afterDate = nil // Beginning of time
		beforeDate = &last30dStart
	case "all":
		// Mark all globally visible entries as read
		if err := h.store.MarkGloballyVisibleFeedsAsRead(userID); err != nil {
			json.ServerError(w, r, err)
			return
		}
		json.OK(w, r, "OK")
		return
	}

	// Mark entries in the specified date range
	if err := h.store.MarkEntriesAsReadInDateRange(userID, afterDate, beforeDate); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}
