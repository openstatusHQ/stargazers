# Stargazers - Project Research

## What It Does

A Go CLI tool that fetches and analyzes stargazers of GitHub repositories. It pulls user data via GitHub's GraphQL API, enriches it with company/organization info, and stores everything in a local SQLite database.

**Core features:**
- Fetch stargazer data (bio, email, followers, social links) from GitHub
- Enrich company data from GitHub organizations
- Track multiple repositories via YAML config
- Persist all data locally in SQLite
- Display repository info in ASCII tables with progress bars during sync

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.23.5 |
| CLI framework | `urfave/cli/v3` (beta) |
| Database | SQLite via `modernc.org/sqlite` + `jmoiron/sqlx` |
| Migrations | `pressly/goose/v3` |
| GitHub API | `shurcooL/githubv4` (GraphQL) |
| Auth | `golang.org/x/oauth2` |
| Config | `knadh/koanf/v2` (YAML) |
| UI | `rodaine/table` (ASCII tables), `schollz/progressbar/v3` |

## Project Structure

```
stargazers/
├── api/
│   └── client.go              # GitHub GraphQL client (GetStargazers, GetCompany)
├── cmd/stargazers/
│   └── main.go                # CLI entry point, command definitions
├── internal/
│   ├── action/
│   │   ├── init.go            # Load repos from YAML into DB
│   │   ├── sync.go            # Fetch stargazers + enrich companies
│   │   ├── repo.go            # Repo list/add/delete commands
│   │   └── migrate.go         # Run DB migrations
│   ├── config/
│   │   └── config.go          # YAML config loader
│   └── db/
│       └── db.go              # SQLite connection setup
├── migrations/
│   └── 20250325071003_init_project.go  # Schema (goose migration)
├── stargazers.yaml            # Config: list of repos to track
├── go.mod / go.sum
└── README.md
```

## Database Schema

5 tables:

- **repository** - Tracked GitHub repos (owner, name)
- **user** - Stargazer profiles (login, avatar, bio, email, followers, following, company_id, linkedin_url, twitter_url, bsky_url)
- **company** - Enriched org data (login, name, description, email, location, members_ct, repositories_ct, website_url)
- **users_to_repositories** - Many-to-many junction (user_id, repository_id)
- **sync** - Sync history log (synced_type, synced_data, synced_at)

Relationships: `user.company_id -> company.id`, junction table links users to repos.

## CLI Commands

| Command | Description |
|---------|-------------|
| `init` | Seed DB with repos from `stargazers.yaml` |
| `sync` | Fetch stargazers from GitHub, optionally enrich companies |
| `repo view` | Display tracked repositories |
| `repo add` | *Not yet implemented* |
| `repo delete` | *Not yet implemented* |
| `migrate` | Run DB migrations (hidden command) |

**Key flags:**
- `--github-token` / `-t` (or `GITHUB_TOKEN` env var) - Required for sync
- `--stargazers-only` / `-s` - Skip company enrichment (default: true)
- `--output` / `-o` - DB file name (default: "db")
- `--config` / `-c` - Config file path (default: "stargazers.yaml")

## Data Flow (Sync)

1. Open SQLite DB, run migrations
2. Load all repos from DB
3. For each repo, paginate through GitHub GraphQL API (100 stargazers/page)
4. Upsert each user into `user` table, link via junction table
5. If `--stargazers-only=false`: fetch org details for each company, upsert into `company` table

Uses `ON CONFLICT DO NOTHING` for deduplication. Progress bars shown during long operations.

## GitHub API Usage

Two GraphQL queries in `api/client.go`:
- **GetStargazers** - Paginated stargazer list with user metadata + social accounts (LinkedIn)
- **GetCompany** - Organization details (members count, repos count, website, etc.)

Auth via OAuth2 personal access token passed as static token source.

## Configuration

`stargazers.yaml` defines repos to track:
```yaml
repositories:
  - owner: openstatusHQ
    name: openstatus
```

Loaded via koanf with YAML parser into `config.Config` struct.

## Current State

- Module: `thibaultleouay.dev/stargazers` v0.0.1
- Active SQLite DB (`database-openstatus.db`, ~1.6MB) exists with real data
- `repo add` and `repo delete` commands are stubbed but not implemented
- Version control uses Jujutsu (jj) alongside git
- `.gitignore` excludes `*.db` files

## Observations

- The `--stargazers-only` flag defaults to `true`, meaning company enrichment is opt-in
- Social account extraction currently only handles LinkedIn from the GitHub GraphQL response; Twitter and Bluesky fields exist in the schema but aren't populated from the API
- The sync command doesn't track incremental progress - it re-fetches all stargazers each run (no cursor persistence between runs)
- The `sync` table exists for logging but its usage pattern suggests it logs each sync operation rather than enabling incremental fetches
- No export functionality yet (CSV, JSON, etc.) - data lives only in SQLite
