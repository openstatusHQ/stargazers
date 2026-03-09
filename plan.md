# Implementation Plan

## 1. Populate Bluesky Field

The `bsky_url` column already exists in the `user` table schema. Only code changes needed.

### Files to modify

**`api/client.go`**

- Add `BskyUrl string` field to `User` struct
- In the `GetStargazers` social accounts loop, check for the Bluesky provider (`githubv4.SocialAccountProviderBLUESKY` or equivalent constant) and extract the URL, same pattern as LinkedIn
- Set `BskyUrl` on the `User` struct returned

**`internal/action/sync.go`**

- Add `bsky_url` to both INSERT statements (with company and without company)
- Pass `record.BskyUrl` as the corresponding parameter
- Update the `ON CONFLICT(login) DO UPDATE SET` clause (see incremental sync below) to also update `bsky_url`

### Steps

1. Check available `SocialAccountProvider` constants in `githubv4` for Bluesky (may be `"BLUESKY"` or similar — verify in the library source or docs)
2. Add `BskyUrl` to `User` struct
3. Extract Bluesky URL in the social accounts loop alongside LinkedIn
4. Add `bsky_url` to both SQL insert paths in `sync.go`

---

## 2. Incremental Sync

Currently every `sync` re-fetches all stargazers from page 1. The `sync` table exists but isn't used. The plan uses it to store the last pagination cursor per repo so subsequent syncs only fetch new stargazers.

### How GitHub stargazer pagination works

GitHub's `stargazers` connection returns users in the order they starred. New stargazers are appended at the end. By saving the `endCursor` after a full sync, the next sync can pass it as the `after` parameter and only fetch users added since then.

### Database changes

**New migration** (`migrations/2025XXXX_add_sync_cursor.go`)

Add a `cursor` column to the `repository` table to store the last pagination cursor:

```sql
ALTER TABLE repository ADD COLUMN last_cursor text;
ALTER TABLE repository ADD COLUMN last_synced_at integer;
```

Alternatively, use the existing `sync` table with `synced_type = 'stargazers'` and `synced_data = cursor` per repo. Using `repository` directly is simpler.

**Recommendation**: Add columns to `repository` table — simpler to query, one row per repo.

### Files to modify

**`api/client.go`** — `GetStargazers`

- Add a `startCursor string` parameter (empty string = start from beginning)
- If `startCursor != ""`, set `variables["after"]` to it instead of `nil`
- Return the final `endCursor` alongside the users: change return type to `([]User, string, error)` where the string is the last cursor
- After the pagination loop ends, return `query.Repository.Stargazers.PageInfo.EndCursor`

**`internal/action/sync.go`** — `Sync`

- Query repos with `last_cursor`: `SELECT id, name, owner, last_cursor FROM repository`
- Pass `repo.LastCursor` to `client.GetStargazers(owner, name, lastCursor)`
- After processing stargazers for a repo, update the cursor: `UPDATE repository SET last_cursor = $1, last_synced_at = $2 WHERE id = $3`
- Change `ON CONFLICT(login) DO NOTHING` to `ON CONFLICT(login) DO UPDATE SET ...` so existing users get their profiles refreshed on re-encounter (followers count changes, new bio, etc.)
- Handle the junction table: use `INSERT ... ON CONFLICT DO NOTHING` (need a unique index on `(user_id, repository_id)`)

**New migration file** (generated via goose CLI)

Run from the `migrations/` directory:

```bash
goose create add_sync_cursor go
```

This generates a timestamped file (e.g. `migrations/20260309XXXXXX_add_sync_cursor.go`) with `Up` and `Down` stubs.

Fill in the migration:

- **Up**: `ALTER TABLE repository ADD COLUMN last_cursor text`, `ALTER TABLE repository ADD COLUMN last_synced_at integer`, `CREATE UNIQUE INDEX idx_user_repo ON users_to_repositories(user_id, repository_id)`
- **Down**: Drop the columns and index

### CLI changes

**`cmd/stargazers/main.go`**

- Add `--full` / `-f` boolean flag to the `sync` command (default `false`)
- When `--full` is set, ignore saved cursors and re-fetch everything (reset cursor to empty)

### Data flow (after changes)

```
1. Load repos with last_cursor from DB
2. For each repo:
   a. Call GetStargazers(owner, name, lastCursor)
   b. API starts paginating from cursor (or beginning if empty)
   c. Upsert each user (INSERT ... ON CONFLICT DO UPDATE)
   d. Upsert junction table entry
   e. Save new endCursor + timestamp to repository row
3. Company enrichment (unchanged)
```

### Edge cases

- **First sync**: `last_cursor` is NULL → pass empty string → fetches from beginning (same as today)
- **No new stargazers**: API returns empty page → cursor stays the same, no inserts
- **Unstars**: Incremental sync won't detect users who unstarred. The `--full` flag handles this by re-fetching everything. Could mark stale users in a future enhancement.
- **`--full` flag**: Resets cursor to nil before fetching, then saves the new end cursor

---

## Todo List

### Phase 1: Bluesky Field

- [x] Verify the `githubv4.SocialAccountProvider` constant for Bluesky — no dedicated constant exists in v0.0.0-20240727; using string literal `"BLUESKY"` instead
- [x] Add `BskyUrl string` field with `db:"bsky_url"` tag to `User` struct in `api/client.go`
- [x] In `GetStargazers` social accounts loop, extract Bluesky URL alongside LinkedIn
- [x] In `sync.go`, add `bsky_url` column to the INSERT statement **with** company
- [x] In `sync.go`, add `bsky_url` column to the INSERT statement **without** company
- [x] Pass `record.BskyUrl` as parameter in both INSERT paths
- [x] Build and verify no compile errors: `go build ./...`

### Phase 2: Database Migration

- [x] goose CLI already available at `/opt/homebrew/bin/goose`
- [x] Run `goose create add_sync_cursor go` from the `migrations/` directory → created `20260309083416_add_sync_cursor.go`
- [x] Fill in the **Up** migration with:
  - [x] `ALTER TABLE repository ADD COLUMN last_cursor text`
  - [x] `ALTER TABLE repository ADD COLUMN last_synced_at integer`
  - [x] `CREATE UNIQUE INDEX idx_user_repo ON users_to_repositories(user_id, repository_id)`
- [x] Fill in the **Down** migration to reverse the above

### Phase 3: API Client — Cursor Support

- [x] Add `startCursor string` parameter to `GetStargazers` in `api/client.go`
- [x] When `startCursor != ""`, initialize `variables["after"]` to `githubv4.NewString(...)` instead of `nil`
- [x] Change return type from `([]User, error)` to `([]User, string, error)`
- [x] After pagination loop, capture final `EndCursor` as string
- [x] Return the final cursor as the second return value
- [x] Handle edge case: if no pages fetched (empty repo), returns empty string from zero-value `EndCursor`

### Phase 4: Sync Action — Incremental Logic

- [x] Update repo query struct to include `LastCursor sql.NullString` with `db:"last_cursor"` tag
- [x] Update SELECT to: `SELECT id, name, owner, last_cursor FROM repository`
- [x] Pass `repo.LastCursor.String` to `client.GetStargazers(owner, name, cursor)`
- [x] Handle new third return value (cursor) from `GetStargazers`
- [x] Read `--full` flag from command; if set, pass empty cursor regardless of DB value
- [x] Change both `ON CONFLICT(login) DO NOTHING` to `ON CONFLICT(login) DO UPDATE SET` with all user fields
- [x] Change junction table INSERT to `INSERT ... ON CONFLICT DO NOTHING` (relies on unique index from Phase 2)
- [x] After processing all stargazers for a repo, run `UPDATE repository SET last_cursor = $1, last_synced_at = $2 WHERE id = $3`
- [x] Fix the junction table insert: now queries user ID by login instead of relying on `LastInsertId()`

### Phase 5: CLI — Full Sync Flag

- [x] Add `--full` / `-f` `BoolFlag` to the sync command in `cmd/stargazers/main.go`
- [x] Read the flag value in `Sync` action: `cmd.Bool("full")`
- [x] When `--full` is true, pass empty string as cursor to `GetStargazers`

### Phase 6: Verification

- [x] `go build ./...` — compiles clean
- [x] `go vet ./...` — no issues
- [ ] Run `migrate` to apply new migration on existing DB
- [ ] Run `sync` — should do a full fetch (no saved cursors yet), verify cursor saved in `repository` table
- [ ] Run `sync` again — should fetch 0 new stargazers (no changes), verify it completes quickly
- [ ] Run `sync --full` — should re-fetch all stargazers from scratch
- [ ] Verify `bsky_url` is populated for users who have Bluesky on their GitHub profile
- [ ] Verify existing users get updated fields on re-sync (upsert working)
