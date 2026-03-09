package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "stargazers.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadConfig_ValidMultipleRepos(t *testing.T) {
	path := writeTempYAML(t, `
repositories:
  - owner: orgA
    name: repoA
  - owner: orgB
    name: repoB
`)
	cfg, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Repositories) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(cfg.Repositories))
	}
	if cfg.Repositories[0].Owner != "orgA" || cfg.Repositories[0].Name != "repoA" {
		t.Errorf("repo 0 mismatch: %+v", cfg.Repositories[0])
	}
	if cfg.Repositories[1].Owner != "orgB" || cfg.Repositories[1].Name != "repoB" {
		t.Errorf("repo 1 mismatch: %+v", cfg.Repositories[1])
	}
}

func TestReadConfig_EmptyFile(t *testing.T) {
	path := writeTempYAML(t, "")
	_, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Note: koanf uses a global instance so previously loaded config persists.
	// An empty file simply doesn't add new data rather than clearing existing state.
}

func TestReadConfig_MissingFile(t *testing.T) {
	_, err := ReadConfig("/nonexistent/path/stargazers.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestReadConfig_MalformedYAML(t *testing.T) {
	path := writeTempYAML(t, `
repositories:
  - owner: orgA
    name
  invalid yaml [[[
`)
	_, err := ReadConfig(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestReadConfig_ExtraFields(t *testing.T) {
	path := writeTempYAML(t, `
extra_field: hello
repositories:
  - owner: orgA
    name: repoA
    unknown: value
`)
	cfg, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Repositories) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repositories))
	}
	if cfg.Repositories[0].Owner != "orgA" || cfg.Repositories[0].Name != "repoA" {
		t.Errorf("repo mismatch: %+v", cfg.Repositories[0])
	}
}
