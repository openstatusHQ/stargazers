package action

import (
	"testing"
)

func TestDoRepoView_EmptyDB(t *testing.T) {
	db := setupTestDB(t)

	err := doRepoView(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoRepoView_WithRepos(t *testing.T) {
	db := setupTestDB(t)
	seedRepo(t, db, "orgA", "repoA")
	seedRepo(t, db, "orgB", "repoB")

	err := doRepoView(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
