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

	// Get section filter from query parameter (default: "all")
	section := request.QueryStringParam(r, "section", "all")

	// Get current time in user's timezone
	now := timezone.Now(user.Timezone)

	// Calculate date boundaries
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	// Week starts on Sunday (weekday 0)
	weekStart := todayStart.AddDate(0, 0, -int(todayStart.Weekday()))

	// Month starts on the 1st
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

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

	countYesterday, err := countForDateRange(&yesterdayStart, &todayStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	countWeek, err := countForDateRange(&weekStart, &yesterdayStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	countMonth, err := countForDateRange(&monthStart, &weekStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	countEarlier, err := countForDateRange(nil, &monthStart)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Initialize empty entry slices
	var todayEntries, yesterdayEntries, weekEntries, monthEntries, earlierEntries []*model.Entry

	// Fetch entries only for the selected section
	switch section {
	case "today":
		todayEntries, err = fetchForDateRange(&todayStart, nil)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "yesterday":
		yesterdayEntries, err = fetchForDateRange(&yesterdayStart, &todayStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "week":
		weekEntries, err = fetchForDateRange(&weekStart, &yesterdayStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "month":
		monthEntries, err = fetchForDateRange(&monthStart, &weekStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	case "earlier":
		earlierEntries, err = fetchForDateRange(nil, &monthStart)
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
		yesterdayEntries, err = fetchForDateRange(&yesterdayStart, &todayStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		weekEntries, err = fetchForDateRange(&weekStart, &yesterdayStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		monthEntries, err = fetchForDateRange(&monthStart, &weekStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		earlierEntries, err = fetchForDateRange(nil, &monthStart)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	}

	// Calculate total count
	countUnread := countToday + countYesterday + countWeek + countMonth + countEarlier

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("todayEntries", todayEntries)
	view.Set("yesterdayEntries", yesterdayEntries)
	view.Set("weekEntries", weekEntries)
	view.Set("monthEntries", monthEntries)
	view.Set("earlierEntries", earlierEntries)
	view.Set("countToday", countToday)
	view.Set("countYesterday", countYesterday)
	view.Set("countWeek", countWeek)
	view.Set("countMonth", countMonth)
	view.Set("countEarlier", countEarlier)
	view.Set("section", section)
	view.Set("menu", "date_entries")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("date_entries"))
}
