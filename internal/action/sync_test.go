package action

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
	"thibaultleouay.dev/stargazers/api"
	_ "thibaultleouay.dev/stargazers/migrations"
)

type stargazerCall struct {
	owner, name, cursor string
}

type mockGitHubClient struct {
	stargazers       []api.User
	endCursor        string
	company          *api.Company
	getStargazersErr error
	getCompanyErr    error
	stargazerCalls   []stargazerCall
	companyCalls     []string
}

func (m *mockGitHubClient) GetStargazers(owner, name, startCursor string) ([]api.User, string, error) {
	m.stargazerCalls = append(m.stargazerCalls, stargazerCall{owner, name, startCursor})
	return m.stargazers, m.endCursor, m.getStargazersErr
}

func (m *mockGitHubClient) GetCompany(login string) (*api.Company, error) {
	m.companyCalls = append(m.companyCalls, login)
	return m.company, m.getCompanyErr
}

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := goose.SetDialect("sqlite"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(db.DB, "."); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedRepo(t *testing.T, db *sqlx.DB, owner, name string) int {
	t.Helper()
	result := db.MustExec("INSERT INTO repository(owner, name) VALUES ($1, $2)", owner, name)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("get last insert id: %v", err)
	}
	return int(id)
}

func TestDoSync_EmptyRepoList(t *testing.T) {
	db := setupTestDB(t)
	mock := &mockGitHubClient{}

	err := doSync(db, mock, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.stargazerCalls) != 0 {
		t.Errorf("expected 0 API calls, got %d", len(mock.stargazerCalls))
	}
}

func TestDoSync_InsertsStargazers(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")

	mock := &mockGitHubClient{
		stargazers: []api.User{
			{Login: "alice", Name: "Alice", Email: "alice@example.com"},
			{Login: "bob", Name: "Bob", Email: "bob@example.com"},
			{Login: "charlie", Name: "Charlie", Email: "charlie@example.com"},
		},
		endCursor: "cursor1",
	}

	err := doSync(db, mock, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	db.Get(&count, "SELECT COUNT(*) FROM user")
	if count != 3 {
		t.Errorf("expected 3 users, got %d", count)
	}
}

func TestDoSync_LinksUsersToRepositories(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")

	mock := &mockGitHubClient{
		stargazers: []api.User{
			{Login: "alice"},
			{Login: "bob"},
		},
		endCursor: "cursor1",
	}

	err := doSync(db, mock, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	db.Get(&count, "SELECT COUNT(*) FROM users_to_repositories")
	if count != 2 {
		t.Errorf("expected 2 links, got %d", count)
	}
}

func TestDoSync_StargazersOnly_SkipsCompanies(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")

	mock := &mockGitHubClient{
		stargazers: []api.User{
			{Login: "alice", Company: "@acme"},
		},
		endCursor: "cursor1",
	}

	err := doSync(db, mock, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var companyCount int
	db.Get(&companyCount, "SELECT COUNT(*) FROM company")
	if companyCount != 0 {
		t.Errorf("expected 0 companies, got %d", companyCount)
	}
	if len(mock.companyCalls) != 0 {
		t.Errorf("expected 0 company API calls, got %d", len(mock.companyCalls))
	}
}

func TestDoSync_WithCompany(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")

	mock := &mockGitHubClient{
		stargazers: []api.User{
			{Login: "alice", Company: "@acme"},
		},
		endCursor: "cursor1",
		company: &api.Company{
			Login:       "acme",
			Name:        "Acme Inc",
			Description: "A company",
		},
	}

	err := doSync(db, mock, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var companyCount int
	db.Get(&companyCount, "SELECT COUNT(*) FROM company")
	if companyCount != 1 {
		t.Errorf("expected 1 company, got %d", companyCount)
	}
	if len(mock.companyCalls) != 1 {
		t.Errorf("expected 1 company API call, got %d", len(mock.companyCalls))
	}
	if len(mock.companyCalls) > 0 && mock.companyCalls[0] != "acme" {
		t.Errorf("expected GetCompany called with 'acme', got %q", mock.companyCalls[0])
	}
}

func TestDoSync_UsesExistingCursor(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")
	db.MustExec("UPDATE repository SET last_cursor = $1 WHERE owner = $2", "saved_cursor", "org")

	mock := &mockGitHubClient{endCursor: "new_cursor"}

	err := doSync(db, mock, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.stargazerCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.stargazerCalls))
	}
	if mock.stargazerCalls[0].cursor != "saved_cursor" {
		t.Errorf("expected cursor 'saved_cursor', got %q", mock.stargazerCalls[0].cursor)
	}
}

func TestDoSync_FullSyncIgnoresCursor(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")
	db.MustExec("UPDATE repository SET last_cursor = $1 WHERE owner = $2", "saved_cursor", "org")

	mock := &mockGitHubClient{endCursor: "new_cursor"}

	err := doSync(db, mock, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.stargazerCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.stargazerCalls))
	}
	if mock.stargazerCalls[0].cursor != "" {
		t.Errorf("expected empty cursor for full sync, got %q", mock.stargazerCalls[0].cursor)
	}
}

func TestDoSync_UpdatesCursorAfterSync(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")

	mock := &mockGitHubClient{
		stargazers: []api.User{{Login: "alice"}},
		endCursor:  "abc123",
	}

	err := doSync(db, mock, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var cursor sql.NullString
	var syncedAt sql.NullInt64
	db.QueryRow("SELECT last_cursor, last_synced_at FROM repository WHERE owner = ?", "org").Scan(&cursor, &syncedAt)
	if !cursor.Valid || cursor.String != "abc123" {
		t.Errorf("expected cursor 'abc123', got %v", cursor)
	}
	if !syncedAt.Valid || syncedAt.Int64 == 0 {
		t.Errorf("expected non-zero last_synced_at, got %v", syncedAt)
	}
}

func TestDoSync_UpsertDuplicateUser(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo1")
	seedRepo(t, db, "org", "repo2")

	mock := &mockGitHubClient{
		stargazers: []api.User{{Login: "alice", Name: "Alice"}},
		endCursor:  "cursor1",
	}

	err := doSync(db, mock, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var userCount int
	db.Get(&userCount, "SELECT COUNT(*) FROM user")
	if userCount != 1 {
		t.Errorf("expected 1 user, got %d", userCount)
	}

	var linkCount int
	db.Get(&linkCount, "SELECT COUNT(*) FROM users_to_repositories")
	if linkCount != 2 {
		t.Errorf("expected 2 links, got %d", linkCount)
	}
}

func TestDoSync_APIError(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "org", "repo")

	mock := &mockGitHubClient{
		getStargazersErr: errors.New("API rate limit exceeded"),
	}

	err := doSync(db, mock, true, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
