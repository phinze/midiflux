# Midiflux - Personal Miniflux Fork

This is a personal fork of [Miniflux](https://github.com/miniflux/v2) maintained by [@phinze](https://github.com/phinze) with custom features for personal use.

## Why Fork?

This fork adds features that are specific to my personal workflow and may not align with Miniflux's minimalist philosophy. Rather than burden the upstream maintainers with niche use cases, I maintain these features separately while staying synchronized with upstream improvements.

## Custom Features

### 1. Date-Based Entry View (`feature/date-based-view`)

A chronological view that groups unread entries into rolling date sections:
- **Today** - entries published today
- **Yesterday** - entries from yesterday
- **This Week** - entries from the current week
- **Last Week** - entries from the previous week
- **This Month** - older entries from the current month
- **Older** - everything else

**Implementation:**
- Handler: `internal/ui/date_entries.go`
- Template: `internal/template/templates/views/date_entries.html`
- Route: `/entries/by-date`
- Minimal modifications to existing files (routing, templates, translations)

## Upstream Synchronization Strategy

### Current Status

- **Base Version:** Miniflux v2.2.14 (+ 13 commits)
- **Upstream Remote:** `https://github.com/miniflux/v2.git`
- **Sync Strategy:** Rebase workflow for clean history

### Maintenance Workflow

#### Regular Sync (every few weeks/months)

```bash
# 1. Check what's new upstream
git fetch upstream
git log HEAD..upstream/main --oneline --graph

# 2. Review changes for potential conflicts
git diff HEAD..upstream/main internal/ui/ui.go
git diff HEAD..upstream/main internal/template/engine.go

# 3. Sync main branch
git checkout main
git rebase upstream/main
# Fix any conflicts if needed
git push origin main --force-with-lease

# 4. Rebase feature branches
git checkout feature/date-based-view
git rebase main
# Fix any conflicts if needed
git push origin feature/date-based-view --force-with-lease

# 5. Test and redeploy
# Run tests, verify features, then deploy via Nomad
```

#### Automated Monitoring

A GitHub Action (`.github/workflows/sync-upstream.yml`) runs daily to:
- Check for upstream changes
- Create an issue when updates are available
- Provide a checklist for syncing

### Conflict Prevention

Custom changes are isolated to minimize conflicts:

**New files (no conflicts):**
- `internal/ui/date_entries.go`
- `internal/template/templates/views/date_entries.html`
- Custom translations

**Modified files (minimal, marked with `// CUSTOM:`):**
- `internal/ui/ui.go` - route registration (~1 line)
- `internal/template/engine.go` - template registration (~1 line)
- `internal/locale/translations/*.json` - translation strings
- `internal/template/templates/common/layout.html` - navigation links (if added)

All custom changes are marked with `// CUSTOM:` comments for easy identification during merges.

## Development Environment

### Quick Start with Nix

```bash
# Enter development environment
nix develop

# Set up development database
./dev-db-setup.sh

# Build and run
make run
```

### Manual Setup

See `DEV.md` for detailed development setup instructions.

## Deployment

### Docker Image

Custom builds are automatically published to GitHub Container Registry:

- **Image:** `ghcr.io/phinze/midiflux:latest`
- **Build Trigger:** Pushes to `main` and `feature/*` branches
- **Workflow:** `.github/workflows/docker.yml`

Branch-specific images are also available:
- `ghcr.io/phinze/midiflux:feature-date-based-view`
- `ghcr.io/phinze/midiflux:main-<sha>`

### Personal Infrastructure

Deployed via Nomad to personal infrastructure:
- **Job:** `../infra/nomad-jobs/miniflux.nomad`
- **URL:** https://mf.inze.ph/

## Architecture Documentation

See `CLAUDE.md` for comprehensive architecture documentation tailored for AI-assisted development.

## Contributing

This is a personal fork for private use. If you're interested in these features:

1. **For general use:** Consider requesting the feature in the [upstream Miniflux project](https://github.com/miniflux/v2/discussions)
2. **For your own fork:** Feel free to cherry-pick commits or use this as a reference
3. **For improvements to fork strategy:** Open an issue or PR!

## License

Same as upstream Miniflux: [Apache License 2.0](LICENSE)

Original work Copyright © Frédéric Guillot
Fork modifications Copyright © 2025 Phil Hintze

## Acknowledgments

Huge thanks to [Frédéric Guillot](https://github.com/fguillot) and the Miniflux community for creating and maintaining such a fantastic feed reader. This fork exists because the upstream project is so well-architected that it's easy to extend.
