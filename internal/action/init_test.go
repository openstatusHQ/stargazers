package action

import (
	"testing"

	"thibaultleouay.dev/stargazers/internal/config"
)

func makeConfig(repos ...struct{ Owner, Name string }) *config.Config {
	cfg := &config.Config{}
	for _, r := range repos {
		cfg.Repositories = append(cfg.Repositories, struct {
			Name  string `koanf:"name"`
			Owner string `koanf:"owner"`
		}{Name: r.Name, Owner: r.Owner})
	}
	return cfg
}

func TestDoInit_InsertsRepos(t *testing.T) {
	db := setupTestDB(t)

	cfg := makeConfig(
		struct{ Owner, Name string }{"orgA", "repoA"},
		struct{ Owner, Name string }{"orgB", "repoB"},
	)

	err := doInit(db, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	db.Get(&count, "SELECT COUNT(*) FROM repository")
	if count != 2 {
		t.Errorf("expected 2 repos, got %d", count)
	}
}

func TestDoInit_EmptyConfig(t *testing.T) {
	db := setupTestDB(t)

	cfg := makeConfig()

	err := doInit(db, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int
	db.Get(&count, "SELECT COUNT(*) FROM repository")
	if count != 0 {
		t.Errorf("expected 0 repos, got %d", count)
	}
}

func TestDoInit_DuplicateRepos(t *testing.T) {
	db := setupTestDB(t)

	cfg := makeConfig(struct{ Owner, Name string }{"orgA", "repoA"})

	err := doInit(db, cfg)
	if err != nil {
		t.Fatalf("first init: %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for duplicate repos, got none")
		}
	}()
	doInit(db, cfg)
}
