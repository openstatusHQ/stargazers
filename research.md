# Stargazers Insights - Project Research

## Overview

A Go CLI tool that fetches and stores GitHub stargazer data for specified repositories. It uses the GitHub GraphQL API (v4) to pull user and company information about people who starred your repos, storing everything in a local SQLite database for analysis.

## Tech Stack

- **Language**: Go 1.25
- **CLI Framework**: `urfave/cli/v3`
- **GitHub API**: `shurcooL/githubv4` (GraphQL v4 API via `golang.org/x/oauth2`)
- **Database**: SQLite via `modernc.org/sqlite` (pure Go, no CGo) + `jmoiron/sqlx`
- **Migrations**: `pressly/goose/v3` (Go-based migrations, no SQL files)
- **Config**: `knadh/koanf` (YAML config parsing)
- **UX**: `schollz/progressbar/v3` (progress bars), `rodaine/table` (table output)
- **VCS**: Git + Jujutsu (`.jj` directory present)

## Project Structure

```
├── cmd/stargazers/main.go    # CLI entrypoint with command definitions
├── api/client.go             # GitHub GraphQL client (stargazers + company fetching)
├── internal/
│   ├── action/               # CLI command handlers
│   │   ├── init.go           # Reads config, seeds repos into DB
│   │   ├── sync.go           # Fetches stargazers from GitHub, upserts into DB
│   │   ├── repo.go           # Lists repos in a table
│   │   └── migrate.go        # Manual DB migration command (hidden)
│   ├── config/config.go      # YAML config reader (koanf)
│   └── db/db.go              # SQLite connection + auto-migration on open
├── migrations/
│   ├── 20250325071003_init_project.go    # Initial schema (user, company, repository, users_to_repositories, sync)
│   └── 20260309083416_add_sync_cursor.go # Adds cursor-based pagination support
├── stargazers.yaml           # Config file listing repos to track
├── go.mod / go.sum
└── .gitignore                # Ignores *.db and db files
```

## Database Schema (SQLite)

### Tables

1. **repository** - Tracked GitHub repos
   - `id`, `owner`, `name`, `created_at`, `updated_at`
   - `last_cursor` (text) - pagination cursor for incremental sync
   - `last_synced_at` (integer) - unix timestamp of last sync
   - Unique constraint on `(owner, name)`

2. **user** - Stargazer profiles
   - `id`, `avatar_url`, `bio`, `email`, `enrichment_data`, `followers_ct`, `following_ct`
   - `fullname`, `is_stargazer`, `is_watcher`, `is_forker`, `login`
   - `company_id` (FK to company), `twitter_url`, `bsky_url`, `linkedin_url`
   - `created_at`, `updated_at`
   - Unique constraint on `login`

3. **company** - Organizations extracted from user profiles
   - `id`, `avatar_url`, `description`, `email`, `name`, `location`, `login`
   - `members_ct`, `repositories_ct`, `website_url`
   - `created_at`, `updated_at`
   - Unique constraint on `login`

4. **users_to_repositories** - Many-to-many join table
   - `id`, `user_id` (FK), `repository_id` (FK), `created_at`
   - Unique index on `(user_id, repository_id)`

5. **sync** - Sync log (defined but not actively used in code)
   - `id`, `synced_type`, `synced_data`, `synced_at`, `created_at`

## CLI Commands

| Command | Description |
|---------|------------|
| `stargazers init -c stargazers.yaml` | Read YAML config and insert repos into DB |
| `stargazers sync -t <GITHUB_TOKEN>` | Fetch stargazers for all repos via GitHub API |
| `stargazers sync --full` | Ignore saved cursors and re-fetch everything |
| `stargazers sync --stargazers-only=false` | Also fetch company/org data |
| `stargazers repo view` | Print tracked repos as a table |
| `stargazers migrate` | Run DB migrations (hidden command) |

Global flag: `--output` / `-o` (default: `db`) - SQLite database file path.

## Data Flow

1. **Init**: User creates `stargazers.yaml` listing repos -> `init` command inserts them into the `repository` table
2. **Sync**: For each repo in DB:
   - Fetches stargazers via GitHub GraphQL API (paginated, 100 per request)
   - Supports incremental sync via saved `last_cursor` (skips already-fetched pages)
   - Upserts each user into `user` table (ON CONFLICT UPDATE)
   - Links users to repos via `users_to_repositories`
   - If `--stargazers-only=false`: also creates company records and fetches org details from GitHub
   - Saves the final pagination cursor + timestamp to `repository.last_cursor`

## Key Data Collected Per Stargazer

- Profile: avatar, bio, email, name, login
- Social: followers count, following count, LinkedIn URL, Bluesky URL
- Company: organization affiliation (resolved to GitHub org when prefixed with `@`)

## Notable Patterns & Design Decisions

- **Auto-migration on DB open**: `db.New()` runs `goose.Up()` every time, so migrations are always applied transparently
- **Incremental sync**: Cursor-based pagination allows fetching only new stargazers since last sync
- **Pure Go SQLite**: Uses `modernc.org/sqlite` (no CGo dependency), making cross-compilation easier
- **Progress bars on stderr**: All progress/status output goes to stderr, keeping stdout clean for data
- **`is_stargazer` stored as unix timestamp**: The field stores `time.Now().Unix()` rather than a boolean - it records *when* the user was seen as a stargazer

## Existing Database Files

- `database-openstatus.db` - Pre-existing database (likely from a previous run for the openstatus repo)
- `db` - Active database file (gitignored)

## Potential Improvements / Observations

- **`repo add` and `repo delete` commands are defined but have no Action handler** - they're stubs
- **`sync` table is created but never written to** - appears to be planned but unused
- **`is_watcher` and `is_forker` columns exist but are never populated** - schema supports watchers/forkers but the API only fetches stargazers
- **`twitter_url` column exists in schema but code writes `linkedin_url` and `bsky_url`** - Twitter/X fetching not implemented
- **`enrichment_data` column exists but is never used** - likely planned for future enrichment integrations
- **Company data fetch only works for `@`-prefixed company fields** - free-text company names are inserted as-is without org lookup
- **No export/query commands** - data is collected into SQLite but there's no built-in way to query or export it (users presumably use external tools like `sqlite3` or datasette)
- **Error handling in sync is mixed** - uses `MustExec` (panics on error) for inserts but graceful error handling for selects
