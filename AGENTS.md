# Miniflux Codebase Architecture Overview

## Project Summary

**Miniflux** is a minimalist, opinionated feed reader written in Go that emphasizes simplicity, performance, and privacy. The codebase is well-organized, follows clean architecture patterns, and is designed for maintainability.

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Gorilla Mux (routing)
- **Database**: PostgreSQL (no ORM, direct SQL)
- **Templating**: Go's html/template (with embedded templates)
- **Architecture**: Server-side rendered MVC-style application
- **Dependencies**: Minimal third-party dependencies

## Project Structure

### Top-Level Organization

```
/home/phinze/src/github.com/phinze/midiflux/
├── main.go                    # Entry point
├── internal/                  # Main application code
│   ├── api/                   # REST API handlers
│   ├── cli/                   # Command-line interface
│   ├── config/                # Configuration management
│   ├── database/              # Database migrations and setup
│   ├── http/                  # HTTP server, routing, middleware
│   ├── model/                 # Data structures
│   ├── storage/               # Database queries and data access
│   ├── template/              # Template engine and views
│   ├── ui/                    # Web UI handlers (101 files)
│   ├── locale/                # i18n translations (20 languages)
│   ├── timezone/              # Timezone handling
│   ├── reader/                # Feed parsing
│   ├── integration/           # Third-party integrations
│   └── worker/                # Background job processing
├── client/                    # Go API client library
└── packaging/                 # Deployment packages
```

## Core Architecture Components

### 1. Routing System (`/internal/http` & `/internal/ui`)

**Routes are defined in**: `/internal/ui/ui.go`

The routing follows a clear pattern using Gorilla Mux:
- **Route Registration**: `ui.Serve()` function registers all UI routes
- **Middleware Chain**: User session → App session → handlers
- **Named Routes**: All routes have names for reverse URL generation
- **Pattern**: `/path/{param}` → Handler function

**Key Route Examples**:
```go
// Entry browsing routes
"/unread" → showUnreadPage
"/history" → showHistoryPage
"/starred" → showStarredPage
"/search" → showSearchPage

// Feed/Category routes
"/feed/{feedID}/entries" → showFeedEntriesPage
"/category/{categoryID}/entries" → showCategoryEntriesPage
```

### 2. UI Handler Pattern (`/internal/ui/`)

**Each view has its own handler file** (101 total):
- `unread_entries.go` - Unread entries page
- `history_entries.go` - History/read entries page
- `starred_entries.go` - Starred entries page
- `category_entries.go` - Category-filtered entries
- `feed_entries.go` - Feed-specific entries
- `search.go` - Search results

**Standard Handler Pattern**:
```go
func (h *handler) showXXXPage(w http.ResponseWriter, r *http.Request) {
    // 1. Get user from request
    user, err := h.store.UserByID(request.UserID(r))

    // 2. Build query with EntryQueryBuilder
    builder := h.store.NewEntryQueryBuilder(user.ID)
    builder.WithStatus(model.EntryStatusUnread)
    builder.WithSorting(user.EntryOrder, user.EntryDirection)
    builder.WithOffset(offset)
    builder.WithLimit(user.EntriesPerPage)

    // 3. Fetch entries
    entries, err := builder.GetEntries()
    count, err := builder.CountEntries()

    // 4. Build view with data
    sess := session.New(h.store, request.SessionID(r))
    view := view.New(h.tpl, r, sess)
    view.Set("entries", entries)
    view.Set("menu", "unread")
    view.Set("user", user)

    // 5. Render template
    html.OK(w, r, view.Render("unread_entries"))
}
```

### 3. Data Layer (`/internal/model` & `/internal/storage`)

#### Models (`/internal/model/`)

**Entry Model** (critical for date grouping):
```go
type Entry struct {
    ID          int64
    UserID      int64
    FeedID      int64
    Status      string        // "unread", "read", "removed"
    Title       string
    URL         string
    Date        time.Time     // Published date (timezone-aware)
    CreatedAt   time.Time     // When added to DB
    ChangedAt   time.Time     // Last status change
    Content     string
    Author      string
    Starred     bool
    ReadingTime int
    Feed        *Feed
    Tags        []string
}
```

#### Storage/Query Builder (`/internal/storage/`)

**EntryQueryBuilder** (`entry_query_builder.go`) - Powerful, chainable query builder:

```go
// Available filters
builder.WithStatus(status)
builder.WithStarred(bool)
builder.WithFeedID(feedID)
builder.WithCategoryID(categoryID)
builder.WithSearchQuery(query)
builder.WithTags(tags)

// Date filters (IMPORTANT for date-based view!)
builder.AfterPublishedDate(date)
builder.BeforePublishedDate(date)
builder.AfterChangedDate(date)
builder.BeforeChangedDate(date)

// Sorting and pagination
builder.WithSorting(column, direction)
builder.WithLimit(limit)
builder.WithOffset(offset)

// Execution
entries, err := builder.GetEntries()
count, err := builder.CountEntries()
```

### 4. Template System (`/internal/template/`)

**Template Engine** (`engine.go`):
- Templates embedded in binary using `//go:embed`
- Two template directories:
  - `templates/common/` - Shared components (layout, pagination, etc.)
  - `templates/views/` - Page-specific templates
- Template composition via Go's template inheritance
- Runtime function injection for i18n (t, plural, elapsed)

**Template Structure**:
```
templates/
├── common/
│   ├── layout.html         # Base layout with header/nav
│   ├── item_meta.html      # Entry metadata/actions
│   ├── pagination.html     # Pagination controls
│   ├── feed_list.html      # Feed listing component
│   └── feed_menu.html      # Feed menu component
└── views/
    ├── unread_entries.html # Unread page
    ├── history_entries.html # History page
    ├── starred_entries.html # Starred page
    └── ... (30 total views)
```

**Template Functions** (`functions.go`):
- `elapsed` - Human-readable time (e.g., "2 hours ago", "yesterday")
- `isodate` - ISO format for datetime attributes
- `t` - Translation function
- `plural` - Pluralization
- `route` - Generate URLs from route names
- `icon` - SVG icon insertion

### 5. Date/Time Handling (`/internal/timezone/`)

**Timezone Conversion** (`timezone.go`):
- All dates stored in PostgreSQL with timezone
- Converted to user's timezone in queries: `e.published_at at time zone u.timezone`
- `timezone.Convert()` ensures proper timezone handling
- User timezone preference stored in User model

**Time Display** (`functions.go` - `elapsedTime()`):
```go
// Current implementation groups by:
< 60s: "just now"
< 1hr: "X minutes ago"
< 24hr: "X hours ago"
1 day: "yesterday"
< 21 days: "X days ago"
< 31 days: "X weeks ago"
< 365 days: "X months ago"
>= 365 days: "X years ago"
```

### 6. Internationalization (`/internal/locale/`)

**Translation files**: 20 languages in JSON format
- Key-value structure: `"page.unread.title": "Unread"`
- Pluralization support
- Runtime translation via template functions

**Adding new strings**:
1. Add to all 20 translation files
2. Use in templates: `{{ t "key" }}` or `{{ plural "key" count }}`

## Existing Entry Views

### 1. Unread Entries (`/unread`)
- **Handler**: `unread_entries.go`
- **Template**: `unread_entries.html`
- **Query**: `WithStatus(EntryStatusUnread)`
- **Sorting**: User preference

### 2. History Entries (`/history`)
- **Handler**: `history_entries.go`
- **Template**: `history_entries.html`
- **Query**: `WithStatus(EntryStatusRead)`
- **Sorting**: `changed_at DESC, published_at DESC`

### 3. Starred Entries (`/starred`)
- **Handler**: `starred_entries.go`
- **Template**: `starred_entries.html`
- **Query**: `WithStarred(true)`

### 4. Search (`/search`)
- **Handler**: `search.go`
- **Template**: `search.html`
- **Query**: `WithSearchQuery(query)` - full-text search

## Database Schema (Relevant Parts)

**Entries table**:
```sql
CREATE TABLE entries (
    id BIGSERIAL,
    user_id int not null,
    feed_id bigint not null,
    hash text not null,
    published_at timestamp with time zone not null,  -- KEY for date grouping!
    title text not null,
    url text not null,
    author text,
    content text,
    status entry_status default 'unread',
    starred boolean default 'f',
    created_at timestamp with time zone not null,
    changed_at timestamp with time zone not null,
);
```

## Pattern: Adding a New Entry List View

### Steps to add a date-grouped view:

1. **Create handler** (`/internal/ui/date_entries.go`):
   ```go
   func (h *handler) showDateEntriesPage(w http.ResponseWriter, r *http.Request) {
       // Get user, build queries with date filters
       // Use builder.AfterPublishedDate() / BeforePublishedDate()
   }
   ```

2. **Register route** (`/internal/ui/ui.go`):
   ```go
   uiRouter.HandleFunc("/entries/by-date", handler.showDateEntriesPage).Name("dateEntries").Methods(http.MethodGet)
   ```

3. **Create template** (`/internal/template/templates/views/date_entries.html`):
   ```html
   {{ define "title"}}{{ t "page.date_entries.title" }}{{ end }}
   {{ define "content"}}
       <!-- Group entries by date sections -->
   {{ end }}
   ```

4. **Register template** (`/internal/template/engine.go`):
   ```go
   "date_entries.html": {"item_meta.html", "layout.html"},
   ```

5. **Add translations** (all 20 files in `/internal/locale/translations/`):
   ```json
   "page.date_entries.title": "By Date"
   ```

6. **Add navigation** (`/internal/template/templates/common/layout.html`):
   ```html
   <li {{ if eq .menu "date_entries" }}class="active"{{ end }}>
       <a href="{{ route "dateEntries" }}">{{ t "menu.date_entries" }}</a>
   </li>
   ```

### Date-Based Querying Example

For grouping entries by publish date:

```go
// Get current time in user's timezone
now := timezone.Now(user.Timezone)

// Calculate date boundaries
todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
yesterdayStart := todayStart.AddDate(0, 0, -1)
weekStart := todayStart.AddDate(0, 0, -int(todayStart.Weekday()))
monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

// Query for "Today" section
todayBuilder := h.store.NewEntryQueryBuilder(user.ID)
todayBuilder.WithStatus(model.EntryStatusUnread)
todayBuilder.AfterPublishedDate(todayStart)
todayEntries, _ := todayBuilder.GetEntries()

// Query for "Yesterday" section
yesterdayBuilder := h.store.NewEntryQueryBuilder(user.ID)
yesterdayBuilder.WithStatus(model.EntryStatusUnread)
yesterdayBuilder.AfterPublishedDate(yesterdayStart)
yesterdayBuilder.BeforePublishedDate(todayStart)
yesterdayEntries, _ := yesterdayBuilder.GetEntries()
```

### Template Grouping Example

```html
{{ define "content"}}
{{ if gt (len .todayEntries) 0 }}
    <section class="date-group">
        <h2>{{ t "date_group.today" }}</h2>
        <div class="items">
            {{ range .todayEntries }}
                {{ template "item_meta" dict "user" $.user "entry" . }}
            {{ end }}
        </div>
    </section>
{{ end }}

{{ if gt (len .yesterdayEntries) 0 }}
    <section class="date-group">
        <h2>{{ t "date_group.yesterday" }}</h2>
        ...
    </section>
{{ end }}
{{ end }}
```

## Code Conventions to Follow

1. **File Organization**: One handler per file, named after the view
2. **Error Handling**: Always check errors, use `html.ServerError()` for failures
3. **Licensing**: SPDX headers on all files: `// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.`
4. **Import Organization**: Standard library → Third party → Internal packages
5. **Query Building**: Always use EntryQueryBuilder, never raw SQL in handlers
6. **Timezone Aware**: Use `timezone.Convert()` and `timezone.Now()`
7. **i18n**: All user-facing strings must be translatable
8. **Accessibility**: Use semantic HTML, ARIA labels, proper heading hierarchy

## Maintainability Strategy for Custom Features

To maintain clean commits that can track over upstream changes:

1. **Follow Miniflux patterns exactly** - Use the same conventions, naming, and structure
2. **Minimize file modifications** - Create new files when possible
3. **Keep changes isolated** - Each feature in its own commit
4. **Document custom additions** - Mark custom code with comments
5. **Test across updates** - Verify features work after upstream merges
6. **Consider configuration** - Make features toggleable via config if possible

## Files to Modify for Date-Based View

**New files to create**:
- `/internal/ui/date_entries.go` - Handler
- `/internal/template/templates/views/date_entries.html` - Template

**Files to modify** (minimal changes):
- `/internal/ui/ui.go` - Add route (1 line)
- `/internal/template/engine.go` - Register template (1 line)
- `/internal/locale/translations/*.json` - Add translations (20 files, few lines each)
- `/internal/template/templates/common/layout.html` - Add navigation link (optional)

This approach minimizes merge conflicts while following Miniflux's architecture patterns.
