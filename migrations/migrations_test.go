package migrations

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

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

func TestSchema_RepositoryRoundTrip(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec("INSERT INTO repository(owner, name) VALUES (?, ?)", "orgA", "repoA")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var owner, name string
	err = db.QueryRow("SELECT owner, name FROM repository WHERE owner = ?", "orgA").Scan(&owner, &name)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if owner != "orgA" || name != "repoA" {
		t.Errorf("got owner=%q name=%q", owner, name)
	}
}

func TestSchema_UserRoundTrip(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec(`INSERT INTO user(login, fullname, email, bio, avatar_url, followers_ct, following_ct, is_stargazer, linkedin_url, bsky_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"jdoe", "John Doe", "john@example.com", "dev", "https://avatar.url", 100, 50, 1, "https://linkedin.com/in/jdoe", "https://bsky.app/profile/jdoe")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var login, fullname, email string
	err = db.QueryRow("SELECT login, fullname, email FROM user WHERE login = ?", "jdoe").Scan(&login, &fullname, &email)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if login != "jdoe" || fullname != "John Doe" || email != "john@example.com" {
		t.Errorf("got login=%q fullname=%q email=%q", login, fullname, email)
	}
}

func TestSchema_CompanyRoundTrip(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec(`INSERT INTO company(login, name, email, location, website_url, members_ct, repositories_ct)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"@acme", "Acme Inc", "info@acme.com", "SF", "https://acme.com", 50, 10)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var login, name string
	err = db.QueryRow("SELECT login, name FROM company WHERE login = ?", "@acme").Scan(&login, &name)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if login != "@acme" || name != "Acme Inc" {
		t.Errorf("got login=%q name=%q", login, name)
	}
}

func TestSchema_UserUniqueLogin(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec("INSERT INTO user(login) VALUES (?)", "jdoe")
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	_, err = db.Exec("INSERT INTO user(login) VALUES (?)", "jdoe")
	if err == nil {
		t.Fatal("expected unique constraint error, got nil")
	}
}

func TestSchema_RepositoryUniqueOwnerName(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec("INSERT INTO repository(owner, name) VALUES (?, ?)", "orgA", "repoA")
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	_, err = db.Exec("INSERT INTO repository(owner, name) VALUES (?, ?)", "orgA", "repoA")
	if err == nil {
		t.Fatal("expected unique constraint error, got nil")
	}
}

func TestSchema_UsersToRepositoriesUniqueIndex(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec("INSERT INTO repository(owner, name) VALUES (?, ?)", "orgA", "repoA")
	if err != nil {
		t.Fatalf("insert repo: %v", err)
	}
	_, err = db.Exec("INSERT INTO user(login) VALUES (?)", "jdoe")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var userID, repoID int64
	db.QueryRow("SELECT id FROM user WHERE login = ?", "jdoe").Scan(&userID)
	db.QueryRow("SELECT id FROM repository WHERE owner = ?", "orgA").Scan(&repoID)

	_, err = db.Exec("INSERT INTO users_to_repositories(user_id, repository_id) VALUES (?, ?)", userID, repoID)
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	_, err = db.Exec("INSERT INTO users_to_repositories(user_id, repository_id) VALUES (?, ?)", userID, repoID)
	if err == nil {
		t.Fatal("expected unique index error, got nil")
	}
}

func TestSchema_UserCompanyForeignKey(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec("INSERT INTO company(login) VALUES (?)", "@acme")
	if err != nil {
		t.Fatalf("insert company: %v", err)
	}

	var companyID int64
	db.QueryRow("SELECT id FROM company WHERE login = ?", "@acme").Scan(&companyID)

	_, err = db.Exec("INSERT INTO user(login, company_id) VALUES (?, ?)", "jdoe", companyID)
	if err != nil {
		t.Fatalf("insert user with company_id: %v", err)
	}

	var gotCompanyID sql.NullInt64
	db.QueryRow("SELECT company_id FROM user WHERE login = ?", "jdoe").Scan(&gotCompanyID)
	if !gotCompanyID.Valid || gotCompanyID.Int64 != companyID {
		t.Errorf("expected company_id=%d, got %v", companyID, gotCompanyID)
	}
}

func TestSchema_DefaultCreatedAt(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.Exec("INSERT INTO repository(owner, name) VALUES (?, ?)", "orgA", "repoA")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var createdAt sql.NullInt64
	db.QueryRow("SELECT created_at FROM repository WHERE owner = ?", "orgA").Scan(&createdAt)
	if !createdAt.Valid || createdAt.Int64 == 0 {
		t.Errorf("expected non-zero created_at, got %v", createdAt)
	}
}
