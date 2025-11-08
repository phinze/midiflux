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

	// Get current time in user's timezone
	now := timezone.Now(user.Timezone)

	// Calculate date boundaries
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	// Week starts on Sunday (weekday 0)
	weekStart := todayStart.AddDate(0, 0, -int(todayStart.Weekday()))

	// Month starts on the 1st
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Query for "Today" entries
	todayBuilder := h.store.NewEntryQueryBuilder(user.ID)
	todayBuilder.WithStatus(model.EntryStatusUnread)
	todayBuilder.WithGloballyVisible()
	todayBuilder.WithSorting(user.EntryOrder, user.EntryDirection)
	todayBuilder.WithSorting("id", user.EntryDirection)
	todayBuilder.AfterPublishedDate(todayStart)
	todayEntries, err := todayBuilder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Query for "Yesterday" entries
	yesterdayBuilder := h.store.NewEntryQueryBuilder(user.ID)
	yesterdayBuilder.WithStatus(model.EntryStatusUnread)
	yesterdayBuilder.WithGloballyVisible()
	yesterdayBuilder.WithSorting(user.EntryOrder, user.EntryDirection)
	yesterdayBuilder.WithSorting("id", user.EntryDirection)
	yesterdayBuilder.AfterPublishedDate(yesterdayStart)
	yesterdayBuilder.BeforePublishedDate(todayStart)
	yesterdayEntries, err := yesterdayBuilder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Query for "This Week" entries (excluding today and yesterday)
	weekBuilder := h.store.NewEntryQueryBuilder(user.ID)
	weekBuilder.WithStatus(model.EntryStatusUnread)
	weekBuilder.WithGloballyVisible()
	weekBuilder.WithSorting(user.EntryOrder, user.EntryDirection)
	weekBuilder.WithSorting("id", user.EntryDirection)
	weekBuilder.AfterPublishedDate(weekStart)
	weekBuilder.BeforePublishedDate(yesterdayStart)
	weekEntries, err := weekBuilder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Query for "This Month" entries (excluding this week)
	monthBuilder := h.store.NewEntryQueryBuilder(user.ID)
	monthBuilder.WithStatus(model.EntryStatusUnread)
	monthBuilder.WithGloballyVisible()
	monthBuilder.WithSorting(user.EntryOrder, user.EntryDirection)
	monthBuilder.WithSorting("id", user.EntryDirection)
	monthBuilder.AfterPublishedDate(monthStart)
	monthBuilder.BeforePublishedDate(weekStart)
	monthEntries, err := monthBuilder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Query for "Earlier" entries (before this month)
	earlierBuilder := h.store.NewEntryQueryBuilder(user.ID)
	earlierBuilder.WithStatus(model.EntryStatusUnread)
	earlierBuilder.WithGloballyVisible()
	earlierBuilder.WithSorting(user.EntryOrder, user.EntryDirection)
	earlierBuilder.WithSorting("id", user.EntryDirection)
	earlierBuilder.BeforePublishedDate(monthStart)
	earlierEntries, err := earlierBuilder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	// Calculate total count
	countUnread, err := h.store.NewEntryQueryBuilder(user.ID).
		WithStatus(model.EntryStatusUnread).
		WithGloballyVisible().
		CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("todayEntries", todayEntries)
	view.Set("yesterdayEntries", yesterdayEntries)
	view.Set("weekEntries", weekEntries)
	view.Set("monthEntries", monthEntries)
	view.Set("earlierEntries", earlierEntries)
	view.Set("menu", "date_entries")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("date_entries"))
}
