// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/timezone"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showDateEntriesPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Get section filter from query parameter (default: "today")
	section := request.QueryStringParam(r, "section", "today")

	// Get current time in user's timezone
	now := timezone.Now(user.Timezone)

	// Calculate date boundaries (rolling windows)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	last2dStart := todayStart.AddDate(0, 0, -3)   // Last 2 days: days -1 and -2
	last7dStart := todayStart.AddDate(0, 0, -10)  // Last 7 days: days -3 to -9
	last30dStart := todayStart.AddDate(0, 0, -40) // Last 30 days: days -10 to -39
	// Earlier: anything before last30dStart (>40 days ago)

	// Helper function to count entries for a date range
	countForDateRange := func(afterDate, beforeDate *time.Time) (int, error) {
		builder := h.store.NewEntryQueryBuilder(user.ID)
		builder.WithStatus(model.EntryStatusUnread)
		builder.WithGloballyVisible()
		if afterDate != nil {
			builder.AfterPublishedDate(*afterDate)
		}
		if beforeDate != nil {
			builder.BeforePublishedDate(*beforeDate)
		}
		return builder.CountEntries()
	}

	// Helper function to fetch entries for a date range
	fetchForDateRange := func(afterDate, beforeDate *time.Time) ([]*model.Entry, error) {
		builder := h.store.NewEntryQueryBuilder(user.ID)
		builder.WithStatus(model.EntryStatusUnread)
		builder.WithGloballyVisible()
		builder.WithSorting(user.EntryOrder, user.EntryDirection)
		builder.WithSorting("id", user.EntryDirection)
		if afterDate != nil {
			builder.AfterPublishedDate(*afterDate)
		}
		if beforeDate != nil {
			builder.BeforePublishedDate(*beforeDate)
		}
		return builder.GetEntries()
	}

	// Get counts for all sections (for navigation)
	countToday, err := countForDateRange(&todayStart, nil)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	countLast2d, err := countForDateRange(&last2dStart, &todayStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	countLast7d, err := countForDateRange(&last7dStart, &last2dStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	countLast30d, err := countForDateRange(&last30dStart, &last7dStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	countEarlier, err := countForDateRange(nil, &last30dStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Initialize empty entry slices
	var todayEntries, last2dEntries, last7dEntries, last30dEntries, earlierEntries []*model.Entry

	// Fetch entries only for the selected section
	switch section {
	case "today":
		todayEntries, err = fetchForDateRange(&todayStart, nil)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "last2d":
		last2dEntries, err = fetchForDateRange(&last2dStart, &todayStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "last7d":
		last7dEntries, err = fetchForDateRange(&last7dStart, &last2dStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "last30d":
		last30dEntries, err = fetchForDateRange(&last30dStart, &last7dStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "earlier":
		earlierEntries, err = fetchForDateRange(nil, &last30dStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	default: // "all" or any other value
		// Fetch all sections
		todayEntries, err = fetchForDateRange(&todayStart, nil)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		last2dEntries, err = fetchForDateRange(&last2dStart, &todayStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		last7dEntries, err = fetchForDateRange(&last7dStart, &last2dStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		last30dEntries, err = fetchForDateRange(&last30dStart, &last7dStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		earlierEntries, err = fetchForDateRange(nil, &last30dStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	}

	// Calculate total count
	countUnread := countToday + countLast2d + countLast7d + countLast30d + countEarlier

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("todayEntries", todayEntries)
	view.Set("last2dEntries", last2dEntries)
	view.Set("last7dEntries", last7dEntries)
	view.Set("last30dEntries", last30dEntries)
	view.Set("earlierEntries", earlierEntries)
	view.Set("countToday", countToday)
	view.Set("countLast2d", countLast2d)
	view.Set("countLast7d", countLast7d)
	view.Set("countLast30d", countLast30d)
	view.Set("countEarlier", countEarlier)
	view.Set("section", section)
	view.Set("menu", "date_entries")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("date_entries"))
}
